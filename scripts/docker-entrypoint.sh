#!/bin/bash
set -e

# Default values
SMTP_PORT=${SMTP_PORT:-2525}
HTTP_PORT=${HTTP_PORT:-8080}
DB_PATH=${DB_PATH:-/app/data/emails.db}
LOG_PATH=${LOG_PATH:-/app/logs/mailcatch.log}
CLEAR_ON_SHUTDOWN=${CLEAR_ON_SHUTDOWN:-true}

# Create directories
mkdir -p "$(dirname "$DB_PATH")"
mkdir -p "$(dirname "$LOG_PATH")"

# Handle signals gracefully
trap 'echo "Received SIGTERM, shutting down gracefully..."; kill -TERM $PID; wait $PID' TERM

echo "Starting MailCatch..."
echo "SMTP Port: $SMTP_PORT"
echo "HTTP Port: $HTTP_PORT" 
echo "Database: $DB_PATH"
echo "Log file: $LOG_PATH"

# Start mailcatch in background
mailcatch \
  --smtp-port="$SMTP_PORT" \
  --http-port="$HTTP_PORT" \
  --db-path="$DB_PATH" \
  --log-path="$LOG_PATH" \
  --clear-on-shutdown="$CLEAR_ON_SHUTDOWN" &

PID=$!

# Wait for process to finish
wait $PID