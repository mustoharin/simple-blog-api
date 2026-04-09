#!/bin/sh
set -e

echo "Waiting for postgres..."
until PGPASSWORD=$POSTGRES_PASSWORD psql -h "postgres" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c '\q' 2>/dev/null; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up - checking for migration tool"

# Check if we can use migrate from the binary
if command -v migrate >/dev/null 2>&1; then
    echo "Running migrations..."
    migrate -path=/root/migrations -database="$DATABASE_URL" -verbose up
else
    echo "Migration tool not found, skipping migrations..."
fi

echo "Starting application..."
exec ./main
