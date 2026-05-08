CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 更新日時を自動的に更新するトリガー関数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    home_team_id UUID NOT NULL,
    away_team_id UUID NOT NULL,
    game_date TIMESTAMP NOT NULL,
    start_time TIMESTAMP NOT NULL,
    venue VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (home_team_id) REFERENCES teams(id),
    FOREIGN KEY (away_team_id) REFERENCES teams(id)
);

CREATE INDEX idx_games_game_date ON games(game_date);

CREATE TRIGGER update_games_updated_at BEFORE UPDATE
    ON games FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();