# pun-sho

[![Latest Release](https://img.shields.io/github/v/release/doutorfinancas/pun-sho)](https://github.com/doutorfinancas/pun-sho/releases)
[![CircleCI](https://circleci.com/gh/circleci/circleci-docs.svg?style=shield)](https://circleci.com/gh/doutorfinancas/pun-sho)
[![codecov](https://codecov.io/gh/doutorfinancas/pun-sho/branch/master/graph/badge.svg?token=JewR1OJdZM)](https://codecov.io/gh/doutorfinancas/pun-sho)
[![Github Actions](https://github.com/doutorfinancas/pun-sho/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/doutorfinancas/pun-sho/actions)
[![APACHE-2.0 License](https://img.shields.io/github/license/doutorfinancas/pun-sho)](LICENSE)

PUNy-SHOrtener - Yet Another URL Shortener

![Panchooooo](img/pun-sho.png)

Spelled pan‧cho - ˈpãnʲ.t͡ʃo

## But, Why?

![because](img/standards.png)

props to [XKCD](https://xkcd.com/927/)

We decided that we need something that doesn't exist on every other project (mix all of them 
, and you would have it).

So, we decided to make yet another URL shortener.

## Usage
you can clone this repo or use one of the precompiled binaries available in the release section

you can also use docker, pre-made images are available for you at `docker pull ghcr.io/doutorfinancas/pun-sho:latest`
or you can
```bash
# this API_PORT is defined in .env file or put it in env itself
export API_PORT=8080
docker run --env-file=.env -p 8080:${API_PORT} -t ghcr.io/doutorfinancas/pun-sho:latest pun-sho 
```

you should also copy the `.env.example` to `.env` and fill the values for the database.
you can use either `cockroach` or `postgres` as value for the `DB_ADAPTOR`

if you use `cockroach`, you can create a free account [here](https://cockroachlabs.cloud/)

## DB migrations
This project uses database migrations.
For any changes on the DB structure to be dealt with or replicated or rolled-back we use a [migration tool](https://github.com/golang-migrate/migrate)

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

### Create a short link
```bash
curl -H 'token: Whatever_Token_you_put_in_your_env' -XPOST -d '{"link": "https://www.google.pt/", "TTL": "2023-03-25T23:59:59Z"}' https://yourdomain.something/api/v1/short
```

this would render an answer like:
```json
{
  "id":"4b677dfe-e17a-46e7-9cd2-25a45e8cb19c",
  "link":"https://www.google.pt/",
  "TTL":"2023-03-25T23:59:59Z",
  "created_at":"2023-03-20T10:50:38.399449Z",
  "deleted_at":null,
  "accesses":null,
  "short_link":"https://env.configured.domain/s/SEdeyZByeP",
  "visits":0,
  "redirects":0
}
```

### Get statistics from a visited link
```bash
curl -H 'token: ThisIsA5uper$ecureAPIToken' https://yourdomain.something/api/v1/short/c62cbe57-7e45-4e87-a7c1-11cfb006870b 
```

this would render an answer like:
```json
{
  "id":"c62cbe57-7e45-4e87-a7c1-11cfb006870b",
  "link":"https://www.google.pt/",
  "TTL":"2023-03-25T23:59:59Z",
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
curl -H 'token: ThisIsA5uper$ecureAPIToken' https://yourdomain.something/api/v1/short/?limit=20&offset=0
```

will return a list of short links, using pagination

### Deleting a link to make it inaccessible
```bash
curl -H 'token: ThisIsA5uper$ecureAPIToken' -XDELETE https://yourdomain.something/api/v1/short/c62cbe57-7e45-4e87-a7c1-11cfb006870b
```

## Releases
We are currently working actively in the project and as such there still isn't a closed API.

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
