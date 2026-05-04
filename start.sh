#!/bin/bash

# RedHook Start Script
# This script builds and starts the RedHook server

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR/backend" || exit 1

echo "🚀 Starting RedHook..."

# Check if .env exists
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "⚠️  Please configure .env file before running!"
    exit 1
fi

# Build the application
echo "🔨 Building RedHook..."
go build -o redhook ./src

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful!"
echo ""
echo "======================================"
echo "🚀 RedHook server is starting..."
echo "======================================"
echo ""
echo "📍 Dashboard URL: http://localhost:8080"
echo ""
echo "🔑 Login Credentials:"
echo "   Email:    admin@redhook.local"
echo "   Password: changeme123"
echo ""
echo "📧 Phishing URLs:"
echo "   - http://localhost:8080/          (Dashboard)"
echo "   - http://localhost:8081/          (Phishing landing)"
echo ""
echo "💡 Press Ctrl+C to stop"
echo ""
echo "======================================"

./redhook 2>&1 | tee /tmp/redhook.log