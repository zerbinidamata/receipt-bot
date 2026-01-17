"""YouTube scraper implementation."""

import logging
from typing import Optional
from .base import BaseScraper, ScrapeResult
from ..video.downloader import VideoDownloader
from ..video.audio_extractor import AudioExtractor
from ..video.transcriber import create_transcriber
from ..utils.cleanup import cleanup_files

logger = logging.getLogger(__name__)


class YouTubeScraper(BaseScraper):
    """Scraper for YouTube videos."""

    def __init__(self, output_dir: str = "/tmp/recipe-bot"):
        """
        Initialize YouTube scraper.

        Args:
            output_dir: Directory for temporary files
        """
        self.downloader = VideoDownloader(output_dir)
        self.audio_extractor = AudioExtractor()
        self.transcriber = create_transcriber()

    def can_handle(self, url: str) -> bool:
        """Check if this scraper can handle the URL."""
        url_lower = url.lower()
        return 'youtube.com' in url_lower or 'youtu.be' in url_lower

    async def scrape(self, url: str, transcribe: bool = True) -> ScrapeResult:
        """
        Scrape a YouTube video.

        Args:
            url: YouTube video URL
            transcribe: Whether to transcribe the audio

        Returns:
            ScrapeResult with extracted content
        """
        video_path = None
        audio_path = None

        try:
            logger.info(f"Scraping YouTube video: {url}")

            # Download video and extract metadata
            download_result = self.downloader.download(url, platform='youtube')
            video_path = download_result['video_path']

            captions = download_result.get('description', '')
            metadata = {
                'title': download_result.get('title', ''),
                'author': download_result.get('author', ''),
                'duration': str(download_result.get('duration', 0)),
            }

            transcript = ""
            if transcribe:
                try:
                    # Extract audio from video
                    logger.info("Extracting audio from video")
                    audio_path = self.audio_extractor.extract_audio(video_path)

                    # Transcribe audio
                    logger.info("Transcribing audio")
                    transcript = self.transcriber.transcribe(audio_path)

                except Exception as e:
                    logger.error(f"Transcription failed: {e}")
                    # Continue without transcript

            result = ScrapeResult(
                captions=captions,
                description=captions,
                transcript=transcript,
                original_url=url,
                metadata=metadata,
            )

            logger.info(f"Successfully scraped YouTube video: {metadata.get('title')}")
            return result

        except Exception as e:
            logger.error(f"Failed to scrape YouTube video {url}: {e}")
            return ScrapeResult(
                captions="",
                description="",
                transcript="",
                original_url=url,
                metadata={},
                error=str(e),
            )

        finally:
            # Clean up temporary files
            cleanup_files(video_path, audio_path)
