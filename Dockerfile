FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o vidsnap ./cmd/server

FROM alpine:latest

WORKDIR /app

# Install yt-dlp dependencies
RUN apk add --no-cache python3 py3-pip ffmpeg curl

# Install yt-dlp
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
  && chmod a+rx /usr/local/bin/yt-dlp

# Copy built binary and web assets
COPY --from=builder /app/vidsnap .
COPY --from=builder /app/web ./web

EXPOSE 8080

CMD ["./vidsnap"]