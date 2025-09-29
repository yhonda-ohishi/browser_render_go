# Browser Render Go - 実行管理ドキュメント

## 1. 開発フェーズ

### Phase 1: 基盤構築（1-2週目）
- [x] プロジェクト計画策定
- [x] 仕様書作成
- [ ] プロジェクト初期化
- [ ] Protocol Buffers定義
- [ ] buf設定
- [ ] 基本的なgRPCサーバー実装

### Phase 2: コア機能実装（3-4週目）
- [ ] ブラウザ操作モジュール（Rod/Chromedp）
- [ ] SQLiteストレージ実装
- [ ] セッション管理
- [ ] Cookie管理

### Phase 3: API実装（5-6週目）
- [ ] gRPCサービス実装
- [ ] HTTP Gateway実装
- [ ] エラーハンドリング
- [ ] ログ実装

### Phase 4: クライアント開発（7週目）
- [ ] TypeScript型生成
- [ ] npmパッケージ作成
- [ ] Worker.js実装
- [ ] クライアントテスト

### Phase 5: テスト・最適化（8週目）
- [ ] 単体テスト
- [ ] 統合テスト
- [ ] パフォーマンス最適化
- [ ] ドキュメント整備

## 2. 実装タスクリスト

### 2.1 初期セットアップ
```bash
# プロジェクト作成
mkdir -p browser_render_go/{src,tests,proto,gen,data}
cd browser_render_go

# Go モジュール初期化
go mod init github.com/yourusername/browser_render_go

# 依存関係インストール
go get google.golang.org/grpc
go get google.golang.org/protobuf
go get github.com/grpc-ecosystem/grpc-gateway/v2
go get github.com/go-rod/rod
go get github.com/mattn/go-sqlite3

# buf インストール
brew install bufbuild/buf/buf  # macOS
# または
curl -sSL https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Linux-x86_64 -o /usr/local/bin/buf  # Linux
chmod +x /usr/local/bin/buf
```

### 2.2 Protocol Buffers セットアップ
```bash
# buf.yaml 作成
cat > buf.yaml << EOF
version: v1
breaking:
  use:
    - FILE
lint:
  use:
    - DEFAULT
EOF

# buf.gen.yaml 作成
cat > buf.gen.yaml << EOF
version: v1
plugins:
  - plugin: go
    out: gen/proto
    opt: paths=source_relative
  - plugin: go-grpc
    out: gen/proto
    opt: paths=source_relative
  - plugin: grpc-gateway
    out: gen/proto
    opt:
      - paths=source_relative
      - generate_unbound_methods=true
  - plugin: openapiv2
    out: gen/openapi
EOF

# Protocol Buffers 生成
buf generate
```

### 2.3 ディレクトリ構造作成
```bash
# ディレクトリ構造を作成
mkdir -p src/{server,browser,storage,config,middleware}
mkdir -p tests/{unit,integration,e2e}
mkdir -p proto/browser_render/v1
mkdir -p gen/{proto,ts,openapi}
mkdir -p scripts
mkdir -p deployments/{docker,k8s}
```

## 3. 実行コマンド集

### 3.1 開発環境
```bash
# 開発サーバー起動
go run src/main.go

# ホットリロード開発
air -c .air.toml

# テスト実行
go test ./...

# テストカバレッジ
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# リント
golangci-lint run

# フォーマット
go fmt ./...
```

### 3.2 ビルド
```bash
# Linux向けビルド
GOOS=linux GOARCH=amd64 go build -o bin/browser_render_linux ./src

# Windows向けビルド
GOOS=windows GOARCH=amd64 go build -o bin/browser_render.exe ./src

# macOS向けビルド
GOOS=darwin GOARCH=amd64 go build -o bin/browser_render_darwin ./src

# Dockerビルド
docker build -t browser-render:latest .
```

### 3.3 Protocol Buffers
```bash
# .proto ファイル更新時
buf generate

# lint チェック
buf lint

# breaking change チェック
buf breaking --against '.git#branch=main'
```

### 3.4 データベース
```bash
# SQLite初期化
sqlite3 data/browser_render.db < scripts/init.sql

# マイグレーション実行
migrate -path migrations -database sqlite3://data/browser_render.db up

# データベースバックアップ
sqlite3 data/browser_render.db ".backup data/backup.db"
```

## 4. 運用手順

### 4.1 起動手順
```bash
# 1. 環境変数設定
export USER_NAME="your_username"
export COMP_ID="your_company_id"
export USER_PASS="your_password"
export GRPC_PORT=50051
export HTTP_PORT=8080

# 2. データベース初期化
./scripts/init-db.sh

# 3. サーバー起動
./bin/browser_render

# 4. ヘルスチェック
curl http://localhost:8080/health
```

