# Start from golang base image
FROM golang:1.20.1-alpine3.17 as build

LABEL org.opencontainers.image.source https://github.com/doutorfinancas/pun-sho

COPY . /pun_sho

WORKDIR /pun_sho

RUN go mod download

RUN CGO_ENABLED=0 go build -o /pun-sho main.go

FROM alpine:3.17.2

COPY --from=build /pun-sho /pun-sho
COPY ./img/logo_df.png /img/logo_df.png

RUN chmod +x /pun-sho

ENTRYPOINT ["/pun-sho"]
