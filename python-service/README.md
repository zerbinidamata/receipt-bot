# Python Scraping Service

This is the Python microservice responsible for:
- Scraping content from TikTok, YouTube, Instagram, and web pages
- Downloading videos temporarily
- Extracting audio from videos
- Transcribing audio to text using Google Cloud Speech-to-Text or Whisper
- Serving content via gRPC to the Go service

## Setup

### 1. Install Python Dependencies

```bash
cd python-service
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
```

### 2. Install FFmpeg

FFmpeg is required for audio extraction:

**macOS:**
```bash
brew install ffmpeg
```

**Ubuntu/Debian:**
```bash
sudo apt-get install ffmpeg
```

**Windows:**
Download from [ffmpeg.org](https://ffmpeg.org/download.html)

### 3. Configure Environment

Copy the example environment file:
```bash
cp .env.example .env
```

Edit `.env` and set:
- `GOOGLE_CLOUD_CREDENTIALS_PATH` - Path to your Google Cloud credentials JSON
- `TRANSCRIPTION_PROVIDER` - Either "google-stt" or "whisper"
- Other optional settings

### 4. Generate Protocol Buffer Code

From the project root:
```bash
cd ../proto
make generate-python
```

This generates:
- `scraper_pb2.py` - Message definitions
- `scraper_pb2_grpc.py` - Service definitions

### 5. Run the Server

```bash
cd python-service
python run_server.py
```

The server will start on port 50051 (configurable via `GRPC_PORT`).

## Architecture

### Scrapers

Each platform has its own scraper:

- **YouTube** (`scrapers/youtube.py`): Uses yt-dlp to download videos and extract metadata
- **TikTok** (`scrapers/tiktok.py`): Uses yt-dlp for TikTok videos
- **Instagram** (`scrapers/instagram.py`): Uses instaloader for posts/reels
- **Web** (`scrapers/web.py`): Uses BeautifulSoup to extract recipe schema or general content

### Video Processing

- **Downloader** (`video/downloader.py`): Downloads videos using yt-dlp
- **Audio Extractor** (`video/audio_extractor.py`): Extracts audio from video using FFmpeg
- **Transcriber** (`video/transcriber.py`): Transcribes audio to text

### Transcription Providers

- **Google Speech-to-Text** (`video/transcription_providers/google_stt.py`): Free tier (60 min/month)
- **Whisper** (`video/transcription_providers/whisper_provider.py`): OpenAI Whisper (API or local)

### Factory Pattern

The `ScraperFactory` (`scrapers/factory.py`) selects the appropriate scraper based on the URL.

## Usage

The service exposes a single gRPC endpoint: `ScrapeContent`

**Request:**
```protobuf
message ScrapeRequest {
  string url = 1;
  Platform platform = 2;  // Optional hint
  bool download_video = 3;
  bool transcribe = 4;
}
```

**Response:**
```protobuf
message ScrapeResponse {
  string captions = 1;
  string transcript = 2;
  string original_url = 3;
  map<string, string> metadata = 4;
  Error error = 5;
}
```

## File Cleanup

All downloaded videos and extracted audio files are automatically deleted after processing to save disk space and reduce costs.

## Logging

Logs are written to stdout with configurable level via `LOG_LEVEL` environment variable.

## Development

### Adding a New Platform Scraper

1. Create a new file in `scrapers/` (e.g., `facebook.py`)
2. Implement the `BaseScraper` interface
3. Add to `ScraperFactory` in `scrapers/factory.py`
4. Update `url_parser.py` to detect the platform

### Adding a New Transcription Provider

1. Create a new file in `video/transcription_providers/`
2. Implement the transcription interface
3. Add to `Transcriber` in `video/transcriber.py`

## Troubleshooting

**Import errors for proto files:**
```bash
cd ../proto
make generate-python
```

**FFmpeg not found:**
Install FFmpeg for your platform (see Setup section)

**Google Cloud authentication errors:**
Ensure `GOOGLE_CLOUD_CREDENTIALS_PATH` points to a valid JSON key file

**yt-dlp download failures:**
yt-dlp may need updates to handle platform changes:
```bash
pip install --upgrade yt-dlp
```
