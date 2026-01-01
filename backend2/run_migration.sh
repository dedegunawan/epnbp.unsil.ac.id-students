#!/bin/bash

# ================================================
# RUN MIGRATION SCRIPT (MYSQL)
# Membaca dari .env dan merakit DB_URL otomatis
# Mendukung: up, down, force <version>, version, drop
# ================================================

# Load .env jika ada
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

# Pastikan variabel utama tersedia
: "${DB_HOST:?Harap set DB_HOST di .env}"
: "${DB_PORT:?Harap set DB_PORT di .env}"
: "${DB_USER:?Harap set DB_USER di .env}"
: "${DB_NAME:?Harap set DB_NAME di .env}"

# DB_PASS bisa kosong, maka jangan langsung pakai di URL
if [ -z "$DB_PASS" ]; then
  AUTH="$DB_USER"
else
  AUTH="$DB_USER:$DB_PASS"
fi

# DB_PARAMS opsional
if [ -z "$DB_PARAMS" ]; then
  DB_PARAMS="charset=utf8mb4&parseTime=True&loc=Local"
fi

# Rakitan URL MySQL untuk golang-migrate
DB_URL="mysql://${AUTH}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?${DB_PARAMS}"

# Path migrasi
DIR="db/migrations"

# Ambil perintah dari argumen (default: up)
COMMAND=${1:-up}

# Jalankan migrasi
case $COMMAND in
  up|down|version|drop)
    echo "🚀 Menjalankan migrate $COMMAND..."
    migrate -path "$DIR" -database "$DB_URL" "$COMMAND"
    ;;
  force)
    VERSION=$2
    if [ -z "$VERSION" ]; then
      echo "❌ Perintah 'force' membutuhkan versi. Contoh:"
      echo "./run_migration.sh force 2"
      exit 1
    fi
    echo "🚀 Menjalankan migrate force $VERSION..."
    migrate -path "$DIR" -database "$DB_URL" force "$VERSION"
    ;;
  *)
    echo "❌ Perintah tidak valid: $COMMAND"
    echo "Gunakan salah satu: up, down, version, force <version>, drop"
    exit 1
    ;;
esac

# Status
if [ $? -eq 0 ]; then
  echo "✅ Migrasi berhasil"
else
  echo "⚠️ Migrasi gagal"
fi
