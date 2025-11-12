#!/bin/bash
# Build script

echo "Building Accounting Web..."

# Clean previous builds
rm -f accounting-web accounting-worker

# Build web server
echo "Building web server..."
go build -o accounting-web ./cmd/web
if [ $? -eq 0 ]; then
    echo "✓ Web server built successfully"
else
    echo "✗ Failed to build web server"
    exit 1
fi

# Build worker
echo "Building worker..."
go build -o accounting-worker ./cmd/worker
if [ $? -eq 0 ]; then
    echo "✓ Worker built successfully"
else
    echo "✗ Failed to build worker"
    exit 1
fi

echo ""
echo "Build completed successfully!"
echo "Run './accounting-web' to start the web server"
echo "Run './accounting-worker' to start the worker"
