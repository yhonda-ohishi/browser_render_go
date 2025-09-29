# Docker Compose 展開手順

## 1. リモートサーバーでの展開

```bash
# リポジトリをクローンまたは更新
git clone https://github.com/yhonda-ohishi/browser_render_go.git
cd browser_render_go

# または既存の場合
git pull origin master

# 環境変数設定ファイルを作成
cp .env.example .env
```

## 2. 環境変数設定 (.env)

```bash
# Venus システム認証情報
USER_NAME=your_username
COMP_ID=your_company_id
USER_PASS=your_password

# ブラウザ設定
BROWSER_DEBUG=false
BROWSER_HEADLESS=true

# スケジューラー設定 (オプション)
CRON_SCHEDULE=*/10 * * * *  # 10分おき
API_URL=http://browser-render:8080
```

## 3. Docker Compose起動

```bash
# 全サービス起動
docker compose up -d

# ログ確認
docker compose logs -f

# 特定サービスのログ確認
docker compose logs -f scheduler
docker compose logs -f browser-render
```

## 4. 動作確認

```bash
# API健康状態チェック
curl http://localhost:8080/health

# 手動ジョブ実行テスト
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d '{"action":"GetVehicleData"}'

# スケジューラーログ確認（10分間隔実行）
docker compose logs scheduler
```

## 5. サービス管理

```bash
# 停止
docker compose down

# 再起動
docker compose restart

# 設定更新後の再構築
docker compose up -d --build

# ボリューム含む完全削除
docker compose down -v
```

## 6. トラブルシューティング

### スケジューラーが動作しない場合
```bash
# コンテナ状態確認
docker compose ps

# スケジューラーコンテナ内確認
docker compose exec scheduler sh
cat /etc/crontabs/root
ps aux | grep crond
```

### API接続エラーの場合
```bash
# ネットワーク確認
docker network ls
docker network inspect browser_render_go_browser-net

# コンテナ間通信テスト
docker compose exec scheduler ping browser-render
```

## 7. カスタマイズ

### スケジュール変更
```bash
# .env ファイル編集
CRON_SCHEDULE=*/5 * * * *  # 5分おき

# 再起動
docker compose restart scheduler
```

### API URL変更
```bash
# .env ファイル編集
API_URL=http://external-server:8080

# 再起動
docker compose restart scheduler
```