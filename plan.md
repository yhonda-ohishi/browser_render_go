# browser_render_go


## 計画
- index.ts　はcloudflare/puppeteer　を使ったworkerのcodeである
- index.ts の機能をgoにて実装する

## 要件

- grpc
- HTTP ハンドラー
- Protocol Buffers
- buf でのバージョン管理
- gRPC/Protocol Buffers のバージョン管理
- gRPC-Web クライアント（Worker用）実装
- gRPC-Web クライアントのnpm での公開
- 将来的にbufconn で結合することも予見
- worker.jsは呼び出し側


## 不明点

### 1. 現在のindex.ts実装の詳細
- Cloudflare WorkerでPuppeteerを使用したWebスクレイピング処理
- VenusBridgeServiceという外部サービスと連携
- 不明点：このVenusBridgeServiceの仕様とAPIの詳細　　
　> website側のjavascript

### 2. Go実装への移行方針
- **ブラウザ自動化**: Go版でPuppeteerの代替（Chromedp、Rod等）の選定基準
- **Worker環境**: Cloudflare Worker相当の実行環境をGoでどう実現するか 
> linux,windowsにて実行
- **KVストレージ**: Cloudflare KVの代替をどう実装するか
> sqlite

### 3. gRPC関連の実装詳細
- **.protoファイルが未作成**: どのようなメッセージ定義が必要か
> 提案して
- **サービス定義**: どのようなRPCメソッドを提供するか
> getだけでいい　getでアクセスしたら、worker.jsに定義されてるaccess　pointにpushする
- **buf設定**: buf.yaml、buf.gen.yamlの設定内容
> 提案して
- **バージョン管理**: gRPC/Protocol Buffersのバージョニング戦略
> 提案して

### 4. gRPC-Webクライアント
- **Worker用実装**: Cloudflare Worker内でgRPC-Webクライアントをどう実装するか
    　現状必要ない
- **npmパッケージ化**: 公開パッケージの構成と依存関係
　提案して
- **TypeScript定義**: 型定義ファイルの生成方法
　buf gen

### 5. アーキテクチャ設計
- **HTTPハンドラー**: gRPCとHTTPの両方をサポートする方法
 つくって
- **プロセス管理（proc）**: 具体的な用途と実装方法
　つくって
- **bufconn統合**: 将来的なbufconnでのサービス間通信の設計
　今必要？

### 6. セキュリティ・認証
- 現在の実装では環境変数から認証情報を取得
- Go版での認証情報の管理方法　ー　環境変数
- セッション管理の実装方法　ー　必要？　httpでどう?

### 7. プロジェクト構造
- Goプロジェクトのディレクトリ構造 - src, testsに分けて入れて
- モジュール分割の方針　ー　今回のモジュールは一つ
- テストコードの配置　　ー　上記