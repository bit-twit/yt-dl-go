package api

import (
	"fmt"
	"github.com/bit-twit/yt-dl-go/types"
	"github.com/bit-twit/yt-dl-go/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	defaultClientSecret = "./resources/client_secret.json"
	defaultPlaylistId   = "PLJhcxmK4B_Nv-tf8OwtuFCr-Xx_1Z3EaR"
)

func TestListPlaylists(t *testing.T) {
	yt := NewYoutubeAPI(types.Config{
		true,
		"youtube.googleapis.com",
		443,
		utils.GetEnv("YOUTUBE_API_KEY", ""),
		defaultClientSecret,
	})

	res, err := yt.ListPlaylists()
	assert.NoError(t, err)
	assert.True(t, len(res) > 0)
}

func TestListVideosForPlaylist(t *testing.T) {
	yt := NewYoutubeAPI(types.Config{
		true,
		"youtube.googleapis.com",
		443,
		utils.GetEnv("YOUTUBE_API_KEY", ""),
		defaultClientSecret,
	})

	vids, err := yt.ListVideosForPlaylist(defaultPlaylistId, 5)
	assert.NoError(t, err)
	fmt.Printf("Found %d videos\n", len(vids))
	assert.True(t, len(vids) > 0)
}

func TestListVideosLiked(t *testing.T) {
	yt := NewYoutubeAPI(types.Config{
		true,
		"youtube.googleapis.com",
		443,
		utils.GetEnv("YOUTUBE_API_KEY", ""),
		defaultClientSecret,
	})

	vids, err := yt.ListVideosLiked(5)
	assert.NoError(t, err)
	fmt.Printf("Found %d videos\n", len(vids))
	assert.True(t, len(vids) > 0)
}
