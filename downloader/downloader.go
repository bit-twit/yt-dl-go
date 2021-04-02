package downloader

import (
	"context"
	"errors"
	"fmt"
	"github.com/bit-twit/yt-dl-go/types"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

const (
	default_api_timeout  = 60
	max_http_connections = 5
)

// Downloader offers high level functions to download videos into files
type Downloader struct {
	// HTTPClient can be used to set a custom HTTP client.
	// If not set, http.DefaultClient will be used
	HTTPClient *http.Client

	// decipherOpsCache cache decipher operations
	decipherOpsCache DecipherOperationsCache
	OutputDir        string // optional directory to store the files
}

func NewDownloader(outputDir string) *Downloader {
	tr := &http.Transport{
		MaxConnsPerHost:       max_http_connections,
		MaxIdleConns:          max_http_connections,
		IdleConnTimeout:       default_api_timeout * time.Second,
		ResponseHeaderTimeout: default_api_timeout * time.Second,
		TLSHandshakeTimeout:   default_api_timeout * time.Second,
	}
	client := &http.Client{Transport: tr}
	return &Downloader{
		HTTPClient:       client,
		decipherOpsCache: &SimpleCache{},
		OutputDir:        outputDir,
	}

}

func (dl *Downloader) getOutputFile(v *types.Video, format *types.Format, outputFile string) (string, error) {
	if outputFile == "" {
		outputFile = SanitizeFilename(v.Title)
		outputFile += pickIdealFileExtension(format.MimeType)
	}

	if dl.OutputDir != "" {
		if err := os.MkdirAll(dl.OutputDir, 0o755); err != nil {
			return "", err
		}
		outputFile = filepath.Join(dl.OutputDir, outputFile)
	}

	return outputFile, nil
}

// Download : Starting download video by arguments.
func (dl *Downloader) Download(v *types.Video, format *types.Format, outputFile string) error {
	destFile, err := dl.getOutputFile(v, format, outputFile)
	if err != nil {
		return err
	}

	// Create output file
	out, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer out.Close()

	fmt.Printf("Download to file=%s", destFile)

	resp, err := dl.getStream(v, format)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// GetVideoInfo fetches video metadata with a context
func (dl *Downloader) GetVideoInfo(ctx context.Context, id string) (*types.Video, error) {

	// Circumvent age restriction to pretend access through googleapis.com
	eurl := "https://youtube.googleapis.com/v/" + id
	finalUrl := "https://youtube.com/get_video_info?video_id=" + id + "&eurl=" + eurl
	fmt.Printf("GET video info : %s", finalUrl)
	body, err := dl.httpGetBodyBytes(ctx, finalUrl)
	if err != nil {
		return nil, err
	}

	v := &types.Video{
		ID: id,
	}

	err = v.ParseVideoInfo(body)

	// If the uploader has disabled embedding the video on other sites, parse video page
	if err == types.ErrNotPlayableInEmbed {
		videoPageUrl := "https://www.youtube.com/watch?v=" + id
		fmt.Printf("EMBED DISABLED, GET video page: %s", videoPageUrl)
		html, err := dl.httpGetBodyBytes(ctx, videoPageUrl)
		if err != nil {
			return nil, err
		}

		return v, v.ParseVideoPage(html)
	}

	return v, err
}

// GetStream returns the HTTP response for a specific format
func (dl *Downloader) getStream(video *types.Video, format *types.Format) (*http.Response, error) {
	url, err := dl.getStreamURL(video, format)
	if err != nil {
		return nil, err
	}

	return dl.httpGet(context.Background(), url)
}

// GetStreamURL returns the url for a specific format
func (dl *Downloader) getStreamURL(video *types.Video, format *types.Format) (string, error) {
	if format.URL != "" {
		return format.URL, nil
	}

	cipher := format.Cipher
	if cipher == "" {
		return "", types.ErrCipherNotFound
	}

	return dl.decipherURL(video.ID, cipher)
}

func (dl *Downloader) decipherURL(videoID string, cipher string) (string, error) {
	queryParams, err := url.ParseQuery(cipher)
	if err != nil {
		return "", err
	}

	/* eg:
	    extract decipher from  https://youtube.com/s/player/4fbb4d5b/player_ias.vflset/en_US/base.js

	    var Mt={
		splice:function(a,b){a.splice(0,b)},
		reverse:function(a){a.reverse()},
		EQ:function(a,b){var c=a[0];a[0]=a[b%a.length];a[b%a.length]=c}};

		a=a.split("");
		Mt.splice(a,3);
		Mt.EQ(a,39);
		Mt.splice(a,2);
		Mt.EQ(a,1);
		Mt.splice(a,1);
		Mt.EQ(a,35);
		Mt.EQ(a,51);
		Mt.splice(a,2);
		Mt.reverse(a,52);
		return a.join("")
	*/

	operations, err := dl.parseDecipherOpsWithCache(videoID)
	if err != nil {
		return "", err
	}

	// apply operations
	bs := []byte(queryParams.Get("s"))
	for _, op := range operations {
		bs = op(bs)
	}

	decipheredURL := fmt.Sprintf("%s&%s=%s", queryParams.Get("url"), queryParams.Get("sp"), string(bs))
	return decipheredURL, nil
}

func (dl *Downloader) parseDecipherOpsWithCache(videoID string) (operations []DecipherOperation, err error) {
	if dl.decipherOpsCache == nil {
		dl.decipherOpsCache = NewSimpleCache()
	}

	if ops := dl.decipherOpsCache.Get(videoID); ops != nil {
		return ops, nil
	}

	ops, err := dl.parseDecipherOps(videoID)
	if err != nil {
		return nil, err
	}

	dl.decipherOpsCache.Set(videoID, ops)
	return ops, err
}

const (
	jsvarStr   = "[a-zA-Z_\\$][a-zA-Z_0-9]*"
	reverseStr = ":function\\(a\\)\\{" +
		"(?:return )?a\\.reverse\\(\\)" +
		"\\}"
	spliceStr = ":function\\(a,b\\)\\{" +
		"a\\.splice\\(0,b\\)" +
		"\\}"
	swapStr = ":function\\(a,b\\)\\{" +
		"var c=a\\[0\\];a\\[0\\]=a\\[b(?:%a\\.length)?\\];a\\[b(?:%a\\.length)?\\]=c(?:;return a)?" +
		"\\}"
)

var (
	basejsPattern = regexp.MustCompile(`(/s/player/\w+/player_ias.vflset/\w+/base.js)`)

	actionsObjRegexp = regexp.MustCompile(fmt.Sprintf(
		"var (%s)=\\{((?:(?:%s%s|%s%s|%s%s),?\\n?)+)\\};", jsvarStr, jsvarStr, swapStr, jsvarStr, spliceStr, jsvarStr, reverseStr))

	actionsFuncRegexp = regexp.MustCompile(fmt.Sprintf(
		"function(?: %s)?\\(a\\)\\{"+
			"a=a\\.split\\(\"\"\\);\\s*"+
			"((?:(?:a=)?%s\\.%s\\(a,\\d+\\);)+)"+
			"return a\\.join\\(\"\"\\)"+
			"\\}", jsvarStr, jsvarStr, jsvarStr))

	reverseRegexp = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, reverseStr))
	spliceRegexp  = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, spliceStr))
	swapRegexp    = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, swapStr))
)

