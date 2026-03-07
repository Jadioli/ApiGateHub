#!/bin/sh
set -e

if [ ! -f ".env" ]; then
  cp .env.example .env
fi

if [ ! -f "web/dist/index.html" ]; then
  echo "Building frontend..."
  (
    cd web
    npm ci
    npm run build
  )
fi

echo "Building backend..."
go build -o apihub ./cmd/server

echo "ApiHub listening on http://localhost:9011"
./apihub
