-- 開発環境用の初期設定
-- 本番環境では使用しないこと

-- データベースが存在しない場合は作成
SELECT 'CREATE DATABASE road_to_dena'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'road_to_dena')\gexec

-- road_to_denaデータベースに接続
\c road_to_dena;

-- 拡張機能の有効化
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";