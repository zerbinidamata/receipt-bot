"""Base scraper interface and data models."""

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Optional, Dict


@dataclass
class ScrapeResult:
    """Result of a scraping operation."""
    captions: str
    description: str
    transcript: Optional[str]
    original_url: str
    metadata: Dict[str, str]
    error: Optional[str] = None


class BaseScraper(ABC):
    """Base interface for all platform scrapers."""

    @abstractmethod
    async def scrape(self, url: str, transcribe: bool = True) -> ScrapeResult:
        """
        Scrape content from a URL.

        Args:
            url: The URL to scrape
            transcribe: Whether to transcribe video audio

        Returns:
            ScrapeResult with extracted content

        Raises:
            Exception: If scraping fails
        """
        pass

    @abstractmethod
    def can_handle(self, url: str) -> bool:
        """
        Check if this scraper can handle the given URL.

        Args:
            url: The URL to check

        Returns:
            True if this scraper can handle the URL
        """
        pass

    def _extract_audio_path(self, video_path: str) -> str:
        """Get the audio file path for a video."""
        return video_path.rsplit('.', 1)[0] + '.mp3'
