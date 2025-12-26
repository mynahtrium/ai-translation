#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "Stopping services..."
docker compose -f deployments/compose.yaml down

echo "Services stopped"
