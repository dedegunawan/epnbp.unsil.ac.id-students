#!/bin/bash
echo "Stopping production environment..."
docker-compose -f docker-compose.yml down
