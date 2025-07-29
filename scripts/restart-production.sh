#!/bin/bash
echo "Restarting production environment..."
docker-compose -f docker-compose.yml down
docker-compose -f docker-compose.yml up -d --build
