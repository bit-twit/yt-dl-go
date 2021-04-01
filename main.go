package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/bit-twit/yt-dl-go/api"
	"github.com/bit-twit/yt-dl-go/types"
	"github.com/bit-twit/yt-dl-go/utils"
	"os"
)

var (
	apiKey       = flag.String("apiKey", "", "The API key from Google developer console.")
	clientSecret = flag.String("clientSecret", "./resources/client_secret.json", "The OAuth web api client secret file from Google developer console.")
	maxResults   = flag.Int64("maxResults", 5, "The maximum number of video resources to fetch from each playlist.")
	playlistId   = flag.String("playlistId", "", "Retrieve information about specific playlist - otherwise it will retrieve all users's playlist.")
	liked        = flag.Bool("liked", true, "Retrieve videos from special liked playlist.")
)

func main() {
	flag.Parse()

	finalApiKey := utils.GetEnv("YOUTUBE_API_KEY", *apiKey)
	if finalApiKey == "" {
		panic(errors.New("Expected YOUTUBE_API_KEY env or -apiKey param!"))
	}

	finalClientSecret := utils.GetEnv("YOUTUBE_CLIENT_SECRET", *clientSecret)
	if _, err := os.Stat(finalClientSecret); err != nil {
		panic(err)
	}

	yt := api.NewYoutubeAPI(types.Config{
		true,
		"youtube.googleapis.com",
		443,
		finalApiKey,
		finalClientSecret,
	})

	var ps []string
	var err error
	if *playlistId != "" {
		ps = []string{*playlistId}
	} else {
		// fetch all mine
		ps, err = yt.ListPlaylists()
		if err != nil {
			panic(err)
		}
		fmt.Print("Playlist ids : ")
		fmt.Printf("%+v\n", ps)
	}

	// fetch video information from playlists
	vs := make([]types.Video, 0, 5000)
	for _, p := range ps {
		if pVideos, err := yt.ListVideosForPlaylist(p, *maxResults); err == nil {
			vs = append(vs, pVideos...)
		}
	}

	// fetch liked videos
	if *liked {
		if lVideos, err := yt.ListVideosLiked(*maxResults); err == nil {
			vs = append(vs, lVideos...)
		}
	}

	fmt.Printf("Found %d videos: \n", len(vs))
	for _, v := range vs {
		fmt.Println(v.ID)
	}
}
