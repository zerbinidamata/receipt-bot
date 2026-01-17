"""Google Cloud Speech-to-Text transcription provider."""

import logging
import os
from pathlib import Path
from google.cloud import speech_v1 as speech

logger = logging.getLogger(__name__)


class GoogleSTTTranscriber:
    """Google Cloud Speech-to-Text transcription provider."""

    def __init__(self, credentials_path: str = None):
        """
        Initialize the Google STT transcriber.

        Args:
            credentials_path: Path to Google Cloud credentials JSON
        """
        if credentials_path:
            os.environ['GOOGLE_APPLICATION_CREDENTIALS'] = credentials_path

        self.client = speech.SpeechClient()

    def transcribe(self, audio_path: str, language_code: str = "en-US") -> str:
        """
        Transcribe an audio file.

        Args:
            audio_path: Path to the audio file
            language_code: Language code (default: en-US)

        Returns:
            Transcribed text

        Raises:
            Exception: If transcription fails
        """
        try:
            # Read the audio file
            with open(audio_path, 'rb') as audio_file:
                content = audio_file.read()

            audio = speech.RecognitionAudio(content=content)

            config = speech.RecognitionConfig(
                encoding=speech.RecognitionConfig.AudioEncoding.MP3,
                sample_rate_hertz=16000,
                language_code=language_code,
                enable_automatic_punctuation=True,
                model='default',
            )

            # For files larger than 1 minute, use long_running_recognize
            file_size = Path(audio_path).stat().st_size
            if file_size > 10 * 1024 * 1024:  # 10MB threshold
                logger.info("Using long-running recognition for large file")
                operation = self.client.long_running_recognize(config=config, audio=audio)
                response = operation.result(timeout=300)
            else:
                response = self.client.recognize(config=config, audio=audio)

            # Combine all transcripts
            transcript_parts = []
            for result in response.results:
                if result.alternatives:
                    transcript_parts.append(result.alternatives[0].transcript)

            transcript = ' '.join(transcript_parts)
            logger.info(f"Transcribed {len(transcript)} characters")

            return transcript

        except Exception as e:
            logger.error(f"Failed to transcribe audio {audio_path}: {e}")
            raise
