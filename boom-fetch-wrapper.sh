#!/bin/bash

awhile=3
sleep $awhile && xdg-open http://localhost:8090/tolpe & ./boom-fetch
