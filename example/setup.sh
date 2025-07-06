#!/bin/bash

# Setup script for dbx example
# This script helps set up the example database and environment

set -e

echo "🚀 Setting up dbx example..."

# Check if PostgreSQL is running
if ! pg_isready -q; then
    echo "❌ PostgreSQL is not running. Please start PostgreSQL first."
    echo "   On macOS: brew services start postgresql"
    echo "   On Ubuntu: sudo systemctl start postgresql"
    echo "   On Windows: Start PostgreSQL service"
    exit 1
fi

echo "✅ PostgreSQL is running"

# Create database if it doesn't exist
echo "📦 Creating database 'dbx_example'..."
createdb dbx_example 2>/dev/null || echo "Database already exists"

# Run schema
echo "🗄️  Setting up schema..."
psql -d dbx_example -f schema.sql

echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Update the connection string in example/main.go"
echo "2. Run: cd example && go run main.go"
echo ""
echo "Connection string format:"
echo "postgres://username:password@localhost:5432/dbx_example?sslmode=disable" 