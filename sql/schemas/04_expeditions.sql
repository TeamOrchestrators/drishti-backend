CREATE TABLE expeditions
(
    id                     UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    expedition_code        TEXT        NOT NULL UNIQUE,
    name                   TEXT        NOT NULL,
    purpose                TEXT,
    origin_station_id      UUID REFERENCES stations (id),
    destination_station_id UUID REFERENCES stations (id),
    lead_personnel_id      UUID REFERENCES personnel (id),
    planned_start_at       TIMESTAMPTZ,
    planned_end_at         TIMESTAMPTZ,
    actual_start_at        TIMESTAMPTZ,
    actual_end_at          TIMESTAMPTZ,
    status                 TEXT        NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'planned', 'ready', 'active', 'sheltered', 'completed', 'cancelled')),
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (planned_end_at IS NULL OR planned_start_at IS NULL OR planned_end_at >= planned_start_at),
    CHECK (actual_end_at IS NULL OR actual_start_at IS NULL OR actual_end_at >= actual_start_at)
);

ALTER TABLE personnel_movements
    ADD CONSTRAINT personnel_movements_expedition_id_fkey
    FOREIGN KEY (expedition_id) REFERENCES expeditions (id);

CREATE TABLE expedition_members
(
    expedition_id   UUID        NOT NULL REFERENCES expeditions (id) ON DELETE CASCADE,
    personnel_id    UUID        NOT NULL REFERENCES personnel (id),
    assignment_role TEXT,
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    released_at     TIMESTAMPTZ,
    PRIMARY KEY (expedition_id, personnel_id),
    CHECK (released_at IS NULL OR released_at >= assigned_at)
);

CREATE TABLE expedition_resource_requirements
(
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expedition_id     UUID           NOT NULL REFERENCES expeditions (id) ON DELETE CASCADE,
    item_id           UUID           NOT NULL REFERENCES items (id),
    required_quantity NUMERIC(14, 3) NOT NULL CHECK (required_quantity > 0),
    notes             TEXT,
    UNIQUE (expedition_id, item_id)
);

CREATE TABLE expedition_resource_allocations
(
    id                   UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    requirement_id       UUID           NOT NULL REFERENCES expedition_resource_requirements (id) ON DELETE CASCADE,
    station_inventory_id UUID           NOT NULL REFERENCES station_inventory (id),
    quantity             NUMERIC(14, 3) NOT NULL CHECK (quantity > 0),
    status               TEXT           NOT NULL DEFAULT 'reserved'
        CHECK (status IN ('reserved', 'issued', 'released')),
    allocated_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_expeditions_status ON expeditions (status);
CREATE INDEX idx_expedition_members_personnel ON expedition_members (personnel_id);
CREATE INDEX idx_expedition_allocations_inventory ON expedition_resource_allocations (station_inventory_id);

CREATE TRIGGER trg_expeditions_updated_at
    BEFORE UPDATE
    ON expeditions
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
