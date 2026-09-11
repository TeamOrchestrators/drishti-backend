-- +goose Up
CREATE TABLE emergency_devices
(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    personnel_id UUID NOT NULL UNIQUE REFERENCES personnel (id),
    device_label TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'maintenance')),
    last_heartbeat_at TIMESTAMPTZ,
    last_latitude NUMERIC(9, 6), last_longitude NUMERIC(9, 6),
    last_accuracy_m NUMERIC(12, 3), last_altitude_m NUMERIC(12, 3),
    last_heading_deg NUMERIC(6, 3), last_speed_mps NUMERIC(12, 3), last_battery_percent NUMERIC(5, 2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (last_latitude IS NULL OR last_latitude BETWEEN -90 AND 90),
    CHECK (last_longitude IS NULL OR last_longitude BETWEEN -180 AND 180),
    CHECK (last_accuracy_m IS NULL OR last_accuracy_m >= 0),
    CHECK (last_heading_deg IS NULL OR last_heading_deg >= 0 AND last_heading_deg < 360),
    CHECK (last_speed_mps IS NULL OR last_speed_mps >= 0),
    CHECK (last_battery_percent IS NULL OR last_battery_percent BETWEEN 0 AND 100)
);

ALTER TABLE emergencies ADD COLUMN emergency_device_id UUID REFERENCES emergency_devices (id);
ALTER TABLE emergency_signals ADD COLUMN device_id UUID REFERENCES emergency_devices (id);
ALTER TABLE emergency_signals ADD COLUMN altitude_m NUMERIC(12, 3);
ALTER TABLE emergency_signals ADD COLUMN heading_deg NUMERIC(6, 3);
ALTER TABLE emergency_signals ADD COLUMN speed_mps NUMERIC(12, 3);
ALTER TABLE emergency_signals ADD COLUMN battery_percent NUMERIC(5, 2);
ALTER TABLE emergency_signals DROP CONSTRAINT emergency_signals_signal_type_check;
ALTER TABLE emergency_signals ADD CONSTRAINT emergency_signals_signal_type_check
    CHECK (signal_type IN ('sos', 'check_in', 'location_update', 'heartbeat'));
ALTER TABLE emergency_signals ADD CONSTRAINT emergency_signals_heading_check CHECK (heading_deg IS NULL OR heading_deg >= 0 AND heading_deg < 360);
ALTER TABLE emergency_signals ADD CONSTRAINT emergency_signals_speed_check CHECK (speed_mps IS NULL OR speed_mps >= 0);
ALTER TABLE emergency_signals ADD CONSTRAINT emergency_signals_battery_check CHECK (battery_percent IS NULL OR battery_percent BETWEEN 0 AND 100);

CREATE TABLE emergency_sos_confirmations
(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL REFERENCES emergency_devices (id),
    personnel_id UUID NOT NULL REFERENCES personnel (id),
    emergency_type TEXT NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('moderate', 'critical', 'immediate_response')),
    summary TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'expired', 'cancelled')),
    initiated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), expires_at TIMESTAMPTZ NOT NULL,
    confirmed_at TIMESTAMPTZ, emergency_id UUID UNIQUE REFERENCES emergencies (id),
    CHECK (expires_at > initiated_at), CHECK (confirmed_at IS NULL OR confirmed_at >= initiated_at)
);

CREATE INDEX idx_emergency_devices_heartbeat ON emergency_devices (last_heartbeat_at DESC);
CREATE INDEX idx_emergency_signals_device_received ON emergency_signals (device_id, received_at DESC);
CREATE INDEX idx_emergency_sos_confirmations_device_status ON emergency_sos_confirmations (device_id, status, expires_at DESC);
CREATE TRIGGER trg_emergency_devices_updated_at BEFORE UPDATE ON emergency_devices FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_emergency_devices_updated_at ON emergency_devices;
DROP TABLE IF EXISTS emergency_sos_confirmations;
ALTER TABLE emergency_signals DROP CONSTRAINT IF EXISTS emergency_signals_battery_check;
ALTER TABLE emergency_signals DROP CONSTRAINT IF EXISTS emergency_signals_speed_check;
ALTER TABLE emergency_signals DROP CONSTRAINT IF EXISTS emergency_signals_heading_check;
ALTER TABLE emergency_signals DROP CONSTRAINT IF EXISTS emergency_signals_signal_type_check;
ALTER TABLE emergency_signals ADD CONSTRAINT emergency_signals_signal_type_check CHECK (signal_type IN ('sos', 'check_in', 'location_update'));
ALTER TABLE emergency_signals DROP COLUMN IF EXISTS battery_percent;
ALTER TABLE emergency_signals DROP COLUMN IF EXISTS speed_mps;
ALTER TABLE emergency_signals DROP COLUMN IF EXISTS heading_deg;
ALTER TABLE emergency_signals DROP COLUMN IF EXISTS altitude_m;
ALTER TABLE emergency_signals DROP COLUMN IF EXISTS device_id;
ALTER TABLE emergencies DROP COLUMN IF EXISTS emergency_device_id;
DROP TABLE IF EXISTS emergency_devices;
