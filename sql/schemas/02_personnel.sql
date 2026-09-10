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

CREATE INDEX idx_personnel_current_station ON personnel (current_station_id);
CREATE INDEX idx_personnel_status ON personnel (status);

CREATE TRIGGER trg_personnel_updated_at
    BEFORE UPDATE
    ON personnel
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
