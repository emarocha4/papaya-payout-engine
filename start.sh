#!/bin/bash

set -e

echo "=== Papaya Payout Engine - Quick Start ==="
echo ""

echo "Step 1: Building application..."
go build -o papaya-payout-engine main.go
echo "✓ Build complete"
echo ""

echo "Step 2: Checking PostgreSQL..."
if ! docker ps | grep -q postgres; then
    echo "PostgreSQL not running. Please start PostgreSQL first:"
    echo "  docker compose up -d"
    exit 1
fi
echo "✓ PostgreSQL is running"
echo ""

echo "Step 3: Creating database (if not exists)..."
docker exec -i $(docker ps | grep postgres | awk '{print $1}') psql -U postgres -c "CREATE DATABASE papaya_payout_engine;" 2>/dev/null || echo "Database already exists"
echo ""

echo "Step 4: Running migrations..."
CONTAINER_ID=$(docker ps | grep postgres | awk '{print $1}')
docker exec -i $CONTAINER_ID psql -U postgres -d papaya_payout_engine < migration/000001_create_merchants.up.sql 2>/dev/null || echo "Merchants table already exists"
docker exec -i $CONTAINER_ID psql -U postgres -d papaya_payout_engine < migration/000002_create_decisions.up.sql 2>/dev/null || echo "Decisions table already exists"
docker exec -i $CONTAINER_ID psql -U postgres -d papaya_payout_engine < migration/000003_create_batches.up.sql 2>/dev/null || echo "Batches table already exists"
echo "✓ Migrations complete"
echo ""

echo "Step 5: Starting server on port 8081..."
echo "To stop: kill the process or press Ctrl+C"
echo ""

DB_USER=postgres DB_PASSWORD=postgres PORT=8081 ./papaya-payout-engine
