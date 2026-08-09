# VidSnap — High-Speed Video & Audio Downloader

A backend-powered video downloader built with **Go** and **yt-dlp** that lets users preview and download videos or extract audio from YouTube, TikTok, Instagram, Twitter/X, Facebook, Vimeo and more — directly from a modern glassmorphic web interface.

## ✨ Key Features

- **Instant Metadata Preview (`POST /info`)**: Live thumbnail, title, duration, and uploader preview before starting a download.
- **Format & Quality Control**: Download as **MP4 (1080p, 720p, 480p, 360p)** or extract **MP3 Audio**.
- **Real-Time Progress & Speed**: Live polling with percentage updates and status feedback.
- **Real Server-Side Cancellation**: Cancel anytime to instantly kill background processes and free server resources.
- **Automated Memory & Disk TTL Cleanup**: Background worker purges expired jobs and orphan files every 15 minutes.
- **Dynamic File Titles**: Streams files with sanitized video titles (e.g. `My_Video.mp4`) and accurate Content-Type headers.
- **Container Health Check (`GET /health`)**: Built-in health probe endpoint for Docker/hosting platforms.
- **CI/CD Ready**: Integrated GitHub Actions pipeline for linting, testing, Docker builds, and deployment webhooks.

## 🛠️ Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go (Golang 1.22) |
| Downloader | yt-dlp + ffmpeg |
| Frontend | HTML5, Vanilla JS, Glassmorphic CSS (Inter Font) |
| Server | Go `net/http` |
| Container | Docker Multi-Stage Build |
| CI/CD | GitHub Actions |

## 🚀 API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| POST | `/info` | Fetch video thumbnail, title, uploader & duration |
| POST | `/download` | Start a background download job |
| POST | `/cancel/:id` | Cancel an active job and kill server process |
| GET | `/progress/:id` | Get job status, progress (0-100%), and output filename |
| GET | `/file/:id` | Stream and download the completed file |
| GET | `/health` | Server health check endpoint |

### 1. `POST /info`
```json
{
  "url": "https://www.youtube.com/watch?v=69i5TywtrSk"
}
```
**Response:**
```json
{
  "title": "Sample Video",
  "thumbnail": "https://i.ytimg.com/...",
  "duration": 214.0,
  "uploader": "Channel Name",
  "platform": "Youtube"
}
```

### 2. `POST /download`
```json
{
  "url": "https://www.youtube.com/watch?v=69i5TywtrSk",
  "format": "mp4",
  "quality": "1080p",
  "title": "Sample Video"
}
```
**Response:**
```json
{
  "job_id": "172323456789"
}
```

---

## 💻 Local Development

### Prerequisites

- Go 1.21+
- `yt-dlp` and `ffmpeg` installed in `$PATH`

```bash
# Install yt-dlp
sudo curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
sudo chmod a+rx /usr/local/bin/yt-dlp
```

### Run Locally

```bash
git clone https://github.com/YOUR_USERNAME/video-downloader.git
cd video-downloader
go run cmd/server/main.go
```
Open `http://localhost:8080` in your browser.

---

## 🌐 Free Low-Cost Hosting & Deployment Guide

> **Note on Hosting Options:**  
> VidSnap requires running binary utilities (`yt-dlp` & `ffmpeg`) and temporary file storage. **Render**, **Koyeb**, and **Fly.io** offer **free Docker hosting** that handles long-running downloads seamlessly.

### Option 1: Render (Recommended - Free Tier)

1. Push your repository to GitHub.
2. Sign up on [Render.com](https://render.com).
3. Click **New +** -> **Web Service**.
4. Connect your GitHub repository.
5. Select **Docker** as the Runtime.
6. Render will automatically detect the `Dockerfile` and deploy your app!
7. *(Optional)* Copy your Render **Deploy Hook URL** into your GitHub repository secrets as `RENDER_DEPLOY_HOOK_URL` to enable automatic CI/CD deployments on `git push`.

### Option 2: Koyeb (Free Tier Container Hosting)

1. Sign up on [Koyeb.com](https://www.koyeb.com).
2. Create a new service from your GitHub repo.
3. Select **Docker** deployment mode.
4. Set port to `8080` and deploy.

### Option 3: Local Docker Container

```bash
docker build -t vidsnap .
docker run -p 8080:8080 vidsnap
```

---

## 📜 Legal & Disclaimer

This tool is intended for **personal use only**. Always respect copyright laws and terms of service of supported platforms.

## 📄 License

MIT

