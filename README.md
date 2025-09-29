# Browser Render Go

Go実装のブラウザ自動化サービス。Cloudflare Puppeteer Workerの機能をgRPC/HTTPサービスとして提供します。

## 🚀 クイックスタート（クローン不要）

### ワンライナー実行

```bash
# Docker環境があれば即座に起動可能
curl -sSL https://raw.githubusercontent.com/yhonda-ohishi/browser_render_go/master/run.sh | bash
```

### Docker Composeのみで実行

```bash
# docker-compose.ymlをダウンロードして起動（10分間隔自動実行スケジューラー付き）
curl -O https://raw.githubusercontent.com/yhonda-ohishi/browser_render_go/master/docker-compose.standalone.yml
docker-compose -f docker-compose.standalone.yml up -d

# ログ確認
docker-compose -f docker-compose.standalone.yml logs -f

# スケジューラーログのみ確認
docker-compose -f docker-compose.standalone.yml logs scheduler
```

## 🚀 特徴

- **gRPC & HTTP API**: 両方のプロトコルをサポート
- **ブラウザ自動化**: Rodを使用したChrome/Chromium操作
- **セッション管理**: SQLiteによる永続的なセッション・Cookie管理
- **Protocol Buffers**: 型安全な通信
- **Docker対応**: コンテナ化されたデプロイメント
- **自動スケジューラー**: 10分間隔でのVenusデータ自動取得
- **クローン不要**: GitHubから直接ビルド可能

## 📋 必要要件

- Docker & Docker Compose（クイックスタート用）
- Go 1.21以上（ローカル開発用）
- Chrome/Chromium
- SQLite

## 🛠️ セットアップ（開発者向け）

### 1. リポジトリクローン

```bash
git clone https://github.com/yhonda-ohishi/browser_render_go.git
cd browser_render_go
go mod download
```

### 2. 環境変数設定

```bash
# .env ファイルを作成
cat > .env << EOF
USER_NAME=your_username
COMP_ID=your_company_id
USER_PASS=your_password
GRPC_PORT=50051
HTTP_PORT=8080
BROWSER_HEADLESS=true
SQLITE_PATH=./data/browser_render.db
EOF
```

### 3. Protocol Buffers生成（オプション）

```bash
# Windows
scripts\generate-proto-go.bat

# Linux/Mac
./scripts/generate-proto-go.sh
```

## 🏃‍♂️ 実行方法

### ローカル実行

```bash
# ビルド
go build -o browser_render ./src

# 実行
./browser_render

# オプション付き実行
./browser_render --http-port=8080 --grpc-port=50051 --headless=true
```

### Docker実行

```bash
# ビルド
docker-compose build

# 実行
docker-compose up -d

# ログ確認
docker-compose logs -f
```

## 📡 API使用例

### HTTP API

```bash
# 車両データ取得（手動実行）
curl http://localhost:8080/v1/vehicle/data

# ジョブ状態確認
curl http://localhost:8080/v1/job/{job-id}

# 全ジョブ一覧
curl http://localhost:8080/v1/jobs

# ヘルスチェック
curl http://localhost:8080/health

# セッション確認
curl "http://localhost:8080/v1/session/check?session_id=xxx"
```

### 自動スケジューラー機能

Docker Compose実行時に、10分間隔でVenusシステムから自動的に車両データを取得し、Hono APIに送信します。

```bash
# スケジューラー設定確認
docker-compose logs scheduler

# スケジューラー間隔変更（環境変数）
export CRON_SCHEDULE="*/5 * * * *"  # 5分間隔に変更
docker-compose up -d
```

### gRPC API

```go
// Go クライアント例
import pb "github.com/yourusername/browser_render_go/gen/proto/browser_render/v1"

// 接続
conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
client := pb.NewBrowserRenderServiceClient(conn)

// データ取得
resp, _ := client.GetVehicleData(context.Background(), &pb.GetVehicleDataRequest{
    BranchId: "00000000",
    FilterId: "0",
})
```

## 📁 プロジェクト構造

```
browser_render_go/
├── src/
│   ├── main.go           # エントリーポイント
│   ├── server/
│   │   ├── grpc.go       # gRPCサーバー
│   │   └── http.go       # HTTPサーバー
│   ├── browser/
│   │   └── renderer.go   # ブラウザ操作
│   ├── storage/
│   │   └── sqlite.go     # データ永続化
│   └── config/
│       └── config.go     # 設定管理
├── proto/
│   └── browser_render.proto  # Protocol Buffers定義
├── tests/                # テストコード
├── scripts/              # ビルド・デプロイスクリプト
└── docker-compose.yml    # Docker設定
```

## 🧪 テスト

```bash
# 全テスト実行
go test ./...

# カバレッジ付きテスト
go test -cover ./...

# 特定のテスト実行
go test ./tests -run TestStorage
```

## 📊 メトリクス

- `/metrics` - Prometheusメトリクス
- `/health` - ヘルスチェック

## 🔧 設定オプション

| 環境変数 | 説明 | デフォルト |
|---------|------|------------|
| `GRPC_PORT` | gRPCサーバーポート | 50051 |
| `HTTP_PORT` | HTTPサーバーポート | 8080 |
| `BROWSER_HEADLESS` | ヘッドレスモード | true |
| `BROWSER_TIMEOUT` | タイムアウト時間 | 30s |
| `SQLITE_PATH` | データベースパス | ./data/browser_render.db |
| `SESSION_TTL` | セッション有効期限 | 10m |
| `CRON_SCHEDULE` | スケジューラー実行間隔 | */10 * * * * |

## 🚀 デプロイメント

### Kubernetes

```yaml
kubectl apply -f deployments/k8s/
```

### Systemd

```bash
sudo cp deployments/systemd/browser-render.service /etc/systemd/system/
sudo systemctl enable browser-render
sudo systemctl start browser-render
```

## 📝 開発

### コード生成

```bash
# Protocol Buffers
buf generate

# モック生成
go generate ./...
```

### リント

```bash
golangci-lint run
```

## 🤝 コントリビューション

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 🔗 関連リンク

- [仕様書](SPEC.md)
- [実行管理](EXECUTION.md)
- [計画書](plan.md)

## ⚠️ 注意事項

- 本番環境では必ず環境変数で認証情報を管理してください
- ブラウザインスタンスは適切にクローズしてください
- セッションは定期的にクリーンアップしてください