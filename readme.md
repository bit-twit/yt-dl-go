# Run

To run you need a ./resources/client_secret.json file with Youtube Data API v3 OAUTH2 API credentials file.
You can set secret and api key through env vars.
!!! Also used in tests.

Application will ask user to get Oauth token from browser using a generated url.
Subsequent calls will reuses saved auth details.

# Youtube information

Uses GO SDK to parse playlist and video information

# Youtube video download

Uses downloader code from https://github.com/kkdai/youtube
