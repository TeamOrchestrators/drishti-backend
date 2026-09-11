-- name: ListEmergencyDeviceFormOptions :many
SELECT p.id AS personnel_id,
       p.full_name AS personnel_name,
       p.role,
       COALESCE(s.name, '') AS current_station_name,
       d.id AS device_id,
       d.device_label,
       d.status AS device_status
FROM personnel p
LEFT JOIN stations s ON s.id = p.current_station_id
LEFT JOIN emergency_devices d ON d.personnel_id = p.id
WHERE p.status <> 'inactive'
ORDER BY p.full_name ASC;

-- name: RegisterEmergencyDevice :one
INSERT INTO emergency_devices (personnel_id, device_label, status)
SELECT p.id, sqlc.arg(device_label), 'active'
FROM personnel p
WHERE p.id = sqlc.arg(personnel_id)
ON CONFLICT (personnel_id) DO UPDATE
    SET device_label = EXCLUDED.device_label,
        status = 'active'
RETURNING id, personnel_id, device_label, status, last_heartbeat_at;

-- name: RecordEmergencyHeartbeat :one
WITH active_device AS (
    SELECT d.id, d.personnel_id
    FROM emergency_devices d
    WHERE d.id = sqlc.arg(device_id)
      AND d.status = 'active'
    FOR UPDATE
), active_emergency AS (
    SELECT e.id
    FROM emergencies e
    JOIN active_device d ON d.id = e.emergency_device_id
    WHERE e.status IN ('active', 'acknowledged', 'responding')
    ORDER BY e.reported_at DESC
    LIMIT 1
), updated_device AS (
    UPDATE emergency_devices d
    SET last_heartbeat_at = sqlc.arg(occurred_at)::TIMESTAMPTZ,
        last_latitude = sqlc.arg(latitude)::NUMERIC,
        last_longitude = sqlc.arg(longitude)::NUMERIC,
        last_accuracy_m = sqlc.arg(location_accuracy_m)::NUMERIC,
        last_altitude_m = sqlc.narg(altitude_m)::NUMERIC,
        last_heading_deg = sqlc.narg(heading_deg)::NUMERIC,
        last_speed_mps = sqlc.narg(speed_mps)::NUMERIC,
        last_battery_percent = sqlc.narg(battery_percent)::NUMERIC
    FROM active_device ad
    WHERE d.id = ad.id
    RETURNING d.id, d.personnel_id
), signal AS (
    INSERT INTO emergency_signals (
        idempotency_key, personnel_id, device_id, emergency_id, signal_type,
        latitude, longitude, location_accuracy_m, altitude_m, heading_deg, speed_mps, battery_percent,
        occurred_at, sync_status
    )
    SELECT sqlc.arg(idempotency_key), ud.personnel_id, ud.id, ae.id, 'heartbeat',
           sqlc.arg(latitude)::NUMERIC, sqlc.arg(longitude)::NUMERIC, sqlc.arg(location_accuracy_m)::NUMERIC,
           sqlc.narg(altitude_m)::NUMERIC, sqlc.narg(heading_deg)::NUMERIC,
           sqlc.narg(speed_mps)::NUMERIC, sqlc.narg(battery_percent)::NUMERIC,
           sqlc.arg(occurred_at)::TIMESTAMPTZ, 'processed'
    FROM updated_device ud
    LEFT JOIN active_emergency ae ON TRUE
    ON CONFLICT (idempotency_key) DO UPDATE SET idempotency_key = EXCLUDED.idempotency_key
    RETURNING emergency_id
)
SELECT ud.id AS device_id, ud.personnel_id, s.emergency_id
FROM updated_device ud
JOIN signal s ON TRUE;

-- name: InitiateEmergencySOS :one
INSERT INTO emergency_sos_confirmations (device_id, personnel_id, emergency_type, severity, summary, expires_at)
SELECT d.id, d.personnel_id, sqlc.arg(emergency_type), sqlc.arg(severity), sqlc.arg(summary), NOW() + INTERVAL '30 seconds'
FROM emergency_devices d
WHERE d.id = sqlc.arg(device_id)
  AND d.status = 'active'
  AND d.last_latitude IS NOT NULL
  AND d.last_longitude IS NOT NULL
RETURNING id, device_id, personnel_id, expires_at, status;

