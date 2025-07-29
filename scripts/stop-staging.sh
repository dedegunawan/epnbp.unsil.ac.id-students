#!/bin/bash
echo "Stopping staging environment..."
docker-compose -f docker-compose.staging.yml down
