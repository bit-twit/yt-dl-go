package downloader

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloader_GetVideoInfo(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	ctx := context.Background()
	dl := NewDownloader("test_data")
	// youtube-dl test video
	video, err := dl.GetVideoInfo(ctx, "BaW_jenozKc")
	require.NoError(err)
	require.NotNil(video)

	fmt.Printf("video: %+v\n", video)
	assert.Equal(`youtube-dl test video "'/\ä↭𝕐`, video.Title)
	assert.Equal(`Philipp Hagemeister`, video.Author)
	assert.Equal(10*time.Second, video.Duration)
	assert.Len(video.Formats, 18)

	assert.Greater(len(video.Formats), 0)
}

func TestDownload_MP4Stream(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	dl := NewDownloader("test_data")
	// youtube-dl test video
	video, err := dl.GetVideoInfo(ctx, "QcHvzNBtlOw")
	assert.NoError(err)
	assert.NotNil(video)

	assert.Equal("Metallica - Frantic (Official Music Video)", video.Title)
	assert.Equal("Warner Records Vault", video.Author)
	assert.Equal(4*time.Minute+58*time.Second, video.Duration)
	assert.Len(video.Formats, 13)

	if assert.Greater(len(video.Formats), 0) {
		assert.NoError(dl.Download(video, video.Formats.FindByMimeType("video/mp4"), "test.mp4"))
	}
}
