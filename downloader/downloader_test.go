package downloader

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownload_FirstStream(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	ctx := context.Background()
	dl := &Downloader{
		OutputDir: "test_data",
	}

	// youtube-dl test video
	video, err := dl.GetVideoInfo(ctx, "BaW_jenozKc")
	require.NoError(err)
	require.NotNil(video)

	assert.Equal(`youtube-dl test video "'/\ä↭𝕐`, video.Title)
	assert.Equal(`Philipp Hagemeister`, video.Author)
	assert.Equal(10*time.Second, video.Duration)
	assert.Len(video.Formats, 18)

	if assert.Greater(len(video.Formats), 0) {
		assert.NoError(dl.Download(ctx, video, &video.Formats[0], ""))
	}
}
