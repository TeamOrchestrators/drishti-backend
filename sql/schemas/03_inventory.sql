CREATE TABLE items
(
    id             UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    item_code      TEXT        NOT NULL UNIQUE,
    name           TEXT        NOT NULL,
    category       TEXT        NOT NULL,
    unit           TEXT        NOT NULL,
    unit_weight_kg NUMERIC(14, 3),
    unit_volume_m3 NUMERIC(14, 3),
    is_critical    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (unit_weight_kg IS NULL OR unit_weight_kg >= 0),
    CHECK (unit_volume_m3 IS NULL OR unit_volume_m3 >= 0)
);

CREATE TABLE station_inventory
(
    id                 UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    station_id         UUID           NOT NULL REFERENCES stations (id),
    item_id            UUID           NOT NULL REFERENCES items (id),
    on_hand_quantity   NUMERIC(14, 3) NOT NULL DEFAULT 0 CHECK (on_hand_quantity >= 0),
    reserved_quantity  NUMERIC(14, 3) NOT NULL DEFAULT 0 CHECK (reserved_quantity >= 0),
    available_quantity NUMERIC(14, 3) GENERATED ALWAYS AS (on_hand_quantity - reserved_quantity) STORED,
    minimum_quantity   NUMERIC(14, 3) NOT NULL DEFAULT 0 CHECK (minimum_quantity >= 0),
    maximum_quantity   NUMERIC(14, 3),
    updated_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    UNIQUE (station_id, item_id),
    CHECK (maximum_quantity IS NULL OR maximum_quantity >= minimum_quantity),
    CHECK (reserved_quantity <= on_hand_quantity)
);

CREATE TABLE inventory_transactions
(
    id                      UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    station_inventory_id    UUID           NOT NULL REFERENCES station_inventory (id),
    transaction_type        TEXT           NOT NULL
        CHECK (transaction_type IN
               ('receipt', 'consumption', 'adjustment', 'reservation', 'release', 'transfer_in', 'transfer_out')),
    on_hand_delta           NUMERIC(14, 3) NOT NULL DEFAULT 0,
    reserved_delta          NUMERIC(14, 3) NOT NULL DEFAULT 0,
    occurred_at             TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    reference_type          TEXT,
    reference_id            UUID,
    notes                   TEXT,
    created_by_personnel_id UUID REFERENCES personnel (id),
    CHECK (on_hand_delta <> 0 OR reserved_delta <> 0)
);

CREATE INDEX idx_items_category ON items (category);
CREATE INDEX idx_inventory_station ON station_inventory (station_id);
CREATE INDEX idx_inventory_item ON station_inventory (item_id);
CREATE INDEX idx_inventory_transactions_inventory_occurred
    ON inventory_transactions (station_inventory_id, occurred_at DESC);

CREATE TRIGGER trg_items_updated_at
    BEFORE UPDATE
    ON items
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_station_inventory_updated_at
    BEFORE UPDATE
    ON station_inventory
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
