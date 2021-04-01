package api

import (
	"github.com/bit-twit/yt-dl-go/types"
	"google.golang.org/api/youtube/v3"
)

type YoutubeAPI struct {
	config  types.Config
	service *youtube.Service
}

func NewYoutubeAPI(c types.Config) *YoutubeAPI {
	oauthClient, err := GetOAuth2HTTPClient(c.ClientSecretPath)
	if err != nil {
		panic(err)
	}
	httpURLPath := "http"
	if c.HTTPSEnabled {
		httpURLPath += "s"
	}
	if googleAPI, err := youtube.New(oauthClient); err == nil {
		youtubeAPI := &YoutubeAPI{
			config:  c,
			service: googleAPI,
		}
		return youtubeAPI
	} else {
		panic(err)
	}
}

func (y *YoutubeAPI) ListPlaylists() ([]string, error) {
	req := y.service.Playlists.List([]string{"id, snippet,status"})
	resp, err := req.Mine(true).Do()
	results := make([]string, len(resp.Items), len(resp.Items))
	for i, _ := range resp.Items {
		results[i] = resp.Items[i].Id
	}
	return results, err
}

func (y *YoutubeAPI) ListVideosForPlaylist(playlistId string, maxResults int64) ([]types.Video, error) {
	req := y.service.PlaylistItems.
		List([]string{"id,snippet,status"}).
		PlaylistId(playlistId).
		MaxResults(maxResults)
	resp, err := req.Do()
	results := make([]types.Video, 0, len(resp.Items))

	responseProcessor := func(playlistItemResponse *youtube.PlaylistItemListResponse) {
		for _, el := range playlistItemResponse.Items {
			if !(len(results) < int(maxResults)) {
				break
			}
			if el.Snippet.ResourceId.Kind == "youtube#video" {
				results = append(results, types.Video{
					ID:          el.Snippet.ResourceId.VideoId,
					Title:       el.Snippet.Title,
					Description: el.Snippet.ResourceId.Kind,
				})
			}
		}
	}
	responseProcessor(resp)

	if resp.NextPageToken != "" {
		for resp.NextPageToken != "" && len(results) < int(maxResults) {
			req = req.PageToken(resp.NextPageToken)
			resp, err = req.Do()
			if err != nil {
				break
			}
			responseProcessor(resp)
		}
	}

	return results, err
}

func (y *YoutubeAPI) ListVideosLiked(maxResults int64) ([]types.Video, error) {
	req := y.service.Videos.
		List([]string{"snippet,contentDetails,statistics"}).MyRating("like").MaxResults(maxResults)
	resp, err := req.Do()
	if err != nil {
		return nil, err
	}
	results := make([]types.Video, 0, len(resp.Items))

	responseProcessor := func(playlistItemResponse *youtube.VideoListResponse) {
		for _, el := range playlistItemResponse.Items {
			if el.Kind == "youtube#video" {
				results = append(results, types.Video{
					ID:          el.Id,
					Title:       el.Snippet.Title,
					Description: el.Kind,
				})
			}
		}
	}
	responseProcessor(resp)

	if resp.NextPageToken != "" {
		for resp.NextPageToken != "" && len(results) < int(maxResults) {
			req = req.PageToken(resp.NextPageToken)
			resp, err = req.Do()
			if err != nil {
				break
			}
			responseProcessor(resp)
		}
	}

	return results, err
}
