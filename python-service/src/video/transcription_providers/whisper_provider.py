"""Whisper transcription provider (alternative to Google STT)."""

import logging

logger = logging.getLogger(__name__)


class WhisperTranscriber:
    """Whisper transcription provider (local or API)."""

    def __init__(self, model_name: str = "base", use_api: bool = False, api_key: str = None):
        """
        Initialize the Whisper transcriber.

        Args:
            model_name: Whisper model name (tiny, base, small, medium, large)
            use_api: Whether to use OpenAI's Whisper API
            api_key: OpenAI API key (required if use_api=True)
        """
        self.use_api = use_api
        self.model_name = model_name

        if use_api:
            if not api_key:
                raise ValueError("API key required for Whisper API")
            import openai
            self.client = openai.OpenAI(api_key=api_key)
        else:
            # Local Whisper model
            try:
                import whisper
                self.model = whisper.load_model(model_name)
                logger.info(f"Loaded Whisper model: {model_name}")
            except ImportError:
                raise ImportError(
                    "openai-whisper package not installed. "
                    "Install with: pip install openai-whisper"
                )

    def transcribe(self, audio_path: str, language: str = None) -> str:
        """
        Transcribe an audio file.

        Args:
            audio_path: Path to the audio file
            language: Language code (optional, e.g., 'en')

        Returns:
            Transcribed text

        Raises:
            Exception: If transcription fails
        """
        try:
            if self.use_api:
                return self._transcribe_api(audio_path)
            else:
                return self._transcribe_local(audio_path, language)

        except Exception as e:
            logger.error(f"Failed to transcribe audio {audio_path}: {e}")
            raise

    def _transcribe_api(self, audio_path: str) -> str:
        """Transcribe using OpenAI Whisper API."""
        with open(audio_path, 'rb') as audio_file:
            transcript = self.client.audio.transcriptions.create(
                model="whisper-1",
                file=audio_file,
                response_format="text"
            )
        return transcript

    def _transcribe_local(self, audio_path: str, language: str = None) -> str:
        """Transcribe using local Whisper model."""
        options = {}
        if language:
            options['language'] = language

        result = self.model.transcribe(audio_path, **options)
        return result['text']
