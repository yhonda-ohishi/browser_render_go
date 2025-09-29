#!/bin/bash

echo "Testing different branch_id and filter_id parameters"
echo "====================================================="
echo ""

# Test 1: Default parameters (what we've been using)
echo "Test 1: Default (branch_id: 00000000, filter_id: 0)"
echo "----------------------------------------------------"
curl -X POST http://localhost:8080/v1/vehicle/data \
  -H "Content-Type: application/json" \
  -d '{
    "branch_id": "00000000",
    "filter_id": "0",
    "force_login": false
  }' | jq '.data | length'
echo ""

sleep 2

# Test 2: Empty branch_id
echo "Test 2: Empty branch_id"
echo "-----------------------"
curl -X POST http://localhost:8080/v1/vehicle/data \
  -H "Content-Type: application/json" \
  -d '{
    "branch_id": "",
    "filter_id": "0",
    "force_login": false
  }' | jq '.data | length'
echo ""

sleep 2

# Test 3: Different filter_id
echo "Test 3: filter_id: 1"
echo "--------------------"
curl -X POST http://localhost:8080/v1/vehicle/data \
  -H "Content-Type: application/json" \
  -d '{
    "branch_id": "00000000",
    "filter_id": "1",
    "force_login": false
  }' | jq '.data | length'
echo ""

sleep 2

# Test 4: filter_id: 2
echo "Test 4: filter_id: 2"
echo "--------------------"
curl -X POST http://localhost:8080/v1/vehicle/data \
  -H "Content-Type: application/json" \
  -d '{
    "branch_id": "00000000",
    "filter_id": "2",
    "force_login": false
  }' | jq '.data | length'
echo ""

echo "Tests completed!"