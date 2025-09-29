# リモートサーバー展開手順 (133.18.115.234)

## 1. サーバー接続・セットアップ

```bash
# SSH接続
ssh user@133.18.115.234

# 必要なツールインストール
sudo apt update
sudo apt install -y git docker.io docker-compose curl

# Dockerサービス開始・自動起動設定
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER

# 新しいグループ設定を反映（再ログインまたは）
newgrp docker
```

## 2. リポジトリクローン・設定

```bash
# リポジトリクローン
git clone https://github.com/yhonda-ohishi/browser_render_go.git
cd browser_render_go

# 環境変数設定ファイル作成
cp .env.example .env

# 環境変数編集
nano .env
```

### .env設定内容:
```bash
# Venus システム認証情報
USER_NAME=your_actual_username
COMP_ID=your_actual_company_id
USER_PASS=your_actual_password

# サーバー設定
GRPC_PORT=50051
HTTP_PORT=8080
BROWSER_HEADLESS=true
BROWSER_DEBUG=false

# データベース
SQLITE_PATH=/app/data/browser_render.db

# セッション設定
SESSION_TTL=600000
COOKIE_TTL=86400

# スケジューラー設定（10分おき自動実行）
CRON_SCHEDULE=*/10 * * * *
API_URL=http://browser-render:8080
```

## 3. ディレクトリ作成・権限設定

```bash
# 必要なディレクトリ作成
mkdir -p data logs scripts

# 権限設定
chmod 755 data logs scripts
chmod +x scripts/call_api.sh
```

## 4. Docker Compose起動

```bash
# 全サービス起動
docker compose up -d

# 起動確認
docker compose ps
docker compose logs -f
```

## 5. 動作確認

```bash
# API健康状態チェック
curl http://localhost:8080/health
curl http://133.18.115.234:8080/health

# 手動ジョブ実行テスト
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d '{"action":"GetVehicleData"}'

# スケジューラーログ確認
docker compose logs scheduler

# データ保存確認
ls -la data/
ls -la data/vehicles_*.json
```

## 6. ファイアウォール設定（必要に応じて）

```bash
# ポート8080を開放
sudo ufw allow 8080/tcp
sudo ufw allow 50051/tcp  # gRPC用

# 状態確認
sudo ufw status
```

## 7. 自動起動設定

```bash
# systemdサービス作成
sudo nano /etc/systemd/system/browser-render.service
```

### サービスファイル内容:
```ini
[Unit]
Description=Browser Render Docker Compose
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/home/user/browser_render_go
ExecStart=/usr/bin/docker compose up -d
ExecStop=/usr/bin/docker compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

```bash
# サービス有効化・起動
sudo systemctl enable browser-render.service
sudo systemctl start browser-render.service
sudo systemctl status browser-render.service
```

## 8. ログ監視・管理

```bash
# リアルタイムログ監視
docker compose logs -f

# 特定サービスログ
docker compose logs scheduler
docker compose logs browser-render

# ログローテーション設定
sudo nano /etc/logrotate.d/browser-render
```

### ログローテーション設定:
```
/home/user/browser_render_go/logs/*.log {
    daily
    missingok
    rotate 7
    compress
    notifempty
    create 644 user user
}
```

## 9. アップデート手順

```bash
# 最新版取得
cd browser_render_go
git pull origin master

# サービス再起動
docker compose down
docker compose up -d --build

# 動作確認
curl http://localhost:8080/health
docker compose logs -f
```

## 10. トラブルシューティング

### ポート競合確認
```bash
sudo netstat -tulpn | grep :8080
sudo netstat -tulpn | grep :50051
```

### Docker状態確認
```bash
docker compose ps
docker network ls
docker volume ls
```

### システムリソース確認
```bash
free -h
df -h
docker system df
```

### コンテナ再起動
```bash
docker compose restart browser-render
docker compose restart scheduler
```

## 11. 外部アクセス設定

サーバーIP: `133.18.115.234`

- HTTP API: `http://133.18.115.234:8080`
- gRPC: `133.18.115.234:50051`
- Health Check: `http://133.18.115.234:8080/health`

自動スケジュール実行により、10分おきにVenusシステムから車両データを取得してHono APIに送信します。