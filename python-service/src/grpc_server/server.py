"""gRPC server main entry point."""

import asyncio
import logging
import os
import sys
from concurrent import futures
from dotenv import load_dotenv

import grpc

# Import generated proto files (after running make generate)
try:
    from src import scraper_pb2_grpc
    from src.grpc_server.servicers import ScraperServicer
except ImportError as e:
    print("Error: Proto files not generated. Run 'make generate' first")
    print(f"Import error: {e}")
    sys.exit(1)

# Load environment variables
load_dotenv()

# Configure logging
logging.basicConfig(
    level=os.getenv('LOG_LEVEL', 'INFO'),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


def serve():
    """Start the gRPC server."""
    port = os.getenv('GRPC_PORT', '50051')
    max_workers = int(os.getenv('GRPC_MAX_WORKERS', '10'))
    temp_dir = os.getenv('TEMP_DIR', '/tmp/recipe-bot')

    # Create server
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=max_workers))

    # Add servicer
    servicer = ScraperServicer(output_dir=temp_dir)

    # Wrap async servicer methods
    class SyncScraperServicer(scraper_pb2_grpc.ScraperServiceServicer):
        def ScrapeContent(self, request, context):
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
            try:
                return loop.run_until_complete(
                    servicer.ScrapeContent(request, context)
                )
            finally:
                loop.close()

    scraper_pb2_grpc.add_ScraperServiceServicer_to_server(
        SyncScraperServicer(),
        server
    )

    # Start server
    server.add_insecure_port(f'[::]:{port}')
    server.start()

    logger.info(f"gRPC server started on port {port}")
    logger.info(f"Temporary directory: {temp_dir}")
    logger.info(f"Transcription provider: {os.getenv('TRANSCRIPTION_PROVIDER', 'google-stt')}")

    try:
        server.wait_for_termination()
    except KeyboardInterrupt:
        logger.info("Shutting down server...")
        server.stop(grace=5)


if __name__ == '__main__':
    serve()
