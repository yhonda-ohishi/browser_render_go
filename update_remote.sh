#!/bin/bash

echo "Updating remote server with Docker Compose..."
echo "=========================================="

# SSHでリモートサーバーに接続してコマンド実行
ssh root@133.18.115.234 << 'EOF'
cd /root/browser_render_go

echo "1. Stopping current containers..."
docker-compose down

echo "2. Pulling latest code from GitHub..."
git pull origin master

echo "3. Rebuilding with new code..."
docker-compose build --no-cache

echo "4. Starting updated containers..."
docker-compose up -d

echo "5. Checking container status..."
docker-compose ps

echo "6. Viewing logs..."
docker-compose logs --tail=20

echo "Update complete!"
EOF