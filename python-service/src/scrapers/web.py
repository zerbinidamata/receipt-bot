"""Generic web scraper for recipe websites."""

import logging
import requests
from bs4 import BeautifulSoup
from typing import Optional, Dict
from .base import BaseScraper, ScrapeResult

logger = logging.getLogger(__name__)


class WebScraper(BaseScraper):
    """Scraper for generic web pages, especially recipe websites."""

    def __init__(self):
        """Initialize web scraper."""
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
        })

    def can_handle(self, url: str) -> bool:
        """Check if this scraper can handle the URL."""
        # Web scraper is the fallback for all other URLs
        return True

    async def scrape(self, url: str, transcribe: bool = False) -> ScrapeResult:
        """
        Scrape a web page for recipe content.

        Args:
            url: Web page URL
            transcribe: Not applicable for web pages

        Returns:
            ScrapeResult with extracted content
        """
        try:
            logger.info(f"Scraping web page: {url}")

            # Fetch the page
            response = self.session.get(url, timeout=30)
            response.raise_for_status()

            soup = BeautifulSoup(response.content, 'lxml')

            # Try to extract structured recipe data (schema.org)
            recipe_data = self._extract_recipe_schema(soup)
            if recipe_data:
                logger.info("Found structured recipe data")
                return ScrapeResult(
                    captions=recipe_data.get('description', ''),
                    description=recipe_data.get('description', ''),
                    transcript="",  # No transcript for web pages
                    original_url=url,
                    metadata=recipe_data,
                )

            # Fallback: extract all text content
            logger.info("No structured data found, extracting all text")
            text_content = self._extract_text_content(soup)

            metadata = {
                'title': soup.title.string if soup.title else '',
            }

            result = ScrapeResult(
                captions=text_content,
                description=text_content,
                transcript="",
                original_url=url,
                metadata=metadata,
            )

            logger.info("Successfully scraped web page")
            return result

        except Exception as e:
            logger.error(f"Failed to scrape web page {url}: {e}")
            return ScrapeResult(
                captions="",
                description="",
                transcript="",
                original_url=url,
                metadata={},
                error=str(e),
            )

    def _extract_recipe_schema(self, soup: BeautifulSoup) -> Optional[Dict]:
        """
        Extract recipe data from schema.org markup.

        Args:
            soup: BeautifulSoup object

        Returns:
            Dictionary with recipe data or None
        """
        import json

        # Look for JSON-LD schema
        scripts = soup.find_all('script', type='application/ld+json')
        for script in scripts:
            try:
                data = json.loads(script.string)

                # Handle both single objects and arrays
                if isinstance(data, list):
                    data = next((d for d in data if d.get('@type') == 'Recipe'), None)

                if data and data.get('@type') == 'Recipe':
                    # Extract recipe information
                    ingredients = data.get('recipeIngredient', [])
                    instructions = data.get('recipeInstructions', [])

                    # Build description from ingredients and instructions
                    description_parts = []

                    if ingredients:
                        description_parts.append("INGREDIENTS:")
                        if isinstance(ingredients, list):
                            description_parts.extend(ingredients)
                        description_parts.append("")

                    if instructions:
                        description_parts.append("INSTRUCTIONS:")
                        if isinstance(instructions, list):
                            for i, instruction in enumerate(instructions, 1):
                                if isinstance(instruction, dict):
                                    text = instruction.get('text', str(instruction))
                                else:
                                    text = str(instruction)
                                description_parts.append(f"{i}. {text}")
                        else:
                            description_parts.append(str(instructions))

                    return {
                        'title': data.get('name', ''),
                        'description': '\n'.join(description_parts),
                        'author': data.get('author', {}).get('name', '') if isinstance(data.get('author'), dict) else data.get('author', ''),
                        'prep_time': data.get('prepTime', ''),
                        'cook_time': data.get('cookTime', ''),
                        'servings': str(data.get('recipeYield', '')),
                    }

            except (json.JSONDecodeError, AttributeError, KeyError) as e:
                logger.debug(f"Failed to parse schema: {e}")
                continue

        return None

    def _extract_text_content(self, soup: BeautifulSoup) -> str:
        """
        Extract readable text content from the page.

        Args:
            soup: BeautifulSoup object

        Returns:
            Extracted text
        """
        # Remove script and style elements
        for element in soup(['script', 'style', 'nav', 'footer', 'header']):
            element.decompose()

        # Get text and clean it
        text = soup.get_text(separator='\n')

        # Clean up whitespace
        lines = [line.strip() for line in text.splitlines()]
        lines = [line for line in lines if line]

        return '\n'.join(lines)
