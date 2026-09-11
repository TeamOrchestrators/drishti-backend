CREATE TABLE logistics_batches
(
    id                     UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    batch_code             TEXT        NOT NULL UNIQUE,
    expedition_id          UUID        NOT NULL REFERENCES expeditions (id),
    origin_station_id      UUID        NOT NULL REFERENCES stations (id),
    destination_station_id UUID        NOT NULL REFERENCES stations (id),
    status                 TEXT        NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'planned', 'packed', 'dispatched', 'received', 'delayed', 'cancelled')),
    planned_dispatch_at    TIMESTAMPTZ,
    estimated_arrival_at   TIMESTAMPTZ,
    dispatched_at          TIMESTAMPTZ,
    received_at            TIMESTAMPTZ,
    notes                  TEXT,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (origin_station_id <> destination_station_id),
    CHECK (estimated_arrival_at IS NULL OR planned_dispatch_at IS NULL OR estimated_arrival_at >= planned_dispatch_at),
    CHECK (received_at IS NULL OR dispatched_at IS NULL OR received_at >= dispatched_at)
);

CREATE TABLE cargo
(
    id                      UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    cargo_code              TEXT        NOT NULL UNIQUE,
    origin_station_id       UUID REFERENCES stations (id),
    destination_station_id  UUID        NOT NULL REFERENCES stations (id),
    expedition_id           UUID REFERENCES expeditions (id),
    logistics_batch_id      UUID REFERENCES logistics_batches (id),
    priority                TEXT        NOT NULL DEFAULT 'standard'
        CHECK (priority IN ('standard', 'high', 'critical')),
    status                  TEXT        NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'packed', 'dispatched', 'in_transit', 'received', 'delayed', 'cancelled')),
    created_by_personnel_id UUID REFERENCES personnel (id),
    dispatched_at           TIMESTAMPTZ,
    received_at             TIMESTAMPTZ,
    notes                   TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (received_at IS NULL OR dispatched_at IS NULL OR received_at >= dispatched_at)
);

CREATE TABLE cargo_items
(
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cargo_id           UUID           NOT NULL REFERENCES cargo (id) ON DELETE CASCADE,
    item_id            UUID           NOT NULL REFERENCES items (id),
    quantity           NUMERIC(14, 3) NOT NULL CHECK (quantity > 0),
    declared_weight_kg NUMERIC(14, 3),
    notes              TEXT,
    UNIQUE (cargo_id, item_id),
    CHECK (declared_weight_kg IS NULL OR declared_weight_kg >= 0)
);

CREATE TABLE cargo_qr_scans
(
    id                      UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    cargo_id                UUID        NOT NULL REFERENCES cargo (id) ON DELETE CASCADE,
    event_type              TEXT        NOT NULL
        CHECK (event_type IN ('created', 'packed', 'dispatched', 'scanned', 'received', 'damaged', 'missing')),
    station_id              UUID REFERENCES stations (id),
    scanned_by_personnel_id UUID REFERENCES personnel (id),
    latitude                NUMERIC(9, 6),
    longitude               NUMERIC(9, 6),
    scanned_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes                   TEXT,
    CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
    CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180)
);

CREATE INDEX idx_cargo_destination_status ON cargo (destination_station_id, status);
CREATE INDEX idx_cargo_expedition ON cargo (expedition_id);
CREATE INDEX idx_cargo_logistics_batch ON cargo (logistics_batch_id);
CREATE INDEX idx_cargo_qr_scans_cargo_scanned ON cargo_qr_scans (cargo_id, scanned_at DESC);

CREATE INDEX idx_logistics_batches_expedition_status ON logistics_batches (expedition_id, status);

CREATE TRIGGER trg_logistics_batches_updated_at
    BEFORE UPDATE
    ON logistics_batches
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_cargo_updated_at
    BEFORE UPDATE
    ON cargo
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
