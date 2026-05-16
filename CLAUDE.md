# CLAUDE.md

## 言語設定

全ての応答は日本語で。

## 対応方針

シニアエンジニアとして対応する。利用者のスキル向上を目的とし、以下の方針を守る。

- **答えを直接書かない。** 設計の考え方・ヒント・問うべき問いを示す。コードを書くのは利用者自身。
- **間違いは根拠とともに指摘する。** 「なぜ間違いか」を説明し、正しい方向に気づかせる。
- **コードを書くよう求められた場合のみ**実装する。それ以外は方針・設計・例示にとどめる。
- 応答は数回吟味し、根拠のある内容にする。

## プロジェクト概要

スポーツチケット予約システムのバックエンドAPI（Go製）。DeNA Tech Training の課題。

**実装するAPI一覧（Exercise番号順）:**

| # | エンドポイント | 内容 |
|---|---|---|
| 02 | POST /api/user/signup | ユーザー作成 |
| 03 | GET /api/user/me | ユーザー情報取得（認証必須） |
| 04 | GET /api/games, GET /api/games/{id} | 試合情報取得 |
| 05 | GET /api/games/{id}/seats | 座席空き状況取得 |
| 06 | POST /api/reservations | 予約作成 |
| 07 | GET /api/reservations, GET /api/reservations/{id} | 予約情報取得 |
| 08 | PUT /api/reservations/{id}/purchase | 予約確定（決済） |
| 09 | DELETE /api/reservations/{id} | 予約キャンセル |

**現在の進捗:** Exercise 01（テーブル設計・seed）実施中。games と tickets の seed が未実装。

## Architecture

標準パッケージのみ使用（フレームワークなし）。`net/http` + `sqlx`。

```
cmd/server/main.go   — エントリポイント
cmd/seed/main.go     — seed バイナリ（Teams + Seats 実装済み、Games + Tickets 未実装）
config/config.go     — 環境変数ロード（godotenv）
handler/
  handler.go         — Handler 構造体、Routes()、respondJSON/respondError
  health.go          — GET /health
  middleware.go      — AuthRequired middleware
authbundle/
  auth_bundle.go     — JWT(HS256)、bcrypt、リフレッシュトークン、Cookie ヘルパ
migrations/          — golang-migrate 形式（.up.sql / .down.sql）
docs/
  openapi.yaml       — API 仕様
  swagger/           — Swagger UI 静的ファイル（/swagger/ で配信）
database.dbml        — DB スキーマ定義（source of truth）
```

## Database Schema

| テーブル | 用途 |
|---|---|
| users | ユーザー情報 |
| refresh_tokens | リフレッシュトークン（SHA-256ハッシュ保存） |
| teams | チームマスター（seed投入） |
| seats | 座席マスター・グレード/価格（seed投入） |
| games | 試合情報（seed投入予定） |
| tickets | 試合×座席の組み合わせ、status管理（seed投入予定） |
| reservations | 予約情報 |

**Enum:**
- `ticket_status`: available / reserved / sold
- `reservation_status`: pending / confirmed / canceled / expired

**Seed の依存順序:** teams → seats → games → tickets

## Commands

```bash
# DB のみ起動
docker compose up -d db

# フルスタック起動（air でホットリロード）
docker compose up -d

# サーバーバイナリをローカルビルド
go build -o ./tmp/main ./cmd/server

# seed 実行
go run ./cmd/seed/main.go

# テスト
go test ./...
go test ./handler/... -run TestFunctionName

# migration 適用
~/go/bin/migrate -database "postgres://postgres:postgres@localhost:55432/sports_tickets?sslmode=disable" -path migrations up

# migration バージョン確認
~/go/bin/migrate -database "postgres://postgres:postgres@localhost:55432/sports_tickets?sslmode=disable" -path migrations version

# DB テーブル確認（psql はホストにないため Docker 経由）
docker exec -it sports_tickets_db psql -U postgres -d sports_tickets -c "\dt"
```

## Environment Variables

| 変数 | デフォルト | 必須 |
|---|---|---|
| `SERVER_PORT` | `8080` | no |
| `SERVER_HOST` | `0.0.0.0` | no |
| `DB_USER` | — | yes |
| `DB_PASSWORD` | — | yes |
| `DB_NAME` | — | yes |
| `DB_HOST` | — | yes |
| `DB_PORT` | `5432` | no |

**ポートの注意:**
- ホスト上で `go run` する場合: `DB_HOST=localhost`, `DB_PORT=55432`（.env の設定）
- Docker コンテナ内アプリの場合: `DB_HOST=db`, `DB_PORT=5432`（compose.yaml で上書き）

## Auth Bundle

`authbundle/` に認証機能一式が提供済み。自前実装不要。

- `GenerateAccessToken` / `ValidateAccessToken` — JWT
- `GenerateRefreshToken` / `ValidateRefreshToken` / `RotateRefreshToken` — リフレッシュトークン
- `HashPassword` / `CheckPassword` — bcrypt
- `SetAuthCookies` — Cookie 設定
- `AuthRequired` — 認証 middleware（`handler/middleware.go`）

認証済みユーザーの ID は `authbundle.GetUserIDFromContext(ctx)` で取得。

## 新規エンドポイント追加手順

1. `handler/` にハンドラメソッドを追加
2. `handler/handler.go` の `Routes()` にルート登録
3. 認証が必要なルートは `h.AuthRequired(...)` でラップ
4. `docs/openapi.yaml` を更新
