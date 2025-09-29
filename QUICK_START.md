# 🚀 Browser Render Go - Quick Start Guide

## 📦 簡単セットアップ (3ステップ)

### 1️⃣ リポジトリをクローン
```bash
git clone https://github.com/yhonda-ohishi/browser_render_go.git
cd browser_render_go
```

### 2️⃣ 環境変数を設定
```bash
cp .env.example .env
# .envファイルを編集して認証情報を設定
```

### 3️⃣ Docker Composeで起動
```bash
docker-compose up -d
```

## ✅ 動作確認

```bash
# ヘルスチェック
curl http://localhost:8080/health

# 車両データ取得（ジョブ作成）
curl http://localhost:8080/v1/vehicle/data

# ジョブ一覧確認
curl http://localhost:8080/v1/jobs
```

## 🛠️ 使用可能なコマンド

### Docker Compose (推奨)
```bash
# 起動
docker-compose up -d

# ログ確認
docker-compose logs -f

# スケジューラーログ確認
docker-compose logs scheduler

# 停止
docker-compose down

# 再起動
docker-compose restart
```

### Makefile を使う場合
```bash
# ヘルプを表示
make help

# Dockerで起動
make docker-run

# 停止
make docker-stop

# テスト実行
make test
```

### 直接実行する場合
```bash
# ビルド
go build -o browser_render ./src

# 実行
./browser_render --server=http
```

## 🔧 設定オプション

### 環境変数 (.env)
```env
# 認証情報 (必須)
USER_NAME=your_username
COMP_ID=your_company_id
USER_PASS=your_password

# サーバー設定
HTTP_PORT=8080
GRPC_PORT=50051

# ブラウザ設定
BROWSER_HEADLESS=true
BROWSER_TIMEOUT=60s
BROWSER_DEBUG=false

# スケジューラー設定
CRON_SCHEDULE=*/10 * * * *  # 10分間隔（変更可能）
```

### Docker Composeプロファイル

#### 開発環境
```bash
docker-compose up -d
```

#### 本番環境
```bash
docker-compose -f docker-compose.prod.yml up -d
```

#### デバッグモード (Chrome DevTools付き)
```bash
docker-compose --profile debug up -d
# Chrome DevTools: http://localhost:9222
```

## 📊 メトリクス & モニタリング

```bash
# メトリクス確認
curl http://localhost:8080/metrics

# コンテナ状態確認
docker-compose ps

# リソース使用状況
docker stats browser-render-server
```

## 🐛 トラブルシューティング

### ポート競合エラー
```bash
# 使用中のポートを確認
netstat -an | grep 8080

# docker-compose.ymlでポートを変更
ports:
  - "8081:8080"  # 8081に変更
```

### メモリ不足エラー
```bash
# Docker Desktopのメモリ制限を増やす
# または、ヘッドレスモードを確実に有効化
BROWSER_HEADLESS=true docker-compose up -d
```

### ログの確認
```bash
# 全ログ
docker-compose logs

# リアルタイムログ
docker-compose logs -f browser-render

# 最新100行
docker-compose logs --tail=100
```

## 🌐 リモートサーバーでの実行

### SSH経由でデプロイ
```bash
# サーバーにSSH接続
ssh user@your-server.com

# リポジトリをクローン
git clone https://github.com/yhonda-ohishi/browser_render_go.git
cd browser_render_go

# 環境変数設定
cp .env.example .env
nano .env  # 認証情報を設定

# Docker Composeで起動
docker-compose -f docker-compose.prod.yml up -d
```

### クラウドプロバイダー向け

#### AWS EC2 / Google Cloud VM
```bash
# Docker & Docker Composeをインストール
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# プロジェクトをデプロイ
git clone https://github.com/yhonda-ohishi/browser_render_go.git
cd browser_render_go
docker-compose up -d
```

#### Heroku
```bash
heroku create your-app-name
heroku stack:set container
git push heroku main
```

## 📚 詳細ドキュメント

- [API仕様](./SPEC.md)
- [テストカバレッジ](./TEST_COVERAGE.md)
- [実行詳細](./EXECUTION.md)

## 🆘 サポート

問題が発生した場合は、[GitHub Issues](https://github.com/yhonda-ohishi/browser_render_go/issues) で報告してください。