### 4.2 停止手順
```bash
# 1. Graceful shutdown (SIGTERM)
kill -TERM $(pidof browser_render)

# 2. セッションクリーンアップ
./scripts/cleanup-sessions.sh

# 3. ログローテーション
./scripts/rotate-logs.sh
```

### 4.3 Docker Compose
```yaml
# docker-compose.yml
version: '3.8'
services:
  browser-render:
    build: .
    ports:
      - "50051:50051"
      - "8080:8080"
    environment:
      - USER_NAME=${USER_NAME}
      - COMP_ID=${COMP_ID}
      - USER_PASS=${USER_PASS}
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
    restart: unless-stopped

  chrome:
    image: chromedp/headless-shell:latest
    ports:
      - "9222:9222"
    restart: unless-stopped
```

## 5. トラブルシューティング

### 5.1 よくある問題と解決方法

| 問題 | 原因 | 解決方法 |
|------|------|----------|
| ブラウザ起動失敗 | Chrome未インストール | Chrome/Chromiumインストール |
| セッションエラー | Cookie期限切れ | セッションクリア＆再ログイン |
| メモリリーク | ブラウザ未クローズ | タイムアウト設定＆リソース管理 |
| SQLiteロック | 同時書き込み | WALモード有効化 |

### 5.2 デバッグコマンド
```bash
# ログ確認
tail -f logs/browser_render.log

# プロセス確認
ps aux | grep browser_render

# ポート確認
netstat -tlnp | grep -E "50051|8080"

# Chrome プロセス確認
ps aux | grep chrome

# データベース状態確認
sqlite3 data/browser_render.db "SELECT * FROM sessions;"
```

## 6. モニタリング

### 6.1 Prometheus設定
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'browser-render'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### 6.2 主要メトリクス
```bash
# リクエスト数
browser_render_requests_total

# レスポンス時間
browser_render_request_duration_seconds

# エラー率
browser_render_errors_total

# ブラウザインスタンス数
browser_render_browser_instances

# SQLite接続数
browser_render_db_connections
```

## 7. CI/CD設定

### 7.1 GitHub Actions
```yaml
# .github/workflows/main.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go test ./...
      - run: go build -o bin/browser_render ./src

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - run: docker build -t browser-render:${{ github.sha }} .
      - run: docker push browser-render:${{ github.sha }}
```

## 8. パフォーマンスチューニング

### 8.1 最適化項目
```go
// ブラウザプール設定
browserPool := &BrowserPool{
    MaxInstances: 5,
    MinInstances: 1,
    IdleTimeout:  10 * time.Minute,
}

// SQLite最適化
db.Exec("PRAGMA journal_mode=WAL")
db.Exec("PRAGMA synchronous=NORMAL")
db.Exec("PRAGMA cache_size=10000")
db.Exec("PRAGMA temp_store=MEMORY")

// gRPC設定
grpcServer := grpc.NewServer(
    grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
    grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
    grpc.MaxConcurrentStreams(100),
)
```

## 9. セキュリティチェックリスト

- [ ] 環境変数による機密情報管理
- [ ] TLS証明書設定
- [ ] SQLインジェクション対策
- [ ] XSS対策
- [ ] CSRF対策
- [ ] Rate Limiting実装
- [ ] ログの機密情報マスキング
- [ ] 定期的な依存関係更新

## 10. リリースノート管理

### バージョニング規則
- Major: 破壊的変更
- Minor: 機能追加
- Patch: バグ修正

### リリースプロセス
1. feature ブランチで開発
2. PR作成＆レビュー
3. テスト実行
4. main ブランチにマージ
5. タグ作成（v1.0.0形式）
6. リリースノート作成
7. デプロイ

## 11. 保守・運用

### 11.1 定期タスク
- **日次**: ログローテーション、バックアップ
- **週次**: パフォーマンス分析、セキュリティスキャン
- **月次**: 依存関係更新、容量管理

### 11.2 監視項目
- CPU/メモリ使用率
- ディスク使用量
- エラー発生率
- レスポンス時間
- 同時接続数

### 11.3 バックアップ戦略
```bash
# 自動バックアップスクリプト
#!/bin/bash
BACKUP_DIR="/backup/browser_render"
DATE=$(date +%Y%m%d_%H%M%S)

# SQLiteバックアップ
sqlite3 data/browser_render.db ".backup ${BACKUP_DIR}/db_${DATE}.db"

# 設定ファイルバックアップ
tar -czf ${BACKUP_DIR}/config_${DATE}.tar.gz config/

# 古いバックアップ削除（30日以上）
find ${BACKUP_DIR} -type f -mtime +30 -delete
```