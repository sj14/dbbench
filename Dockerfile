FROM golang:1.20-alpine3.17 AS build

WORKDIR /app

COPY . .
RUN go mod download

RUN go get -v -t -d ./...
RUN go build -v ./cmd/dbbench/...

## Deploy
FROM alpine:3.17

WORKDIR /app

COPY --from=build /app/dbbench dbbench

ENTRYPOINT ["./dbbench"]
