FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN apk add --no-cache git ca-certificates && \
    go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/video-store ./main.go

FROM alpine:3.18
RUN apk add --no-cache ffmpeg ca-certificates && update-ca-certificates
RUN apk add --no-cache ca-certificates && update-ca-certificates || true

WORKDIR /app
COPY --from=builder /out/video-store /app/video-store

ENV VIDEO_STORAGE_PATH=/var/videos
RUN mkdir -p ${VIDEO_STORAGE_PATH}
VOLUME ["/var/videos"]

EXPOSE 8080

ENTRYPOINT ["/app/video-store"]