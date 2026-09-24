-- name: ListLogisticsBatches :many
SELECT b.id, b.batch_code, e.name AS expedition_name,
       origin.name AS origin_station_name, destination.name AS destination_station_name,
       b.status, b.planned_dispatch_at, b.estimated_arrival_at,
       COUNT(DISTINCT c.id)::INTEGER AS cargo_count
FROM logistics_batches b
JOIN expeditions e ON e.id = b.expedition_id
JOIN stations origin ON origin.id = b.origin_station_id
JOIN stations destination ON destination.id = b.destination_station_id
LEFT JOIN cargo c ON c.logistics_batch_id = b.id
GROUP BY b.id, e.name, origin.name, destination.name
ORDER BY b.planned_dispatch_at ASC NULLS LAST, b.created_at DESC;

-- name: ListCargo :many
SELECT c.id, c.cargo_code, origin.name AS origin_station_name,
       destination.name AS destination_station_name, COALESCE(e.name, '') AS expedition_name,
       COALESCE(b.batch_code, '') AS logistics_batch_code, c.priority, c.status, c.notes
FROM cargo c
LEFT JOIN stations origin ON origin.id = c.origin_station_id
JOIN stations destination ON destination.id = c.destination_station_id
LEFT JOIN expeditions e ON e.id = c.expedition_id
LEFT JOIN logistics_batches b ON b.id = c.logistics_batch_id
ORDER BY c.created_at DESC;

-- name: CreateLogisticsBatch :one
INSERT INTO logistics_batches (batch_code, expedition_id, origin_station_id, destination_station_id, status, planned_dispatch_at, estimated_arrival_at, notes)
SELECT sqlc.arg(batch_code), e.id, e.origin_station_id, e.destination_station_id,
       sqlc.arg(status), sqlc.arg(planned_dispatch_at), sqlc.arg(estimated_arrival_at), sqlc.narg(notes)
FROM expeditions e
WHERE e.id = sqlc.arg(expedition_id)
  AND e.status NOT IN ('completed', 'cancelled')
  AND e.origin_station_id IS NOT NULL AND e.destination_station_id IS NOT NULL
RETURNING *;

-- name: UpdateLogisticsBatch :one
UPDATE logistics_batches
SET status = sqlc.arg(status), planned_dispatch_at = sqlc.arg(planned_dispatch_at),
    estimated_arrival_at = sqlc.arg(estimated_arrival_at), notes = sqlc.narg(notes)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CreateCargo :one
INSERT INTO cargo (cargo_code, origin_station_id, destination_station_id, expedition_id, logistics_batch_id, priority, status, notes)
VALUES (sqlc.arg(cargo_code), sqlc.arg(origin_station_id), sqlc.arg(destination_station_id),
        sqlc.narg(expedition_id), sqlc.narg(logistics_batch_id), sqlc.arg(priority), sqlc.arg(status), sqlc.narg(notes))
RETURNING *;

-- name: AssignCargoToLogisticsBatch :one
UPDATE cargo c
SET logistics_batch_id = b.id, expedition_id = b.expedition_id,
    origin_station_id = b.origin_station_id, destination_station_id = b.destination_station_id
FROM logistics_batches b
WHERE c.id = sqlc.arg(cargo_id) AND b.id = sqlc.arg(logistics_batch_id)
  AND b.status NOT IN ('received', 'cancelled')
RETURNING c.*;

-- name: GetCargoForQR :one
SELECT c.id,
       c.cargo_code,
       COALESCE(origin.name, '') AS origin_station_name,
       destination.name AS destination_station_name,
       COALESCE(e.expedition_code, '') AS expedition_code,
       COALESCE(e.name, '') AS expedition_name,
       COALESCE(b.batch_code, '') AS logistics_batch_code,
       c.priority,
       c.status,
       c.notes,
       c.dispatched_at,
       c.received_at,
       c.created_at
FROM cargo c
LEFT JOIN stations origin ON origin.id = c.origin_station_id
JOIN stations destination ON destination.id = c.destination_station_id
LEFT JOIN expeditions e ON e.id = c.expedition_id
LEFT JOIN logistics_batches b ON b.id = c.logistics_batch_id
WHERE c.id = sqlc.arg(cargo_id);

-- name: ListCargoItemsForQR :many
SELECT i.id AS item_id,
       i.item_code,
       i.name AS item_name,
       i.unit,
       ci.quantity::DOUBLE PRECISION AS quantity,
       COALESCE(ci.declared_weight_kg, 0)::DOUBLE PRECISION AS declared_weight_kg,
       ci.notes
FROM cargo_items ci
JOIN items i ON i.id = ci.item_id
WHERE ci.cargo_id = sqlc.arg(cargo_id)
ORDER BY i.name ASC;

-- name: ListCargoQRScans :many
SELECT q.id,
       q.event_type,
       COALESCE(s.name, '') AS station_name,
       COALESCE(p.full_name, '') AS scanned_by_personnel_name,
       COALESCE(q.latitude, 0)::DOUBLE PRECISION AS latitude,
       COALESCE(q.longitude, 0)::DOUBLE PRECISION AS longitude,
       q.scanned_at,
       q.notes
FROM cargo_qr_scans q
LEFT JOIN stations s ON s.id = q.station_id
LEFT JOIN personnel p ON p.id = q.scanned_by_personnel_id
WHERE q.cargo_id = sqlc.arg(cargo_id)
ORDER BY q.scanned_at DESC;

-- name: RecordCargoQRScan :one
INSERT INTO cargo_qr_scans (
    cargo_id, event_type, station_id, scanned_by_personnel_id,
    latitude, longitude, notes
)
VALUES (
    sqlc.arg(cargo_id), sqlc.arg(event_type), sqlc.narg(station_id),
    sqlc.narg(scanned_by_personnel_id), sqlc.narg(latitude)::DOUBLE PRECISION,
    sqlc.narg(longitude)::DOUBLE PRECISION, sqlc.narg(notes)
)
RETURNING id, scanned_at;

-- name: UpdateCargoStatusFromQR :one
UPDATE cargo
SET status = sqlc.arg(status),
    dispatched_at = CASE
        WHEN sqlc.arg(status) IN ('dispatched', 'in_transit') THEN COALESCE(dispatched_at, NOW())
        ELSE dispatched_at
    END,
    received_at = CASE
        WHEN sqlc.arg(status) = 'received' THEN COALESCE(received_at, NOW())
        ELSE received_at
    END
WHERE id = sqlc.arg(cargo_id)
RETURNING id;