func (dl *Downloader) parseDecipherOps(videoID string) (operations []DecipherOperation, err error) {
	embedURL := fmt.Sprintf("https://youtube.com/embed/%s?hl=en", videoID)
	embedBody, err := dl.httpGetBodyBytes(context.Background(), embedURL)
	if err != nil {
		return nil, err
	}

	// example: /s/player/f676c671/player_ias.vflset/en_US/base.js
	escapedBasejsURL := string(basejsPattern.Find(embedBody))
	if escapedBasejsURL == "" {
		log.Println("playerConfig:", string(embedBody))
		return nil, errors.New("unable to find basejs URL in playerConfig")
	}

	basejsBody, err := dl.httpGetBodyBytes(context.Background(), "https://youtube.com"+escapedBasejsURL)
	if err != nil {
		return nil, err
	}

	objResult := actionsObjRegexp.FindSubmatch(basejsBody)
	funcResult := actionsFuncRegexp.FindSubmatch(basejsBody)
	if len(objResult) < 3 || len(funcResult) < 2 {
		return nil, fmt.Errorf("error parsing signature tokens (#obj=%d, #func=%d)", len(objResult), len(funcResult))
	}

	obj := objResult[1]
	objBody := objResult[2]
	funcBody := funcResult[1]

	var reverseKey, spliceKey, swapKey string

	if result := reverseRegexp.FindSubmatch(objBody); len(result) > 1 {
		reverseKey = string(result[1])
	}
	if result := spliceRegexp.FindSubmatch(objBody); len(result) > 1 {
		spliceKey = string(result[1])
	}
	if result := swapRegexp.FindSubmatch(objBody); len(result) > 1 {
		swapKey = string(result[1])
	}

	regex, err := regexp.Compile(fmt.Sprintf("(?:a=)?%s\\.(%s|%s|%s)\\(a,(\\d+)\\)", obj, reverseKey, spliceKey, swapKey))
	if err != nil {
		return nil, err
	}

	var ops []DecipherOperation
	for _, s := range regex.FindAllSubmatch(funcBody, -1) {
		switch string(s[1]) {
		case reverseKey:
			ops = append(ops, reverseFunc)
		case swapKey:
			arg, _ := strconv.Atoi(string(s[2]))
			ops = append(ops, newSwapFunc(arg))
		case spliceKey:
			arg, _ := strconv.Atoi(string(s[2]))
			ops = append(ops, newSpliceFunc(arg))
		}
	}
	return ops, nil
}

// httpGet does a HTTP GET request, checks the response to be a 200 OK and returns it
func (dl *Downloader) httpGet(ctx context.Context, url string) (resp *http.Response, err error) {
	fmt.Printf("GET %s", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err = dl.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, &types.HttpError{Status: strconv.Itoa(resp.StatusCode), Reason: "Unexpected status code"}
	}

	return
}

// httpGetBodyBytes reads the whole HTTP body and returns it
func (dl *Downloader) httpGetBodyBytes(ctx context.Context, url string) ([]byte, error) {
	resp, err := dl.httpGet(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
