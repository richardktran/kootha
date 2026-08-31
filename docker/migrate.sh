#!/bin/sh
set -e

HOST="${POSTGRES_HOST:-postgres}"
PORT="${POSTGRES_PORT:-5432}"
USER="${POSTGRES_USER:?POSTGRES_USER is required}"
PASS="${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"

migrate -path=/migrations/quiz-session \
  -database "postgres://${USER}:${PASS}@${HOST}:${PORT}/quiz_session?sslmode=disable" up

migrate -path=/migrations/quiz-bank \
  -database "postgres://${USER}:${PASS}@${HOST}:${PORT}/quiz_bank?sslmode=disable" up

echo "migrations complete"
