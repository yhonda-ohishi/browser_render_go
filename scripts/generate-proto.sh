#!/bin/bash

# Protocol Buffers generation script

echo "Generating Protocol Buffers code..."

# Check if buf is installed
if ! command -v buf &> /dev/null; then
    echo "buf is not installed. Please run ./scripts/install-buf.sh first"
    exit 1
fi

# Install required Go plugins if not already installed
echo "Installing/updating Go protoc plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Generate code
echo "Running buf generate..."
buf generate

echo "Protocol Buffers generation completed!"
echo "Generated files are in the gen/ directory"