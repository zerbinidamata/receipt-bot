"""Scraper factory for selecting the appropriate scraper based on URL."""

import logging
from typing import Optional
from .base import BaseScraper
from .youtube import YouTubeScraper
from .tiktok import TikTokScraper
from .instagram import InstagramScraper
from .web import WebScraper
from ..utils.url_parser import detect_platform, Platform

logger = logging.getLogger(__name__)


class ScraperFactory:
    """Factory for creating platform-specific scrapers."""

    def __init__(self, output_dir: str = "/tmp/recipe-bot"):
        """
        Initialize the scraper factory.

        Args:
            output_dir: Directory for temporary files
        """
        self.output_dir = output_dir
        self._scrapers = {
            Platform.PLATFORM_YOUTUBE: YouTubeScraper(output_dir),
            Platform.PLATFORM_TIKTOK: TikTokScraper(output_dir),
            Platform.PLATFORM_INSTAGRAM: InstagramScraper(output_dir),
            Platform.PLATFORM_WEB: WebScraper(),
        }

    def get_scraper(self, url: str, platform: Optional[Platform] = None) -> BaseScraper:
        """
        Get the appropriate scraper for a URL.

        Args:
            url: The URL to scrape
            platform: Optional platform hint

        Returns:
            Appropriate BaseScraper instance
        """
        # Detect platform if not provided
        if platform is None or platform == Platform.PLATFORM_UNKNOWN:
            platform = detect_platform(url)

        # Get the specific scraper or fall back to web scraper
        scraper = self._scrapers.get(platform)

        if scraper is None or platform == Platform.PLATFORM_WEB:
            logger.info(f"Using web scraper for {url}")
            return self._scrapers[Platform.PLATFORM_WEB]

        # Verify the scraper can handle this URL
        if not scraper.can_handle(url):
            logger.warning(f"Platform scraper cannot handle {url}, using web scraper")
            return self._scrapers[Platform.PLATFORM_WEB]

        logger.info(f"Using {platform.name} scraper for {url}")
        return scraper


async def scrape_url(url: str, platform: Optional[Platform] = None, transcribe: bool = True) -> 'ScrapeResult':
    """
    Convenience function to scrape a URL.

    Args:
        url: The URL to scrape
        platform: Optional platform hint
        transcribe: Whether to transcribe video audio

    Returns:
        ScrapeResult with extracted content
    """
    factory = ScraperFactory()
    scraper = factory.get_scraper(url, platform)
    return await scraper.scrape(url, transcribe)
