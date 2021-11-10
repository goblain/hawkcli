#!/bin/bash

URL=http://localhost:8000/platform/list
ID=hgW8HQMGS3cnqCsV
KEY=hwJRswos3qkbzzVNrZfcaMUoHyswqUMMre8VnSUYsn59hyxfZREvKH5WKjwY37sy
APP=b9109141ec8f44afbbb27d41fc5294af
METHOD=POST

HEADER=$(go run main.go header -i $ID -k $KEY -a $APP -m $METHOD -u $URL)

curl -X $METHOD -H "Authorization: $HEADER" -d '{"name": "static"}' $URL

#kfFD3ESNZkSieBq0ccxq0C9YotCaRKwUniJbFkzuxBHwnegvyXFTvW9bd2sT5+w8hcHTw7ZpUJlKJ/09Zhkc0dROxnfED63kS/0D2y5fAvuOy/zTgqfcXNYfuOI=

export HAWK_ID=hgW8HQMGS3cnqCsV
export HAWK_KEY=hwJRswos3qkbzzVNrZfcaMUoHyswqUMMre8VnSUYsn59hyxfZREvKH5WKjwY37sy
export HAWK_APP=b9109141ec8f44afbbb27d41fc5294af