#!/bin/sh
set -e

# Install dependencies if needed
if [ ! -d "node_modules" ] || [ "package.json" -nt "node_modules/.package-lock.json" ]; then
  echo "[sandbox] Installing dependencies..."
  npm install --prefer-offline 2>/dev/null || npm install
fi

echo "[sandbox] Starting dev server..."
exec npm run dev -- --hostname 0.0.0.0
