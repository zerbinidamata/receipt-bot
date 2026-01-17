"""Audio extraction from video files using FFmpeg."""

import logging
import ffmpeg
from pathlib import Path

logger = logging.getLogger(__name__)


class AudioExtractor:
    """Extracts audio from video files."""

    @staticmethod
    def extract_audio(video_path: str, audio_path: str = None) -> str:
        """
        Extract audio from a video file.

        Args:
            video_path: Path to the input video file
            audio_path: Path for the output audio file (optional)

        Returns:
            Path to the extracted audio file

        Raises:
            Exception: If extraction fails
        """
        if audio_path is None:
            # Generate audio path from video path
            video_path_obj = Path(video_path)
            audio_path = str(video_path_obj.with_suffix('.mp3'))

        try:
            # Extract audio using ffmpeg
            stream = ffmpeg.input(video_path)
            stream = ffmpeg.output(
                stream,
                audio_path,
                acodec='libmp3lame',
                audio_bitrate='128k',
                ar='16000',  # 16kHz sample rate (good for speech recognition)
                ac=1,  # Mono channel
            )
            ffmpeg.run(stream, overwrite_output=True, quiet=True)

            logger.info(f"Extracted audio to: {audio_path}")
            return audio_path

        except ffmpeg.Error as e:
            logger.error(f"FFmpeg error extracting audio from {video_path}: {e}")
            raise
        except Exception as e:
            logger.error(f"Failed to extract audio from {video_path}: {e}")
            raise

    @staticmethod
    def get_audio_duration(audio_path: str) -> float:
        """
        Get the duration of an audio file in seconds.

        Args:
            audio_path: Path to the audio file

        Returns:
            Duration in seconds

        Raises:
            Exception: If probe fails
        """
        try:
            probe = ffmpeg.probe(audio_path)
            duration = float(probe['format']['duration'])
            return duration
        except Exception as e:
            logger.error(f"Failed to get audio duration for {audio_path}: {e}")
            raise
