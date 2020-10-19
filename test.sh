#!/bin/bash

URL=
HOST=
ID=
KEY=
APP=
METHOD=POST

HEADER=$(go run main.go header -i $ID -k $KEY -a $APP -m $METHOD -u $URL)

curl -X $METHOD -H "Authorization: $HEADER" -d '{"name": "static"}' $URL