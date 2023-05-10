![](static/boominator-pleasure.png)

Fetches Bandcamp albums featured in Tolpa Bumov on Radio Študent and arranges them in a quasi playlist.


## Run

Download the repository and run for boom-fetch (linux) or boom-fetch.exe (wins). Go to localhost:8090/tolpe?from=2021-10-20.

The from parameter is mandatory. The required date format is YYYY-MM-DD.

Other query string parameters are to (YYYY-MM-DD) and update (default false). If update is sent to true, the list of albums is refreshed ie. fetched from radiostudent.si anew.


## Feature ideas

- add to your bandcamp wishlist button
- add db to save links
- some sort of loading feedback
