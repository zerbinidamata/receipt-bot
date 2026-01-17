.PHONY: generate clean

# Generate Go code from proto files
generate-go:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		scraper.proto
	@echo "Go protobuf code generated successfully"

# Generate Python code from proto files
generate-python:
	@cd ../python-service && \
	export PATH="$$HOME/.local/bin:$$PATH" && \
	poetry run python -m grpc_tools.protoc -I../proto \
		--python_out=src \
		--grpc_python_out=src \
		../proto/scraper.proto
	# Fix imports in generated files for proper relative imports
	@if [ -f ../python-service/src/scraper_pb2_grpc.py ]; then \
		sed -i.bak 's/import scraper_pb2 as scraper__pb2/from . import scraper_pb2 as scraper__pb2/' ../python-service/src/scraper_pb2_grpc.py && \
		rm -f ../python-service/src/scraper_pb2_grpc.py.bak; \
	fi
	@echo "Python protobuf code generated successfully"

# Generate both Go and Python code
generate: generate-go generate-python

# Clean generated files
clean:
	rm -f *.pb.go *_grpc.pb.go
	rm -f ../python-service/src/*_pb2.py ../python-service/src/*_pb2_grpc.py
	@echo "Cleaned generated protobuf files"
