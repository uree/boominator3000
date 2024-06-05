![](static/boominator-pleasure.png)

Fetches Bandcamp albums featured in Tolpa Bumov on Radio Študent and arranges them in a quasi playlist.


## Run

Download the repository and run for boom-fetch (linux) or boom-fetch.exe (wins). Go to localhost:8090/tolpe?from=2021-10-20.

The from parameter is mandatory. The required date format is YYYY-MM-DD.

Other query string parameters are to (YYYY-MM-DD) and update (default false). If update is sent to true, the list of albums is refreshed ie. fetched from radiostudent.si anew.

## Development

`go run .`
or
`./boom-fetch`

Build

`go build`
&
`env GOOS=windows GOARCH=amd64 go build`

## Feature ideas
- some sort of loading feedback
- export list of favourites
- download favourites via soulseek
