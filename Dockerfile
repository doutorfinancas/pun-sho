# Start from golang base image
FROM golang:1.20.1-alpine3.17

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git bash build-base

RUN go install github.com/cosmtrek/air@latest
RUN export PATH="${GOPATH}/bin:${PATH}"

# Expose port 8080 to the outside world
EXPOSE 8080

# Run the executable
ENTRYPOINT ["air", "run", "main.go"]