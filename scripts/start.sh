#!/bin/bash
# Start script for development

echo "Starting Accounting Web (Development Mode)"
echo "=========================================="

# Check if .env exists
if [ ! -f .env ]; then
    echo "Error: .env file not found"
    echo "Please copy .env.example to .env and configure it"
    exit 1
fi

# Create storage directories if not exist
mkdir -p storage/uploads
mkdir -p storage/exports
mkdir -p storage/logs

# Start web server in background
echo "Starting web server..."
go run cmd/web/main.go &
WEB_PID=$!

# Start worker in background
echo "Starting worker..."
go run cmd/worker/main.go &
WORKER_PID=$!

echo ""
echo "Services started:"
echo "  Web Server PID: $WEB_PID"
echo "  Worker PID: $WORKER_PID"
echo ""
echo "Web server running at: http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for interrupt
trap "kill $WEB_PID $WORKER_PID; exit" INT
wait
