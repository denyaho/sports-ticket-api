CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 更新日時を自動的に更新するトリガー関数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TYPE ticket_status AS ENUM (
    'available',
    'reserved',
    'sold'
);

CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    seat_id UUID NOT NULL,
    game_id UUID NOT NULL,
    reservation_id UUID,
    price INTEGER NOT NULL,
    status ticket_status NOT NULL DEFAULT 'available',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (seat_id) REFERENCES seats(id),
    FOREIGN KEY (game_id) REFERENCES games(id),
    FOREIGN KEY (reservation_id) REFERENCES reservations(id),

    CONSTRAINT unique_ticket_seat_game UNIQUE (seat_id, game_id)
);

CREATE INDEX idx_tickets_game_id ON tickets(game_id);
CREATE INDEX idx_tickets_seat_id ON tickets(seat_id);
CREATE INDEX idx_tickets_reservation_id ON tickets(reservation_id);

CREATE TRIGGER update_tickets_updated_at BEFORE UPDATE
    ON tickets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();