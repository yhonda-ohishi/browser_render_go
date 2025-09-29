#!/bin/bash

echo "Testing Browser Render Go HTTP API"
echo "==================================="

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s http://localhost:8080/health | jq .
echo ""

# Test vehicle data endpoint
echo "2. Testing vehicle data endpoint..."
curl -X POST http://localhost:8080/v1/vehicle/data \
  -H "Content-Type: application/json" \
  -d '{
    "branch_id": "00000000",
    "filter_id": "0",
    "force_login": false
  }' | jq .

echo ""
echo "Test completed!"