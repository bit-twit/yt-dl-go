package downloader

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_changeExtension(t *testing.T) {
	test1 := "tttt/fdfd/file.mp4a"
	res1 := changeExtension(test1, "mp3")
	assert.Equal(t, "tttt/fdfd/file.mp3", res1)

	test2 := "Metallica/Frantic (Official Music Video).mp4a"
	res2 := changeExtension(test2, "mp3")
	assert.Equal(t, "Metallica/Frantic (Official Music Video).mp3", res2)
}

func TestConvertMP4aToMP3(t *testing.T) {
	assert := assert.New(t)
	test := "./test_data/Metallica/Frantic (Official Music Video).mp4a"
	resFile, err := ConvertMP4aToMP3(test)
	if err != nil {
		fmt.Println(err)
	}
	assert.NoError(err)
	assert.Equal("./test_data/Metallica/Frantic (Official Music Video).mp3", resFile)
	assert.FileExists("./test_data/Metallica/Frantic (Official Music Video).mp3")
}
