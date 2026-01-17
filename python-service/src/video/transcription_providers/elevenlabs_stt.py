"""ElevenLabs Speech-to-Text transcription provider."""

import os
import logging
from io import BytesIO
from typing import Optional

logger = logging.getLogger(__name__)


class ElevenLabsTranscriber:
    """Transcriber using ElevenLabs Speech-to-Text API (Scribe v1)."""

    def __init__(self, api_key: Optional[str] = None):
        """
        Initialize ElevenLabs transcriber.

        Args:
            api_key: ElevenLabs API key. If not provided, uses ELEVENLABS_API_KEY env var.
        """
        try:
            from elevenlabs import ElevenLabs
        except ImportError:
            raise ImportError(
                "elevenlabs package not installed. Run: poetry add elevenlabs"
            )

        self.api_key = api_key or os.getenv("ELEVENLABS_API_KEY")
        if not self.api_key:
            raise ValueError(
                "ElevenLabs API key not found. Set ELEVENLABS_API_KEY environment variable "
                "or pass api_key parameter."
            )

        self.client = ElevenLabs(api_key=self.api_key)
        logger.info("Initialized ElevenLabs transcriber")

    def transcribe(
        self,
        audio_path: str,
        language_code: Optional[str] = None,
        diarize: bool = False,
        tag_audio_events: bool = False,
    ) -> str:
        """
        Transcribe audio file to text using ElevenLabs Scribe v1.

        Args:
            audio_path: Path to the audio file
            language_code: Optional language code (e.g., "eng", "spa", "por")
                          If None, auto-detects language
            diarize: Whether to annotate speaker changes
            tag_audio_events: Whether to tag audio events like laughter, applause

        Returns:
            Transcribed text
        """
        logger.info(f"Transcribing audio file: {audio_path}")

        # Read audio file
        with open(audio_path, "rb") as audio_file:
            audio_data = BytesIO(audio_file.read())

        # Get the filename for the API
        filename = os.path.basename(audio_path)

        try:
            # Build API call parameters
            api_params = {
                "file": (filename, audio_data),
                "model_id": "scribe_v1",
                "diarize": diarize,
                "tag_audio_events": tag_audio_events,
            }
            # Only include language_code if it's a non-empty string
            if language_code and language_code.strip():
                api_params["language_code"] = language_code

            # Call ElevenLabs Speech-to-Text API
            result = self.client.speech_to_text.convert(**api_params)

            # Extract text from result
            if hasattr(result, "text"):
                transcript = result.text
            elif isinstance(result, dict) and "text" in result:
                transcript = result["text"]
            else:
                # Try to get transcript from segments if available
                transcript = self._extract_text_from_result(result)

            logger.info(f"Transcription completed. Length: {len(transcript)} chars")
            return transcript

        except Exception as e:
            logger.error(f"ElevenLabs transcription failed: {e}")
            raise

    def _extract_text_from_result(self, result) -> str:
        """
        Extract text from various result formats.

        Args:
            result: The API response object

        Returns:
            Extracted text
        """
        # If result has segments, join them
        if hasattr(result, "segments"):
            return " ".join(
                segment.text for segment in result.segments if hasattr(segment, "text")
            )

        # If result has words, join them
        if hasattr(result, "words"):
            return " ".join(
                word.text for word in result.words if hasattr(word, "text")
            )

        # Try converting to string as fallback
        return str(result)

    def transcribe_with_timestamps(
        self,
        audio_path: str,
        language_code: Optional[str] = None,
    ) -> dict:
        """
        Transcribe audio file with word-level timestamps.

        Args:
            audio_path: Path to the audio file
            language_code: Optional language code

        Returns:
            Dictionary with 'text' and 'words' (with timestamps)
        """
        logger.info(f"Transcribing with timestamps: {audio_path}")

        with open(audio_path, "rb") as audio_file:
            audio_data = BytesIO(audio_file.read())

        filename = os.path.basename(audio_path)

        try:
            # Build API call parameters
            api_params = {
                "file": (filename, audio_data),
                "model_id": "scribe_v1",
                "diarize": True,
                "tag_audio_events": True,
            }
            # Only include language_code if it's a non-empty string
            if language_code and language_code.strip():
                api_params["language_code"] = language_code

            result = self.client.speech_to_text.convert(**api_params)

            # Build response with timestamps if available
            response = {
                "text": "",
                "words": [],
                "segments": [],
            }

            if hasattr(result, "text"):
                response["text"] = result.text

            if hasattr(result, "words"):
                response["words"] = [
                    {
                        "text": w.text if hasattr(w, "text") else str(w),
                        "start": getattr(w, "start", None),
                        "end": getattr(w, "end", None),
                    }
                    for w in result.words
                ]

            if hasattr(result, "segments"):
                response["segments"] = [
                    {
                        "text": s.text if hasattr(s, "text") else str(s),
                        "start": getattr(s, "start", None),
                        "end": getattr(s, "end", None),
                        "speaker": getattr(s, "speaker", None),
                    }
                    for s in result.segments
                ]

            return response

        except Exception as e:
            logger.error(f"ElevenLabs transcription with timestamps failed: {e}")
            raise
