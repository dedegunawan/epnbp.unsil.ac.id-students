#!/bin/bash
echo "Restarting staging environment..."
docker compose -f docker-compose.staging.yml down
docker compose -f docker-compose.staging.yml up -d --build