-- name: ConfirmEmergencySOS :one
WITH confirmation AS (
    SELECT c.id, c.device_id, c.personnel_id, c.emergency_type, c.severity, c.summary,
           d.last_latitude, d.last_longitude, d.last_accuracy_m, d.last_altitude_m,
           d.last_heading_deg, d.last_speed_mps, d.last_battery_percent, d.last_heartbeat_at
    FROM emergency_sos_confirmations c
    JOIN emergency_devices d ON d.id = c.device_id
    WHERE c.id = sqlc.arg(confirmation_id)
      AND c.device_id = sqlc.arg(device_id)
      AND c.status = 'pending'
      AND c.expires_at >= NOW()
    FOR UPDATE OF c
), new_emergency AS (
    INSERT INTO emergencies (
        emergency_code, emergency_type, station_id, reported_by_personnel_id, emergency_device_id,
        report_channel, severity, status, latitude, longitude, location_accuracy_m, summary
    )
    SELECT CONCAT('EMG-', UPPER(SUBSTRING(REPLACE(gen_random_uuid()::TEXT, '-', '') FROM 1 FOR 8))),
           c.emergency_type, p.current_station_id, c.personnel_id, c.device_id,
           'mobile_sos', c.severity, 'active', c.last_latitude, c.last_longitude, c.last_accuracy_m, c.summary
    FROM confirmation c
    JOIN personnel p ON p.id = c.personnel_id
    RETURNING id, emergency_code, status, reported_at
), affected AS (
    INSERT INTO emergency_personnel (emergency_id, personnel_id, involvement_type, status)
    SELECT e.id, c.personnel_id, 'affected', 'awaiting_response'
    FROM new_emergency e CROSS JOIN confirmation c
), signal AS (
    INSERT INTO emergency_signals (
        idempotency_key, personnel_id, device_id, emergency_id, signal_type,
        latitude, longitude, location_accuracy_m, altitude_m, heading_deg, speed_mps, battery_percent,
        occurred_at, sync_status
    )
    SELECT sqlc.arg(idempotency_key), c.personnel_id, c.device_id, e.id, 'sos',
           c.last_latitude, c.last_longitude, c.last_accuracy_m, c.last_altitude_m,
           c.last_heading_deg, c.last_speed_mps, c.last_battery_percent,
           COALESCE(c.last_heartbeat_at, NOW()), 'processed'
    FROM confirmation c CROSS JOIN new_emergency e
), timeline AS (
    INSERT INTO emergency_timeline_events (emergency_id, event_type, details, recorded_by_personnel_id)
    SELECT e.id, 'sos_confirmed', 'Two-step SOS confirmed from mobile HTTP simulator.', c.personnel_id
    FROM confirmation c CROSS JOIN new_emergency e
), updated_confirmation AS (
    UPDATE emergency_sos_confirmations sc
    SET status = 'confirmed', confirmed_at = NOW(), emergency_id = e.id
    FROM new_emergency e
    WHERE sc.id = sqlc.arg(confirmation_id)
    RETURNING sc.id
)
SELECT e.id, e.emergency_code, e.status, e.reported_at
FROM new_emergency e
JOIN updated_confirmation uc ON TRUE;

-- name: ListEmergencies :many
SELECT e.id,
       e.emergency_code,
       e.emergency_type,
       e.severity,
       e.status,
       e.summary,
       COALESCE(e.latitude, 0)::DOUBLE PRECISION AS latitude,
       COALESCE(e.longitude, 0)::DOUBLE PRECISION AS longitude,
       COALESCE(e.location_accuracy_m, 0)::DOUBLE PRECISION AS location_accuracy_m,
       e.reported_at,
       e.resolved_at,
       COALESCE(p.full_name, '') AS reported_by_name,
       d.id AS device_id,
       COALESCE(d.device_label, '') AS device_label,
       d.last_heartbeat_at,
       COALESCE(d.last_heading_deg, 0)::DOUBLE PRECISION AS heading_deg,
       COALESCE(d.last_speed_mps, 0)::DOUBLE PRECISION AS speed_mps,
       COALESCE(d.last_battery_percent, 0)::DOUBLE PRECISION AS battery_percent,
       COALESCE(affected.people_affected, 0)::INTEGER AS people_affected,
       COALESCE(resources.resources_needed, '') AS resources_needed
FROM emergencies e
LEFT JOIN personnel p ON p.id = e.reported_by_personnel_id
LEFT JOIN emergency_devices d ON d.id = e.emergency_device_id
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS people_affected
    FROM emergency_personnel ep
    WHERE ep.emergency_id = e.id AND ep.involvement_type = 'affected'
) affected ON TRUE
LEFT JOIN LATERAL (
    SELECT string_agg(i.name, ', ' ORDER BY i.name) AS resources_needed
    FROM emergency_resource_requests er
    JOIN items i ON i.id = er.item_id
    WHERE er.emergency_id = e.id
) resources ON TRUE
ORDER BY CASE WHEN e.status IN ('active', 'acknowledged', 'responding') THEN 0 ELSE 1 END,
         e.reported_at DESC;

-- name: UpdateEmergencyStatus :one
UPDATE emergencies
SET status = sqlc.arg(status),
    resolved_at = CASE WHEN sqlc.arg(status) IN ('resolved', 'cancelled') THEN NOW() ELSE NULL END
WHERE id = sqlc.arg(emergency_id)
RETURNING id, emergency_code, status, resolved_at;
