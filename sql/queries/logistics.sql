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
