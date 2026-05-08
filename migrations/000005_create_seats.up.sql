CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 更新日時を自動的に更新するトリガー関数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS seats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    grade VARCHAR(5) NOT NULL,
    price INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_seats_grade ON seats(grade);

CREATE TRIGGER update_seats_updated_at BEFORE UPDATE
    ON seats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();