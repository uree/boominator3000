# BOOMINATOR 3000

Fetches Bandcamp albums featured in Tolpa Bumov on Radio Študent and arranges them in a quasi playlist.


## Run

Run boom-fetch (linux) or boom-fetch.exe (wins) and go to localhost:8090/tolpe?from=2021-10-20.

The from parameter is mandatory. Date format is YYYY-MM-DD.

Other parameters are to (YYYY-MM-DD) and update (default false). If update is sent to true, the list of albums is refreshed ie. fetched from radiostudent.si anew.
