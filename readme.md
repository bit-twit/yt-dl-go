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
