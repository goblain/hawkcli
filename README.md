## Usage

    HEADER=$(hawk header -u <url> -i <id> -k <key> -m <method> -a <app>)
    curl -H "Authorization: $HEADER" ...

## Build

go build -o hawk