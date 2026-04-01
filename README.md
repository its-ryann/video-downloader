# VidSnap — Free Video Downloader

A backend-powered video downloader built with **Go** and **yt-dlp** that lets users download videos and audio from YouTube, TikTok, Instagram, Facebook, Twitter, Vimeo and more — directly from a clean web interface.

## Features

- Download videos as **MP4** or extract audio as **MP3**
- **Real-time progress bar** showing download percentage
- **Concurrent downloads** — handles multiple users at once (max 3 simultaneous)
- **Cancel** a download mid-way
- Auto-clears input after a successful download
- No sign up, no watermarks, no ads
- Clean dark-mode UI with support for all major platforms

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go (Golang) |
| Downloader | yt-dlp |
| Frontend | HTML, CSS, Vanilla JS |
| Server | Go net/http |

## Project Structure
```
video-downloader/
├── cmd/
│   └── server/
│       └── main.go          # Entry point
├── internal/
│   ├── api/
│   │   └── handler.go       # HTTP handlers + job tracking
│   ├── downloader/
│   │   └── downloader.go    # yt-dlp integration + progress parsing
│   └── server/
│       └── server.go        # Server setup + routing
├── web/
│   ├── index.html           # Frontend UI
│   ├── script.js            # Download logic + progress polling
│   └── style.css            # Dark mode styles
├── go.mod
└── README.md
```

## How It Works

1. User pastes a video URL and selects a format (MP4 or MP3)
2. Frontend sends a `POST /download` request to the backend
3. Backend starts a background job and returns a job ID instantly
4. Frontend polls `GET /progress/{id}` every second to get live progress
5. Once complete, frontend fetches the file from `GET /file/{id}`
6. File is streamed to the user's browser and deleted from the server

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| POST | `/download` | Start a download job |
| GET | `/progress/:id` | Get job progress (0-100%) |
| GET | `/file/:id` | Download the completed file |

### POST `/download`

**Request body:**
```json
{
  "url": "https://www.youtube.com/watch?v=...",
  "format": "mp4"
}
```

**Response:**
```json
{
  "job_id": "1234567890"
}
```

### GET `/progress/:id`

**Response:**
```json
{
  "status": "processing",
  "progress": 45,
  "error": ""
}
```

Status values: `processing`, `done`, `error`

## Local Development

### Prerequisites

- Go 1.21+
- yt-dlp installed and available in `$PATH`

**Install yt-dlp:**
```bash
sudo curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
sudo chmod a+rx /usr/local/bin/yt-dlp
```

### Run locally
```bash
git clone https://github.com/YOUR_USERNAME/video-downloader.git
cd video-downloader
go run cmd/server/main.go
```

Then open `http://localhost:8080` in your browser.

### Test the API directly
```bash
# Start a download
curl -X POST http://localhost:8080/download \
  -H "Content-Type: application/json" \
  -d '{"url":"https://www.youtube.com/watch?v=69i5TywtrSk", "format":"mp4"}' \
  --output test.mp4
```

## Deployment

This project is designed to run on any Linux server with Go and yt-dlp installed.
See deployment instructions for Railway, Render, or a VPS below.

## Limitations

- Downloads are temporary — files are deleted after being sent to the client
- Maximum 3 concurrent downloads (configurable in `server.go`)
- Large videos may take several minutes depending on server bandwidth
- Some platforms may require updated yt-dlp versions

## Legal

This tool is intended for **personal use only**. Always respect copyright laws and the terms of service of the platforms you download from. The developers are not responsible for misuse.

## License

MIT
