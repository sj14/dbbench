FROM golang:1-alpine AS build

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -v ./cmd/dbbench/...

## Deploy
FROM gcr.io/distroless/static-debian13

WORKDIR /app

COPY --from=build /app/dbbench dbbench

ENTRYPOINT ["./dbbench"]
