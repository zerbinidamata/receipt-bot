"""URL parsing and platform detection utilities."""

from enum import IntEnum
from urllib.parse import urlparse


class Platform(IntEnum):
    """Platform enumeration matching proto definition."""
    PLATFORM_UNKNOWN = 0
    PLATFORM_TIKTOK = 1
    PLATFORM_YOUTUBE = 2
    PLATFORM_INSTAGRAM = 3
    PLATFORM_WEB = 4


def detect_platform(url: str) -> Platform:
    """
    Detect the platform from a URL.

    Args:
        url: The URL to analyze

    Returns:
        Platform enum value
    """
    url_lower = url.lower()

    if 'tiktok.com' in url_lower:
        return Platform.PLATFORM_TIKTOK
    elif 'youtube.com' in url_lower or 'youtu.be' in url_lower:
        return Platform.PLATFORM_YOUTUBE
    elif 'instagram.com' in url_lower:
        return Platform.PLATFORM_INSTAGRAM
    else:
        return Platform.PLATFORM_WEB


def normalize_url(url: str) -> str:
    """
    Normalize a URL by removing tracking parameters and fragments.

    Args:
        url: The URL to normalize

    Returns:
        Normalized URL
    """
    parsed = urlparse(url)
    # Remove fragments and common tracking params
    return f"{parsed.scheme}://{parsed.netloc}{parsed.path}"


def is_valid_url(url: str) -> bool:
    """
    Check if a URL is valid.

    Args:
        url: The URL to validate

    Returns:
        True if valid, False otherwise
    """
    try:
        result = urlparse(url)
        return all([result.scheme, result.netloc])
    except Exception:
        return False
