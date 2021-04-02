package downloader

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ConvertMP4aToMP3(fileName string) (string, error) {
	destFile := changeExtension(fileName, "mp3")

	ffmpegVersionCmd := exec.Command("ffmpeg",
		"-y",
		"-loglevel", "warning",
		"-i", fileName,
		"-acodec", "libmp3lame",
		"-ab", "128k", // medium quality
		destFile,
	)
	ffmpegVersionCmd.Stderr = os.Stderr
	ffmpegVersionCmd.Stdout = os.Stdout

	convErr := ffmpegVersionCmd.Run()
	if convErr != nil {
		return "", convErr
	}

	return destFile, nil
}

func changeExtension(fileName string, newExt string) string {
	ext := filepath.Ext(fileName)
	if ext == "" {
		return fileName + "." + newExt
	}
	return strings.Replace(fileName, ext, "."+newExt, 1)
}
