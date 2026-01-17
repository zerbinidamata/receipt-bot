"""Video download utilities using yt-dlp."""

import logging
import yt_dlp
from pathlib import Path
from typing import Optional, Dict

logger = logging.getLogger(__name__)


class VideoDownloader:
    """Handles video downloading from various platforms."""

    def __init__(self, output_dir: str = "/tmp/recipe-bot"):
        """
        Initialize the video downloader.

        Args:
            output_dir: Directory to save downloaded videos
        """
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)

    def download(self, url: str, platform: str = None) -> Dict[str, str]:
        """
        Download a video from a URL.

        Args:
            url: The video URL
            platform: Platform name (for optimization)

        Returns:
            Dictionary with 'video_path', 'title', 'description', 'author'

        Raises:
            Exception: If download fails
        """
        output_template = str(self.output_dir / '%(id)s.%(ext)s')

        ydl_opts = {
            'format': 'best[ext=mp4]/best',  # Prefer MP4
            'outtmpl': output_template,
            'quiet': True,
            'no_warnings': True,
            'extract_flat': False,
        }

        try:
            with yt_dlp.YoutubeDL(ydl_opts) as ydl:
                info = ydl.extract_info(url, download=True)

                video_id = info.get('id', 'unknown')
                ext = info.get('ext', 'mp4')
                video_path = str(self.output_dir / f"{video_id}.{ext}")

                result = {
                    'video_path': video_path,
                    'title': info.get('title', ''),
                    'description': info.get('description', ''),
                    'author': info.get('uploader', '') or info.get('channel', ''),
                    'duration': info.get('duration', 0),
                }

                logger.info(f"Downloaded video: {result['title']}")
                return result

        except Exception as e:
            logger.error(f"Failed to download video from {url}: {e}")
            raise

    def extract_metadata(self, url: str) -> Dict[str, str]:
        """
        Extract video metadata without downloading.

        Args:
            url: The video URL

        Returns:
            Dictionary with metadata

        Raises:
            Exception: If extraction fails
        """
        ydl_opts = {
            'quiet': True,
            'no_warnings': True,
            'extract_flat': True,
        }

        try:
            with yt_dlp.YoutubeDL(ydl_opts) as ydl:
                info = ydl.extract_info(url, download=False)

                return {
                    'title': info.get('title', ''),
                    'description': info.get('description', ''),
                    'author': info.get('uploader', '') or info.get('channel', ''),
                    'duration': info.get('duration', 0),
                    'thumbnail': info.get('thumbnail', ''),
                }

        except Exception as e:
            logger.error(f"Failed to extract metadata from {url}: {e}")
            raise
