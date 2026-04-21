# Start from golang base image
FROM golang:1.25.0-alpine3.22 as build

LABEL org.opencontainers.image.source https://github.com/doutorfinancas/pun-sho

COPY . /pun_sho

WORKDIR /pun_sho

RUN go mod download

RUN CGO_ENABLED=0 go build -o /pun-sho main.go

FROM alpine:3.22

COPY --from=build /pun-sho /pun-sho
COPY --from=build /pun_sho/templates/ /templates/
COPY --from=build /pun_sho/static/ /static/
COPY ./img/logo_df.png /img/logo_df.png

RUN chmod +x /pun-sho

ENTRYPOINT ["/pun-sho"]
