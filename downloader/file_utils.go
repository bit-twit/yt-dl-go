package downloader

import (
	"regexp"
	"strings"
)

// Rely on hardcoded canonical mime types, as the ones provided by Go aren't exhaustive [1].
// This seems to be a recurring problem for youtube downloaders, see [2].
// The implementation is based on mozilla's list [3], IANA [4] and Youtube's support [5].
// [1] https://github.com/golang/go/blob/ed7888aea6021e25b0ea58bcad3f26da2b139432/src/mime/type.go#L60
// [2] https://github.com/ZiTAL/youtube-dl/blob/master/mime.types
// [3] https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
// [4] https://www.iana.org/assignments/media-types/media-types.xhtml#video
// [5] https://support.google.com/youtube/troubleshooter/2888402?hl=en
var canonicals = map[string]string{
	"audio/mp4":        ".mp4a",
	"video/quicktime":  ".mov",
	"video/x-msvideo":  ".avi",
	"video/x-matroska": ".mkv",
	"video/mpeg":       ".mpeg",
	"video/webm":       ".webm",
	"video/3gpp2":      ".3g2",
	"video/x-flv":      ".flv",
	"video/3gpp":       ".3gp",
	"video/mp4":        ".mp4",
	"video/ogg":        ".ogv",
	"video/mp2t":       ".ts",
}

func pickIdealFileExtension(mediaType string) string {
	for k, v := range canonicals {
		if strings.Contains(mediaType, k) {
			return v
		}
	}
	return ""
}

func SanitizeFilename(fileName string) string {
	// Characters not allowed on mac
	//	:/
	// Characters not allowed on linux
	//	/
	// Characters not allowed on windows
	//	<>:"/\|?*

	// Ref https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file#naming-conventions

	fileName = regexp.MustCompile(`[:/<>\:"\\|?*]`).ReplaceAllString(fileName, "")
	fileName = regexp.MustCompile(`\s+`).ReplaceAllString(fileName, " ")

	return fileName
}
