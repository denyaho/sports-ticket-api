# 42Tokyo Road to DeNA Server

Go言語で実装されたAPIサーバーです。
DockerコンテナかLocalか慣れてる方選んでください。

## 技術スタック

- **言語**: Go 1.24.5
- **Webフレームワーク**: net/http (標準ライブラリ)
- **データベース**: PostgreSQL 15
- **ORM/クエリビルダー**: sqlx
- **コンテナ**: Docker & Docker Compose

## セットアップ

### 前提条件

- Go 1.24以上
- Docker & Docker Compose

42 Tokyoの校舎PCでGo 1.24以上を使用する場合、以下のスクリプトを実行してください。
```
# Go 1.24.5をダウンロード
wget https://go.dev/dl/go1.24.5.linux-amd64.tar.gz

# ホームディレクトリに解凍
tar -C $HOME -xzf go1.24.5.linux-amd64.tar.gz

# アーカイブファイルを削除
rm go1.24.5.linux-amd64.tar.gz

# ~/.zshrcに以下の行を追加(既存のGo設定があれば置き換える):
GOPATH=$HOME/go-workspace
PATH=$HOME/go/bin:$GOPATH/bin:$PATH

# シェル設定をリロード
source ~/.zshrc

# インストールを確認
go version
```

### 環境変数

`.env.example`を参考に`.env`ファイルを作成してください。


### 初回セットアップ

```bash
# リポジトリのクローン
git clone <repository-url>
cd 42tokyo-road-to-dena-server-base

# 開発環境のセットアップ
go mod download
go mod tidy
```

## 開発

### サーバーの起動

```bash
docker compose up -d
```

### APIドキュメント（Swagger UI）

- OpenAPI定義: `docs/openapi.yaml`
- サーバー起動後に `http://localhost:8080/swagger/` を開くとSwagger UIが表示されます
