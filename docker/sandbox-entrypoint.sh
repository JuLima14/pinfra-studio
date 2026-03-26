#!/bin/sh
set -e

# If no package.json, scaffold Next.js app inside container
if [ ! -f "package.json" ]; then
  echo "[sandbox] No project found, scaffolding Next.js..."
  npx --yes create-next-app@latest . \
    --yes --typescript --tailwind --eslint --app \
    --src-dir --import-alias "@/*" --use-npm --no-turbopack 2>&1
  echo "[sandbox] Scaffold complete"

  # Write CLAUDE.md for Claude CLI
  cat > CLAUDE.md << 'CLAUDEEOF'
# Project Instructions

You are building a Next.js application with the App Router.

## Stack
- Next.js 14+ with App Router (src/ directory)
- TypeScript (strict mode)
- Tailwind CSS for styling
- npm as package manager

## Rules
- Always use TypeScript
- Use the App Router (src/app/) not Pages Router
- Use Tailwind for styling, never inline styles or CSS modules
- Use server components by default, 'use client' only when needed
- Keep components small and focused
CLAUDEEOF
  echo "[sandbox] CLAUDE.md written"
elif [ ! -d "node_modules" ] || [ "package.json" -nt "node_modules/.package-lock.json" ]; then
  echo "[sandbox] Installing dependencies..."
  npm install 2>&1
fi

echo "[sandbox] Starting dev server..."
exec npm run dev -- --hostname 0.0.0.0
