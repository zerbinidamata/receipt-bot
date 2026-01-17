"""Instagram scraper implementation."""

import logging
import instaloader
from pathlib import Path
from .base import BaseScraper, ScrapeResult
from ..video.audio_extractor import AudioExtractor
from ..video.transcriber import create_transcriber
from ..utils.cleanup import cleanup_files

logger = logging.getLogger(__name__)


class InstagramScraper(BaseScraper):
    """Scraper for Instagram posts/reels."""

    def __init__(self, output_dir: str = "/tmp/recipe-bot"):
        """
        Initialize Instagram scraper.

        Args:
            output_dir: Directory for temporary files
        """
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)
        self.loader = instaloader.Instaloader(
            dirname_pattern=str(self.output_dir),
            filename_pattern='{shortcode}',
            download_pictures=False,
            download_video_thumbnails=False,
            download_geotags=False,
            download_comments=False,
            save_metadata=False,
            compress_json=False,
        )
        self.audio_extractor = AudioExtractor()
        self.transcriber = create_transcriber()

    def can_handle(self, url: str) -> bool:
        """Check if this scraper can handle the URL."""
        return 'instagram.com' in url.lower()

    async def scrape(self, url: str, transcribe: bool = True) -> ScrapeResult:
        """
        Scrape an Instagram post/reel.

        Args:
            url: Instagram post/reel URL
            transcribe: Whether to transcribe the audio (for videos)

        Returns:
            ScrapeResult with extracted content
        """
        video_path = None
        audio_path = None

        try:
            logger.info(f"Scraping Instagram post: {url}")

            # Extract shortcode from URL
            shortcode = self._extract_shortcode(url)
            if not shortcode:
                raise ValueError("Could not extract shortcode from URL")

            # Get post
            post = instaloader.Post.from_shortcode(
                self.loader.context,
                shortcode
            )

            captions = post.caption or ""
            metadata = {
                'title': captions[:100] if captions else "Instagram Post",
                'author': post.owner_username,
                'likes': str(post.likes),
            }

            transcript = ""

            # Only process if it's a video
            if post.is_video:
                try:
                    # Download the video
                    video_filename = f"{shortcode}.mp4"
                    video_path = str(self.output_dir / video_filename)

                    self.loader.download_post(post, target=shortcode)

                    # Find the downloaded video file
                    downloaded_files = list(self.output_dir.glob(f"{shortcode}*.mp4"))
                    if downloaded_files:
                        video_path = str(downloaded_files[0])

                        if transcribe:
                            # Extract audio
                            logger.info("Extracting audio from Instagram video")
                            audio_path = self.audio_extractor.extract_audio(video_path)

                            # Transcribe audio
                            logger.info("Transcribing Instagram audio")
                            transcript = self.transcriber.transcribe(audio_path)

                except Exception as e:
                    logger.error(f"Failed to process Instagram video: {e}")
                    # Continue without transcript

            result = ScrapeResult(
                captions=captions,
                description=captions,
                transcript=transcript,
                original_url=url,
                metadata=metadata,
            )

            logger.info(f"Successfully scraped Instagram post by {metadata.get('author')}")
            return result

        except Exception as e:
            logger.error(f"Failed to scrape Instagram post {url}: {e}")
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
            # Clean up instaloader files
            if video_path:
                shortcode = self._extract_shortcode(url)
                if shortcode:
                    for file in self.output_dir.glob(f"{shortcode}*"):
                        try:
                            file.unlink()
                        except Exception:
                            pass

    def _extract_shortcode(self, url: str) -> str:
        """
        Extract Instagram shortcode from URL.

        Args:
            url: Instagram URL

        Returns:
            Shortcode or empty string
        """
        # Handle different Instagram URL formats
        # https://www.instagram.com/p/SHORTCODE/
        # https://www.instagram.com/reel/SHORTCODE/
        parts = url.rstrip('/').split('/')
        if 'p' in parts or 'reel' in parts:
            idx = parts.index('p') if 'p' in parts else parts.index('reel')
            if idx + 1 < len(parts):
                return parts[idx + 1]
        return ""
