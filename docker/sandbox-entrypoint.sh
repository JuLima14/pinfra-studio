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

You are building a Next.js application inside Pinfra Studio.
A live preview with hot reload is running — every file change you make is visible instantly.

## Stack
- Next.js with App Router (src/app/ directory)
- TypeScript (strict mode)
- Tailwind CSS for all styling
- npm as package manager

## Project Structure
- src/app/page.tsx — main page
- src/app/layout.tsx — root layout
- src/app/globals.css — global styles (Tailwind imports)
- src/components/ — reusable components (create if needed)

## Rules
- Always use TypeScript
- Use the App Router (src/app/) — NOT Pages Router
- Use Tailwind CSS utility classes for ALL styling
- Never use inline styles, CSS modules, or styled-components
- Use server components by default, add 'use client' only when needed (onClick, useState, etc.)
- Keep components small and focused — one component per file
- Do NOT run 'npm run dev' — the dev server is already running
- Do NOT run 'npm install' unless you need a new dependency
- After making changes, briefly tell the user what you changed
CLAUDEEOF
  echo "[sandbox] CLAUDE.md written"
elif [ ! -d "node_modules" ] || [ "package.json" -nt "node_modules/.package-lock.json" ]; then
  echo "[sandbox] Installing dependencies..."
  npm install 2>&1
fi

echo "[sandbox] Starting dev server..."
exec npm run dev -- --hostname 0.0.0.0
