# Browser Render Go

Go実装のブラウザ自動化サービス。Cloudflare Puppeteer Workerの機能をgRPC/HTTPサービスとして提供します。

## 🚀 特徴

- **gRPC & HTTP API**: 両方のプロトコルをサポート
- **ブラウザ自動化**: Rodを使用したChrome/Chromium操作
- **セッション管理**: SQLiteによる永続的なセッション・Cookie管理
- **Protocol Buffers**: 型安全な通信
- **Docker対応**: コンテナ化されたデプロイメント

## 📋 必要要件

- Go 1.21以上
- Chrome/Chromium
- SQLite
- Protocol Buffers コンパイラ (オプション)

## 🛠️ セットアップ

### 1. クローンとセットアップ

```bash
git clone https://github.com/yourusername/browser_render_go.git
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
# 車両データ取得
curl -X POST http://localhost:8080/v1/vehicle/data \
  -H "Content-Type: application/json" \
  -d '{
    "branch_id": "00000000",
    "filter_id": "0",
    "force_login": false
  }'

# ヘルスチェック
curl http://localhost:8080/health

# セッション確認
curl "http://localhost:8080/v1/session/check?session_id=xxx"
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