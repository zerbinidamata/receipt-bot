"""Main transcription interface that uses different providers."""

import logging
import os
from typing import Optional

logger = logging.getLogger(__name__)


class Transcriber:
    """Main transcription service that delegates to specific providers."""

    def __init__(self, provider: str = "elevenlabs", **kwargs):
        """
        Initialize the transcriber with a specific provider.

        Args:
            provider: Provider name ('elevenlabs', 'google-stt', or 'whisper')
            **kwargs: Provider-specific arguments
        """
        self.provider = provider

        if provider == "elevenlabs":
            from .transcription_providers.elevenlabs_stt import ElevenLabsTranscriber
            api_key = kwargs.get('api_key', os.getenv('ELEVENLABS_API_KEY'))
            self.transcriber = ElevenLabsTranscriber(api_key=api_key)

        elif provider == "google-stt":
            from .transcription_providers.google_stt import GoogleSTTTranscriber
            credentials_path = kwargs.get(
                'credentials_path',
                os.getenv('GOOGLE_CLOUD_CREDENTIALS_PATH')
            )
            self.transcriber = GoogleSTTTranscriber(credentials_path)

        elif provider == "whisper":
            from .transcription_providers.whisper_provider import WhisperTranscriber
            use_api = kwargs.get('use_api', False)
            api_key = kwargs.get('api_key', os.getenv('OPENAI_API_KEY'))
            model_name = kwargs.get('model_name', 'base')
            self.transcriber = WhisperTranscriber(
                model_name=model_name,
                use_api=use_api,
                api_key=api_key
            )

        else:
            raise ValueError(f"Unknown transcription provider: {provider}. "
                           f"Supported: elevenlabs, google-stt, whisper")

        logger.info(f"Initialized transcriber with provider: {provider}")

    def transcribe(self, audio_path: str, language: str = None) -> str:
        """
        Transcribe an audio file.

        Args:
            audio_path: Path to the audio file
            language: Language code (optional)

        Returns:
            Transcribed text

        Raises:
            Exception: If transcription fails
        """
        try:
            if self.provider == "elevenlabs":
                # ElevenLabs uses language codes like "eng", "spa", "por"
                language_code = self._convert_to_elevenlabs_language(language)
                return self.transcriber.transcribe(audio_path, language_code=language_code)
            elif self.provider == "google-stt":
                language_code = language or "en-US"
                return self.transcriber.transcribe(audio_path, language_code)
            else:  # whisper
                return self.transcriber.transcribe(audio_path, language)

        except Exception as e:
            logger.error(f"Transcription failed: {e}")
            raise

    def _convert_to_elevenlabs_language(self, language: str) -> Optional[str]:
        """Convert language codes to ElevenLabs format."""
        if not language or language.strip() == "":
            return None  # Auto-detect

        # Map common language codes to ElevenLabs format
        language_map = {
            "en": "eng", "en-US": "eng", "en-GB": "eng",
            "es": "spa", "es-ES": "spa", "es-MX": "spa",
            "pt": "por", "pt-BR": "por", "pt-PT": "por",
            "fr": "fra", "fr-FR": "fra",
            "de": "deu", "de-DE": "deu",
            "it": "ita", "it-IT": "ita",
            "ja": "jpn", "ja-JP": "jpn",
            "ko": "kor", "ko-KR": "kor",
            "zh": "cmn", "zh-CN": "cmn", "zh-TW": "cmn",
        }
        return language_map.get(language, language)


def create_transcriber(provider: str = None) -> Transcriber:
    """
    Factory function to create a transcriber.

    Args:
        provider: Provider name (if None, uses TRANSCRIPTION_PROVIDER env var)
                  Supported: 'elevenlabs' (default), 'google-stt', 'whisper'

    Returns:
        Transcriber instance
    """
    if provider is None:
        provider = os.getenv('TRANSCRIPTION_PROVIDER', 'elevenlabs')

    return Transcriber(provider=provider)
