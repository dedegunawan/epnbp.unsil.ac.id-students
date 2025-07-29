#!/bin/bash
echo "Starting staging environment..."
docker-compose -f docker-compose.staging.yml up -d --build
