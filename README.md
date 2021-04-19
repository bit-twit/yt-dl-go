WIP - Youtube downloader in GO
===============================

Simple cli to fetch video information from a youtube account (playlists, liked, history). 

Uses Google Youtube Go SDK.

# TODO

+ integrate with GO SDK to pull video information for user
+ add downloading capabilities for a video ID
- download videos to a specific folder (right now it just outputs video ids from Google Youtube API - downloading needs to be called separately from Downloader)
- add Youtube history support
- advanced cli options
- run as server

# Run

To run you need a ./resources/client_secret.json file with Youtube Data API v3 OAUTH2 API credentials file.
You can set secret and api key through env vars.
!!! Also used in tests.

Example run :

```bash
go run main.go -apiKey ... -clientSecret ./resources/client_secret.json
```

Example tests:
```bash
go test -v  -cover ./...
```

Application will ask user to get Oauth token from browser using a generated url - get token from browser redirect and input back to program console.
Subsequent calls will reuses saved auth details.

# Youtube information

Uses GO SDK to parse playlist and video information

# Youtube video download

Uses downloader code from https://github.com/kkdai/youtube
