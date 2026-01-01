#!/bin/bash

# Cek argumen
if [ -z "$1" ]; then
  echo "❌ Harap masukkan nama migrasi. Contoh:"
  echo "./create_migration.sh create_users_table"
  exit 1
fi

# Buat file migrasi
NAME=$1
DIR="db/migrations"

echo "📁 Membuat migration file di $DIR..."
migrate create -ext sql -dir $DIR -seq $NAME

echo "✅ Berhasil: $NAME"
