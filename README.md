# pun-sho

[![Latest Release](https://img.shields.io/github/v/release/doutorfinancas/pun-sho)](https://github.com/doutorfinancas/pun-sho/releases)
[![CircleCI](https://circleci.com/gh/circleci/circleci-docs.svg?style=shield)](https://circleci.com/gh/doutorfinancas/pun-sho)
[![codecov](https://codecov.io/gh/doutorfinancas/pun-sho/branch/master/graph/badge.svg?token=JewR1OJdZM)](https://codecov.io/gh/doutorfinancas/pun-sho)
[![Github Actions](https://github.com/doutorfinancas/pun-sho/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/doutorfinancas/pun-sho/actions)
[![APACHE-2.0 License](https://img.shields.io/github/license/doutorfinancas/pun-sho)](LICENSE)
[![CodeFactor](https://www.codefactor.io/repository/github/doutorfinancas/pun-sho/badge)](https://www.codefactor.io/repository/github/doutorfinancas/pun-sho)

PUNy-SHOrtener - Yet Another URL Shortener

![Panchooooo](img/pun-sho.png)

Spelled pan‧cho - ˈpãnʲ.t͡ʃo

## But, Why?

![because](img/standards.png)

props to [XKCD](https://xkcd.com/927/)

We decided that we need something that doesn't exist on every other project (mix all of them, and you would have it).

So, we decided to make yet another URL shortener.

## Usage
You can clone this repo or use one of the precompiled binaries available in the release section

You can also use docker, pre-made images are available for you at `docker pull ghcr.io/doutorfinancas/pun-sho:latest`
or you can:
```bash
# this API_PORT is defined in .env file or put it in env itself
export API_PORT=8080
docker run --env-file=.env -p 8080:${API_PORT} -t ghcr.io/doutorfinancas/pun-sho:latest pun-sho 
```

You should also copy the `.env.example` to `.env` and fill the values for the database.
you can use either `cockroach` or `postgres` as value for the `DB_ADAPTOR`

If you want to use `cockroach`, you can create a free account [here](https://cockroachlabs.cloud/)

### Create a short link
```bash
read -r -d '' BODY <<EOF
{                
  "link": "https://www.google.pt/",
  "TTL": "2023-03-25T23:59:59Z",
  "redirection_limit": 5,
  "qr_code": {
    "create": true,
    "width" : 50,
    "height": 50,
    "foreground_color": "#000000",
    "background_color": "#ffffff",
    "shape": "circle"
  }
}
EOF

# you could use "background_color": "transparent" to request a png without background
# by setting env property QR_PNG_LOGO to a png filepath, 
# it will overlay the logo on qrcode center

curl -XPOST http://localhost:8080/api/v1/short \
  -H 'token: ThisIsA5uper$ecureAPIToken' \
  -H 'Content-Type: application/json' \
  -d $BODY
```

This would render an answer like:
```json
{
  "id":"4b677dfe-e17a-46e7-9cd2-25a45e8cb19c",
  "link":"https://www.google.pt/",
  "TTL":"2023-03-25T23:59:59Z",
  "redirection_limit": 5,
  "created_at":"2023-03-20T10:50:38.399449Z",
  "deleted_at":null,
  "accesses":null,
  "qr_code": "data:image/png;base64,ASfojih134kjhas9f8798134lk2fasf...",
  "short_link":"https://env.configured.domain/s/SEdeyZByeP",
  "visits":0,
  "redirects":0
}
```

If you want to preview the QR code only, you can use the preview endpoint with the same body as above
No TTL exists in that endpoint though (as its only preview mode), and the link is exactly the one you sent
```bash
read -r -d '' BODY <<EOF
{                
  "link": "https://www.google.pt/",
  "qr_code": {
    "create": true,
    "width" : 50,
    "height": 50,
    "foreground_color": "#000000",
    "background_color": "#ffffff",
    "shape": "circle"
  }
}
EOF

curl -XPOST http://localhost:8080/api/v1/preview \
  -H 'token: ThisIsA5uper$ecureAPIToken' \
  -H 'Content-Type: application/json' \
  -d $BODY 
```

### Get statistics from a visited link
```bash
curl -H 'token: ThisIsA5uper$ecureAPIToken' http://localhost:8080/api/v1/short/c62cbe57-7e45-4e87-a7c1-11cfb006870b 
```

This would render an answer like ("visits" and "redirects" will only be equal to 1 if you access the link once):
```json
{
  "id":"c62cbe57-7e45-4e87-a7c1-11cfb006870b",
  "link":"https://www.google.pt/",
  "TTL":"2023-03-25T23:59:59Z",
  "redirection_limit": 5,
  "created_at":"2023-03-19T18:56:06.8404Z",
  "deleted_at":null,
  "accesses": [
    {
      "created_at":"2023-03-19T18:56:09.615403Z",
      "meta": {
        "meta_collection": [
          {
            "name":"Accept-Encoding",
            "values":["gzip, deflate, br"]
          },
          {
            "name":"Accept-Language",
            "values":["pt-PT,pt;q=0.9,en-GB;q=0.8,en;q=0.7,en-US;q=0.6,es;q=0.5"]
          }
        ]
      },
      "user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36",
      "ip":"127.0.0.1",
      "extra":"Map: map[]",
      "os":"macOS 10.15.7",
      "browser":"Chrome 110.0.0.0",
      "status":"redirected"
    }
  ],
  "short_link":"https://env.configured.domain/s/SE345ZByeP",
  "visits":1,
  "redirects":1
}
```

### Get a list of links
```bash
curl -H 'token: ThisIsA5uper$ecureAPIToken' http://localhost:8080/api/v1/short/?limit=20&offset=0
```

Which will return a list of short links, using pagination.

### Deleting a link to make it inaccessible
```bash
curl -H 'token: ThisIsA5uper$ecureAPIToken' -XDELETE http://localhost:8080/api/v1/short/c62cbe57-7e45-4e87-a7c1-11cfb006870b
```

## Tests
You can execute all the tests of the application by using `make test`.

If you want to only execute one of the types we have, then you can run:
- `make test/go` for Go tests 
- `make test/http-requests` for http-requests tests
(uses Docker but you can, locally, execute them through Intellij IDEA or through [http-client cli](https://www.jetbrains.com/help/idea/http-client-cli.html)).

## DB migrations
This project uses database migrations.
For any changes on the DB structure to be dealt with or replicated or rolled-back we use a [migration tool](https://github.com/golang-migrate/migrate).

### Install golang-migrate
```shell
# cockroach
go install -tags 'cockroachdb' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
# postgres
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Create new migration files
```shell
make migration/create
```

### Migrate
To go versions up:
```shell
make migration/up
```

To go versions down:
```shell
make migration/clean
```

## Releases
We are currently working actively in the project and as such there still isn't a closed API.

However, we are already using it in production, in our own projects.

We consider this to be an MVP and as such use it at your own risk.

we will be releasing 0.X until we bind a contract to the API

## Next Steps
- [ ] Define stable contract version
- [ ] Add GUI (web based) with:
  - [ ] Base login page
  - [ ] Dashboard with overview
  - [ ] Ability to track a specific link data
  - [ ] Show list of links with filters (by date range, status)
- [ ] Allow better security (oauth2 or even simple jwt)
- [ ] Add GitHub pages with openapi/swagger definition
