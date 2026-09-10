-- +goose Up
CREATE TABLE emergencies
(
    id                       UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    emergency_code           TEXT        NOT NULL UNIQUE,
    emergency_type           TEXT        NOT NULL,
    station_id               UUID REFERENCES stations (id),
    expedition_id            UUID REFERENCES expeditions (id),
    reported_by_personnel_id UUID REFERENCES personnel (id),
    report_channel           TEXT        NOT NULL DEFAULT 'manual'
        CHECK (report_channel IN ('mobile_sos', 'manual', 'control_room')),
    severity                 TEXT        NOT NULL
        CHECK (severity IN ('moderate', 'critical', 'immediate_response')),
    status                   TEXT        NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'acknowledged', 'responding', 'resolved', 'cancelled')),
    latitude                 NUMERIC(9, 6),
    longitude                NUMERIC(9, 6),
    location_accuracy_m      NUMERIC(12, 3),
    summary                  TEXT        NOT NULL,
    reported_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at              TIMESTAMPTZ,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
    CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180),
    CHECK (location_accuracy_m IS NULL OR location_accuracy_m >= 0),
    CHECK (resolved_at IS NULL OR resolved_at >= reported_at)
);

CREATE TABLE emergency_signals
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    idempotency_key      UUID        NOT NULL UNIQUE,
    personnel_id         UUID        NOT NULL REFERENCES personnel (id),
    emergency_id         UUID REFERENCES emergencies (id),
    signal_type          TEXT        NOT NULL
        CHECK (signal_type IN ('sos', 'check_in', 'location_update')),
    transmission_channel TEXT        NOT NULL DEFAULT 'mobile_http_simulator'
        CHECK (transmission_channel = 'mobile_http_simulator'),
    latitude             NUMERIC(9, 6),
    longitude            NUMERIC(9, 6),
    location_accuracy_m  NUMERIC(12, 3),
    occurred_at          TIMESTAMPTZ NOT NULL,
    received_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sync_status          TEXT        NOT NULL DEFAULT 'received'
        CHECK (sync_status IN ('received', 'processed', 'rejected')),
    payload_notes        TEXT,
    CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
    CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180),
    CHECK (location_accuracy_m IS NULL OR location_accuracy_m >= 0)
);

CREATE TABLE emergency_personnel
(
    emergency_id     UUID NOT NULL REFERENCES emergencies (id) ON DELETE CASCADE,
    personnel_id     UUID NOT NULL REFERENCES personnel (id),
    involvement_type TEXT NOT NULL
        CHECK (involvement_type IN ('affected', 'responder', 'coordinator')),
    status           TEXT,
    PRIMARY KEY (emergency_id, personnel_id)
);

CREATE TABLE emergency_resource_requests
(
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    emergency_id      UUID           NOT NULL REFERENCES emergencies (id) ON DELETE CASCADE,
    item_id           UUID           NOT NULL REFERENCES items (id),
    required_quantity NUMERIC(14, 3) NOT NULL CHECK (required_quantity > 0),
    notes             TEXT,
    UNIQUE (emergency_id, item_id)
);

CREATE TABLE emergency_resource_allocations
(
    id                            UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    emergency_resource_request_id UUID           NOT NULL REFERENCES emergency_resource_requests (id) ON DELETE CASCADE,
    station_inventory_id          UUID           NOT NULL REFERENCES station_inventory (id),
    quantity                      NUMERIC(14, 3) NOT NULL CHECK (quantity > 0),
    status                        TEXT           NOT NULL DEFAULT 'reserved'
        CHECK (status IN ('reserved', 'dispatched', 'delivered', 'released')),
    allocated_at                  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE emergency_timeline_events
(
    id                       UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    emergency_id             UUID        NOT NULL REFERENCES emergencies (id) ON DELETE CASCADE,
    event_type               TEXT        NOT NULL,
    details                  TEXT        NOT NULL,
    occurred_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    recorded_by_personnel_id UUID REFERENCES personnel (id)
);

CREATE INDEX idx_emergencies_status_reported ON emergencies (status, reported_at DESC);
CREATE INDEX idx_emergency_signals_personnel_received ON emergency_signals (personnel_id, received_at DESC);
CREATE INDEX idx_emergency_signals_emergency ON emergency_signals (emergency_id);
CREATE INDEX idx_emergency_personnel_personnel ON emergency_personnel (personnel_id);
CREATE INDEX idx_emergency_timeline_emergency_occurred
    ON emergency_timeline_events (emergency_id, occurred_at);

CREATE TRIGGER trg_emergencies_updated_at
    BEFORE UPDATE
    ON emergencies
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE emergency_timeline_events;
DROP TABLE emergency_resource_allocations;
DROP TABLE emergency_resource_requests;
DROP TABLE emergency_personnel;
DROP TABLE emergency_signals;
DROP TABLE emergencies;
