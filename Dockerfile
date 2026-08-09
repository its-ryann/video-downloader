FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o vidsnap ./cmd/server

FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (Python3, ffmpeg, curl)
RUN apk add --no-cache python3 py3-pip ffmpeg curl ca-certificates

# Install latest yt-dlp executable
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
  && chmod a+rx /usr/local/bin/yt-dlp

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy built binary and web assets
COPY --from=builder /app/vidsnap .
COPY --from=builder /app/web ./web

RUN chown -R appuser:appgroup /app /tmp

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["./vidsnap"]