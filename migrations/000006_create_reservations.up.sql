CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


-- 更新日時を自動的に更新するトリガー関数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TYPE reservation_status AS ENUM (
    'pending',
    'confirmed',
    'cancelled',
    'expired'
);

CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    game_id UUID NOT NULL,
    status reservation_status NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (game_id) REFERENCES games(id)
);

CREATE INDEX idx_reservations_user_id ON reservations(user_id);
CREATE INDEX idx_reservations_game_id ON reservations(game_id);

CREATE TRIGGER update_reservations_updated_at BEFORE UPDATE
    ON reservations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();