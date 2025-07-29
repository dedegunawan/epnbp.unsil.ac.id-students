#!/bin/bash

containers=("ujian-frontend-production" "ujian-backend-production")

for container in "${containers[@]}"; do
  ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$container")
  if grep -q "$container" /etc/hosts; then
    sudo sed -i.bak "/$container/d" /etc/hosts
  fi
  echo "$ip $container" | sudo tee -a /etc/hosts
done
