"""Temporary file cleanup utilities."""

import os
import logging
from pathlib import Path
from typing import List

logger = logging.getLogger(__name__)


def cleanup_files(*file_paths: str) -> None:
    """
    Delete temporary files.

    Args:
        *file_paths: Variable number of file paths to delete
    """
    for file_path in file_paths:
        if not file_path:
            continue

        try:
            path = Path(file_path)
            if path.exists() and path.is_file():
                path.unlink()
                logger.debug(f"Deleted temporary file: {file_path}")
        except Exception as e:
            logger.warning(f"Failed to delete temporary file {file_path}: {e}")


def cleanup_directory(directory: str, extensions: List[str] = None) -> None:
    """
    Clean up files in a directory, optionally filtering by extension.

    Args:
        directory: Directory path to clean
        extensions: List of file extensions to delete (e.g., ['.mp4', '.mp3'])
                   If None, deletes all files
    """
    try:
        dir_path = Path(directory)
        if not dir_path.exists() or not dir_path.is_dir():
            return

        for file_path in dir_path.iterdir():
            if not file_path.is_file():
                continue

            if extensions is None or file_path.suffix in extensions:
                try:
                    file_path.unlink()
                    logger.debug(f"Deleted: {file_path}")
                except Exception as e:
                    logger.warning(f"Failed to delete {file_path}: {e}")

    except Exception as e:
        logger.error(f"Failed to cleanup directory {directory}: {e}")


def ensure_temp_directory(base_dir: str = "/tmp/recipe-bot") -> Path:
    """
    Ensure a temporary directory exists.

    Args:
        base_dir: Base directory path

    Returns:
        Path object for the directory
    """
    path = Path(base_dir)
    path.mkdir(parents=True, exist_ok=True)
    return path
