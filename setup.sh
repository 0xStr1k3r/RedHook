#!/bin/bash

# RedHook Setup Script
# This script sets up the environment for running RedHook locally

echo "🔧 Setting up RedHook..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed. Please install Go 1.26+ first."
    exit 1
fi

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    echo "⚠️  Warning: PostgreSQL client not found. Make sure PostgreSQL is running."
fi

# Navigate to backend
cd "$(dirname "$0")/backend" || exit 1

# Copy environment file
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
fi

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod download

echo ""
echo "======================================"
echo "✅ Setup Complete!"
echo "======================================"
echo ""
echo "Next steps:"
echo "1. Configure .env with your email provider (Resend/AWS SES)"
echo "2. Install ngrok (if not installed):"
echo "   curl -sSL https://ngrok-agent.s3.amazonaws.com/ngrok.asc \\"
echo "     | sudo tee /etc/apt/trusted.gpg.d/ngrok.asc >/dev/null \\"
echo "   && echo \"deb https://ngrok-agent.s3.amazonaws.com bookworm main\" \\"
echo "     | sudo tee /etc/apt/sources.list.d/ngrok.list \\"
echo "   && sudo apt update \\"
echo "   && sudo apt install ngrok"
echo ""
echo "3. Start ngrok: ngrok http 8081"
echo "4. Run: ./start.sh"
echo ""
echo "📍 Dashboard: http://localhost:8080"
echo "🔑 Login: admin@redhook.local / changeme123"
echo "======================================"