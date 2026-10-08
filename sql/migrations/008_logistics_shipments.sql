-- +goose Up
CREATE TABLE logistics_shipments
(
    id                     UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    shipment_code          TEXT        NOT NULL UNIQUE,
    logistics_batch_id     UUID        NOT NULL REFERENCES logistics_batches (id) ON DELETE CASCADE,
    origin_station_id      UUID        NOT NULL REFERENCES stations (id),
    destination_station_id UUID        NOT NULL REFERENCES stations (id),
    transport_mode         TEXT        NOT NULL DEFAULT 'ground'
        CHECK (transport_mode IN ('ground', 'airlift', 'sea_lift', 'mixed', 'other')),
    priority               TEXT        NOT NULL DEFAULT 'standard'
        CHECK (priority IN ('standard', 'high', 'critical')),
    status                 TEXT        NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'planned', 'packed', 'dispatched', 'received', 'delayed', 'cancelled')),
    is_default_route       BOOLEAN     NOT NULL DEFAULT FALSE,
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

ALTER TABLE cargo
    ADD COLUMN logistics_shipment_id UUID REFERENCES logistics_shipments (id),
    ADD COLUMN source_cargo_id UUID REFERENCES cargo (id);

-- Existing batch cargo continues on a generated default route, preserving the
-- old one-batch/one-route behaviour until it is explicitly split.
INSERT INTO logistics_shipments (
    shipment_code, logistics_batch_id, origin_station_id, destination_station_id,
    transport_mode, priority, status, is_default_route, planned_dispatch_at,
    estimated_arrival_at, dispatched_at, received_at, notes
)
SELECT b.batch_code || '-DEFAULT', b.id, b.origin_station_id, b.destination_station_id,
       'ground', 'standard', b.status, TRUE, b.planned_dispatch_at,
       b.estimated_arrival_at, b.dispatched_at, b.received_at,
       'Compatibility default route created during shipment migration.'
FROM logistics_batches b;

UPDATE cargo c
SET logistics_shipment_id = s.id
FROM logistics_shipments s
WHERE s.logistics_batch_id = c.logistics_batch_id
  AND s.is_default_route = TRUE;

CREATE INDEX idx_cargo_logistics_shipment ON cargo (logistics_shipment_id);
CREATE INDEX idx_cargo_source_cargo ON cargo (source_cargo_id);
CREATE INDEX idx_logistics_shipments_batch_status ON logistics_shipments (logistics_batch_id, status);
CREATE UNIQUE INDEX uq_logistics_shipments_default_route
    ON logistics_shipments (logistics_batch_id)
    WHERE is_default_route;

CREATE TRIGGER trg_logistics_shipments_updated_at
    BEFORE UPDATE
    ON logistics_shipments
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose Down
ALTER TABLE cargo
    DROP COLUMN IF EXISTS logistics_shipment_id,
    DROP COLUMN IF EXISTS source_cargo_id;
DROP TRIGGER IF EXISTS trg_logistics_shipments_updated_at ON logistics_shipments;
DROP TABLE IF EXISTS logistics_shipments;
