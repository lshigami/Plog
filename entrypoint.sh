#!/bin/sh
set -e

# Try to fetch secrets from AWS Secrets Manager if credentials are available
if aws sts get-caller-identity >/dev/null 2>&1; then
  echo "AWS credentials found, fetching secrets from AWS Secrets Manager..."
  if aws secretsmanager get-secret-value --secret-id goPlog --region ap-southeast-1 --query 'SecretString' --output text | jq -r 'to_entries|map("\(.key)=\(.value|tostring)")|.[]' > .env 2>/dev/null; then
    echo "Successfully fetched secrets from AWS Secrets Manager"
    # Load environment variables from .env file
    set -a
    . ./.env
    set +a
  else
    echo "Failed to fetch secrets from AWS Secrets Manager. Using environment variables provided directly."
  fi
else
  echo "No AWS credentials found. Using environment variables provided directly."
fi

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