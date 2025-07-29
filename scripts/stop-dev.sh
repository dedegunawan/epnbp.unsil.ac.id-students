#!/bin/bash
echo "Stopping development environment..."
docker compose -f docker-compose.development.yml down
