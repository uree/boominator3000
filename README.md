![](static/boominator-pleasure.png)

Fetches Bandcamp albums featured in Tolpa Bumov on Radio Študent and arranges them in a quasi playlist.


## Run

Download the repository and run for boom-fetch (linux) or boom-fetch.exe (wins).

Go to `http://localhost:8090/tolpe?from=2021-10-20`.

The `from` parameter is mandatory. The required date format is `YYYY-MM-DD`.

Other query string parameters are `to (YYYY-MM-DD)` and `update (default false)`. If update is set to true,

`/tolpe?from=2021-10-20&update=true`

the list of albums is refreshed ie. fetched from radiostudent.si anew. By default `update=true` fetches the albums which are newer than the ones already in the db.

To fetch older albums, use `/full-update`.

You can add an album to your favourites by clicking on the star. Inspect your favourites at

`/fav`


## Development

Run

`go run .`
or
`./boom-fetch`

Make sure boominator or boom-fetch are not already running on the system otherwise you'll get no stdout.

```
ps aux | grep boominator
ps aux | boom-fetch
```

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
- improve `/full-update` logic so that it ignores the savedLastPage value and updates/overwrites all records
