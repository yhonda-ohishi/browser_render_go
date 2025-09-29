#!/bin/bash
# Browser Render Go - ワンライナー実行スクリプト

echo "🚀 Browser Render Go - 簡単セットアップ"
echo "=================================="

# Docker Composeファイルをダウンロード
echo "📥 Docker Compose設定をダウンロード中..."
curl -s -O https://raw.githubusercontent.com/yhonda-ohishi/browser_render_go/master/docker-compose.standalone.yml

# .envファイルが存在しない場合は作成
if [ ! -f .env ]; then
    echo ""
    echo "⚙️ 環境設定"
    echo "認証情報を入力してください (スキップする場合はEnterキー):"

    read -p "USER_NAME: " user_name
    read -p "COMP_ID: " comp_id
    read -s -p "USER_PASS: " user_pass
    echo ""

    cat > .env << EOF
USER_NAME=${user_name:-your_username}
COMP_ID=${comp_id:-your_company_id}
USER_PASS=${user_pass:-your_password}
EOF
    echo "✅ .envファイルを作成しました"
fi

# Docker Composeで起動
echo ""
echo "🐳 Dockerコンテナを起動中..."
docker-compose -f docker-compose.standalone.yml up -d

# 起動確認
echo ""
echo "⏳ サービスの起動を待っています..."
sleep 5

# ヘルスチェック
if curl -s http://localhost:8080/health | grep -q "healthy"; then
    echo "✅ サーバーが正常に起動しました!"
    echo ""
    echo "📌 アクセス情報:"
    echo "   HTTP API: http://localhost:8080"
    echo "   Health: http://localhost:8080/health"
    echo "   Metrics: http://localhost:8080/metrics"
    echo ""
    echo "📝 使い方:"
    echo "   curl -X POST http://localhost:8080/v1/vehicle/data \\"
    echo "     -H 'Content-Type: application/json' \\"
    echo "     -d '{\"branch_id\":\"\",\"filter_id\":\"0\",\"force_login\":false}'"
else
    echo "⚠️ サーバーの起動に失敗した可能性があります"
    echo "ログを確認: docker-compose -f docker-compose.standalone.yml logs"
fi