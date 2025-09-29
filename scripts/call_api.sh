#!/bin/sh

# 10分おきにAPI実行するスクリプト
API_URL=${API_URL:-"http://browser-render:8080"}
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$TIMESTAMP] Starting API call to $API_URL/api/jobs"

# APIジョブ作成
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/api/jobs" \
  -H "Content-Type: application/json" \
  -d '{"action":"GetVehicleData"}')

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "201" ]; then
    echo "[$TIMESTAMP] API call successful (HTTP $HTTP_CODE)"
    echo "Response: $BODY"
else
    echo "[$TIMESTAMP] API call failed (HTTP $HTTP_CODE)"
    echo "Response: $BODY"
fi