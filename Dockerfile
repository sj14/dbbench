FROM golang:1-alpine AS build

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -v ./cmd/dbbench/...

## Deploy

# cc image is necessary for turso
FROM gcr.io/distroless/cc-debian13

WORKDIR /app

COPY --from=build /app/dbbench dbbench

ENTRYPOINT ["./dbbench"]
