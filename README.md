![](static/boominator-pleasure.png)

Fetches Bandcamp albums featured in Tolpa Bumov on Radio Študent and arranges them in a quasi playlist.


## Run

Download the repository and run for boom-fetch (linux) or boom-fetch.exe (wins).

Go to http://localhost:8090/tolpe?from=2021-10-20.

The `from` parameter is mandatory. The required date format is `YYYY-MM-DD`.

Other query string parameters are `to (YYYY-MM-DD)` and `update (default false)`. If update is set to true,

http://localhost:8090/tolpe?from=2021-10-20&update=true

the list of albums is refreshed ie. fetched from radiostudent.si anew.

You can add an album to your favourites by clicking on the star. Inspect your favourites at

http://localhost:8090/fav

## Development

Run

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

## TO DO
- test on wins
- better install instructions (how to add to dash on linux)
