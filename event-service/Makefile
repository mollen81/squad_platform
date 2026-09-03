.PHONY: proto build run clean

proto:
	@cd internal/core/proto && protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		event.proto
	@echo "Protobuf code generated successfully"

build:
	@go build -o bin/event-service cmd/main.go
	@echo "Build completed"

run: build
	@./bin/event-service
	@echo "Running event-service"

clean:
	@rm -rf bin/
	@echo "Clean completed"
