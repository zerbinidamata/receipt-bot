"""gRPC servicer implementations."""

import logging
import asyncio
from typing import Optional

# Import generated proto files
from src import scraper_pb2

from src.scrapers.factory import ScraperFactory
from src.utils.url_parser import Platform

logger = logging.getLogger(__name__)


class ScraperServicer:
    """Implementation of ScraperService."""

    def __init__(self, output_dir: str = "/tmp/recipe-bot"):
        """
        Initialize the scraper servicer.

        Args:
            output_dir: Directory for temporary files
        """
        self.factory = ScraperFactory(output_dir)
        logger.info("ScraperServicer initialized")

    async def ScrapeContent(self, request, context):
        """
        Handle ScrapeContent RPC.

        Args:
            request: ScrapeRequest proto message
            context: gRPC context

        Returns:
            ScrapeResponse proto message
        """
        try:
            logger.info(f"Received scrape request for URL: {request.url}")

            # Convert proto Platform to our Platform enum
            platform = self._convert_platform(request.platform)

            # Get appropriate scraper
            scraper = self.factory.get_scraper(request.url, platform)

            # Perform scraping
            result = await scraper.scrape(
                url=request.url,
                transcribe=request.transcribe
            )

            # Build response
            response = scraper_pb2.ScrapeResponse(
                captions=result.captions,
                transcript=result.transcript or "",
                original_url=result.original_url,
                metadata=result.metadata,
            )

            if result.error:
                response.error.message = result.error
                response.error.code = "SCRAPING_ERROR"

            logger.info(f"Successfully scraped {request.url}")
            return response

        except Exception as e:
            logger.error(f"Error in ScrapeContent: {e}", exc_info=True)

            return scraper_pb2.ScrapeResponse(
                captions="",
                transcript="",
                original_url=request.url,
                metadata={},
                error=scraper_pb2.Error(
                    message=str(e),
                    code="INTERNAL_ERROR"
                )
            )

    def _convert_platform(self, proto_platform: int) -> Optional[Platform]:
        """
        Convert proto Platform enum to our Platform enum.

        Args:
            proto_platform: Proto platform enum value

        Returns:
            Our Platform enum value or None
        """
        mapping = {
            0: Platform.PLATFORM_UNKNOWN,
            1: Platform.PLATFORM_TIKTOK,
            2: Platform.PLATFORM_YOUTUBE,
            3: Platform.PLATFORM_INSTAGRAM,
            4: Platform.PLATFORM_WEB,
        }
        return mapping.get(proto_platform, Platform.PLATFORM_UNKNOWN)
