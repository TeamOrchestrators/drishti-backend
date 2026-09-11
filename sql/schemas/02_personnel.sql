CREATE TABLE personnel
(
    id                       UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    personnel_code           TEXT        NOT NULL UNIQUE,
    full_name                TEXT        NOT NULL,
    role                     TEXT        NOT NULL,
    medical_clearance_status TEXT        NOT NULL DEFAULT 'pending'
        CHECK (medical_clearance_status IN ('pending', 'cleared', 'restricted', 'expired')),
    current_station_id       UUID REFERENCES stations (id),
    status                   TEXT        NOT NULL DEFAULT 'available'
        CHECK (status IN ('available', 'assigned', 'in_transit', 'unavailable', 'inactive')),
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE personnel_movements
(
    id                     UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    personnel_id           UUID        NOT NULL REFERENCES personnel (id),
    expedition_id          UUID,
    movement_type          TEXT        NOT NULL
        CHECK (movement_type IN ('deployment', 'return_to_nation', 'station_transfer', 'expedition_transfer')),
    origin_station_id      UUID REFERENCES stations (id),
    destination_station_id UUID REFERENCES stations (id),
    status                 TEXT        NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned', 'in_transit', 'arrived', 'cancelled')),
    departed_at            TIMESTAMPTZ,
    estimated_arrival_at   TIMESTAMPTZ,
    arrived_at             TIMESTAMPTZ,
    notes                  TEXT,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (arrived_at IS NULL OR departed_at IS NULL OR arrived_at >= departed_at)
);

CREATE INDEX idx_personnel_current_station ON personnel (current_station_id);
CREATE INDEX idx_personnel_status ON personnel (status);
CREATE INDEX idx_personnel_movements_personnel_status ON personnel_movements (personnel_id, status);
CREATE INDEX idx_personnel_movements_destination_status ON personnel_movements (destination_station_id, status);

CREATE TRIGGER trg_personnel_updated_at
    BEFORE UPDATE
    ON personnel
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_personnel_movements_updated_at
    BEFORE UPDATE
    ON personnel_movements
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
