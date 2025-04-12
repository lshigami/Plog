#!/bin/sh
set -e

# Fetch secrets from AWS Secrets Manager
echo "Fetching secrets from AWS Secrets Manager..."
aws secretsmanager get-secret-value --secret-id goPlog --region ap-southeast-1 --query 'SecretString' --output text | jq -r 'to_entries|map("\(.key)=\(.value|tostring)")|.[]' > .env

# Load environment variables from .env file
set -a
. ./.env
set +a

echo "Waiting for PostgreSQL to be ready..."

# Extract host and port from DATABASE_URL
DB_HOST=$(echo $DATABASE_URL | sed -n 's/.*@\([^:]*\).*/\1/p')
DB_PORT=$(echo $DATABASE_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')

# Wait for PostgreSQL to be ready
until pg_isready -h $DB_HOST -p $DB_PORT -U admin; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up - running migrations"

# Run migrations with proper error handling
if ! migrate -path /app/migrations -database "${DATABASE_URL}" -verbose up; then
  echo "Migration failed!"
  exit 1
fi

echo "Migrations completed successfully"

# Khởi động server
echo "Starting server..."
exec /app/server