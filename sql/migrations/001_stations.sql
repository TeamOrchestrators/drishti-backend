-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TABLE stations
(
    id                 UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    code               TEXT        NOT NULL UNIQUE,
    name               TEXT        NOT NULL,
    station_type       TEXT        NOT NULL DEFAULT 'station'
        CHECK (station_type IN ('station', 'depot', 'camp', 'hub', 'other')),
    latitude           NUMERIC(9, 6),
    longitude          NUMERIC(9, 6),
    personnel_capacity INTEGER,
    notes              TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
    CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180),
    CHECK (personnel_capacity IS NULL OR personnel_capacity >= 0)
);

CREATE INDEX idx_stations_type ON stations (station_type);

CREATE TRIGGER trg_stations_updated_at
    BEFORE UPDATE
    ON stations
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE stations;
DROP FUNCTION IF EXISTS set_updated_at();
