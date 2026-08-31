#!/bin/sh
set -e

API_URL="${NEXT_PUBLIC_API_URL:-http://localhost:8080}"
WS_URL="${NEXT_PUBLIC_WS_URL:-ws://localhost:8086/ws}"

find . -type f \( -name '*.js' -o -name '*.html' -o -name '*.json' \) | while IFS= read -r file; do
  sed -i "s#__NEXT_PUBLIC_API_URL__#${API_URL}#g" "$file"
  sed -i "s#__NEXT_PUBLIC_WS_URL__#${WS_URL}#g" "$file"
done

exec node server.js
