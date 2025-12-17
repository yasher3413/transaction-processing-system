#!/bin/bash
# Generate go.sum using Docker (if Go is not installed locally)

echo "Generating go.sum using Docker..."

docker run --rm \
  -v "$(pwd):/workspace" \
  -w /workspace \
  golang:1.22-alpine \
  sh -c "go mod tidy && chown -R $(id -u):$(id -g) go.sum"

echo "go.sum generated successfully!"


