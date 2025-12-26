#!/bin/bash
set -e

cd "$(dirname "$0")/.."

if [ ! -f .env ]; then
    echo "Creating .env from .env.example..."
    cp .env.example .env
    echo "Please configure .env with your credentials"
    exit 1
fi

export $(grep -v '^#' .env | xargs)

echo "Generating proto files..."
./scripts/gen-proto.sh

echo "Starting services with Docker Compose..."
docker compose -f deployments/compose.yaml up --build -d

echo "Waiting for services to start..."
sleep 5

echo "Checking service health..."
curl -s http://localhost:8080/health || echo "Gateway not ready yet"

echo ""
echo "Services started:"
echo "  Gateway: http://localhost:8080"
echo "  ASR:     localhost:50051"
echo "  Translator: localhost:50052"
echo "  TTS:     localhost:50053"
echo ""
echo "View logs: docker compose -f deployments/compose.yaml logs -f"
