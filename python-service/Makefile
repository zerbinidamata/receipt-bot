.PHONY: generate clean run install

# Generate Python code from proto files
generate:
	@echo "Generating Python protobuf code..."
	@poetry run python -m grpc_tools.protoc -I../proto \
		--python_out=src \
		--grpc_python_out=src \
		../proto/scraper.proto
	# Fix imports in generated files for proper relative imports
	@if [ -f src/scraper_pb2_grpc.py ]; then \
		sed -i.bak 's/import scraper_pb2 as scraper__pb2/from . import scraper_pb2 as scraper__pb2/' src/scraper_pb2_grpc.py && \
		rm -f src/scraper_pb2_grpc.py.bak; \
	fi
	@echo "Python protobuf code generated successfully"

# Clean generated files
clean:
	rm -f src/*_pb2.py src/*_pb2_grpc.py
	@echo "Cleaned generated protobuf files"

# Install dependencies
install:
	poetry install

# Run the server
run: generate
	poetry run python run_server.py

# Run without regenerating protos
run-only:
	poetry run python run_server.py
