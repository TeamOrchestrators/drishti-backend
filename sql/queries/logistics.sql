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
       COALESCE(b.batch_code, '') AS logistics_batch_code,
       c.logistics_shipment_id, COALESCE(shipment.shipment_code, '') AS logistics_shipment_code,
       COALESCE(shipment.transport_mode, '') AS transport_mode, c.priority, c.status, c.notes
FROM cargo c
LEFT JOIN stations origin ON origin.id = c.origin_station_id
JOIN stations destination ON destination.id = c.destination_station_id
LEFT JOIN expeditions e ON e.id = c.expedition_id
LEFT JOIN logistics_batches b ON b.id = c.logistics_batch_id
LEFT JOIN logistics_shipments shipment ON shipment.id = c.logistics_shipment_id
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

-- name: GetLogisticsExpeditionWindow :one
SELECT planned_start_at, planned_end_at
FROM expeditions
WHERE id = sqlc.arg(expedition_id);

-- name: GetLogisticsShipmentWindow :one
SELECT e.planned_start_at, e.planned_end_at
FROM logistics_batches b
JOIN expeditions e ON e.id = b.expedition_id
WHERE b.id = sqlc.arg(logistics_batch_id)
  AND b.status NOT IN ('received', 'cancelled');

-- name: CreateDefaultLogisticsShipment :one
INSERT INTO logistics_shipments (
    shipment_code, logistics_batch_id, origin_station_id, destination_station_id,
    transport_mode, priority, status, is_default_route, planned_dispatch_at,
    estimated_arrival_at, notes
)
SELECT b.batch_code || '-DEFAULT', b.id, b.origin_station_id, b.destination_station_id,
       'ground', 'standard', b.status, TRUE, b.planned_dispatch_at,
       b.estimated_arrival_at, 'Default route for the parent manifest.'
FROM logistics_batches b
WHERE b.id = sqlc.arg(logistics_batch_id)
RETURNING *;

-- name: CreateLogisticsShipment :one
INSERT INTO logistics_shipments (
    shipment_code, logistics_batch_id, origin_station_id, destination_station_id,
    transport_mode, priority, status, planned_dispatch_at, estimated_arrival_at, notes
)
SELECT sqlc.arg(shipment_code), b.id, b.origin_station_id, b.destination_station_id,
       sqlc.arg(transport_mode), sqlc.arg(priority), sqlc.arg(status),
       sqlc.arg(planned_dispatch_at), sqlc.arg(estimated_arrival_at), sqlc.narg(notes)
FROM logistics_batches b
WHERE b.id = sqlc.arg(logistics_batch_id)
  AND b.status NOT IN ('received', 'cancelled')
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
SET logistics_batch_id = b.id, logistics_shipment_id = shipment.id, expedition_id = b.expedition_id,
    origin_station_id = b.origin_station_id, destination_station_id = b.destination_station_id
FROM logistics_batches b
JOIN logistics_shipments shipment ON shipment.logistics_batch_id = b.id AND shipment.is_default_route = TRUE
WHERE c.id = sqlc.arg(cargo_id) AND b.id = sqlc.arg(logistics_batch_id)
  AND b.status NOT IN ('received', 'cancelled')
RETURNING c.*;

-- name: AssignCargoToLogisticsShipment :one
UPDATE cargo c
SET logistics_shipment_id = shipment.id,
    logistics_batch_id = shipment.logistics_batch_id,
    origin_station_id = shipment.origin_station_id,
    destination_station_id = shipment.destination_station_id
FROM logistics_shipments shipment
WHERE c.id = sqlc.arg(cargo_id)
  AND shipment.id = sqlc.arg(logistics_shipment_id)
  AND shipment.status NOT IN ('received', 'cancelled')
  AND (c.logistics_batch_id IS NULL OR c.logistics_batch_id = shipment.logistics_batch_id)
RETURNING c.*;

-- name: ClearCargoLogisticsShipment :one
UPDATE cargo
SET logistics_shipment_id = NULL
WHERE id = sqlc.arg(cargo_id)
RETURNING *;

-- name: SplitCargoManifest :one
WITH source AS (
    SELECT c.*
    FROM cargo c
    WHERE c.id = sqlc.arg(source_cargo_id)
      AND c.status NOT IN ('received', 'cancelled')
    FOR UPDATE
), requested AS (
    SELECT (entry ->> 'item_id')::UUID AS item_id,
           (entry ->> 'quantity')::NUMERIC(14, 3) AS quantity
    FROM jsonb_array_elements(sqlc.arg(items_json)::JSONB) AS entry
), valid_items AS (
    SELECT ci.id AS cargo_item_id, ci.item_id, ci.quantity AS source_quantity,
           requested.quantity
    FROM cargo_items ci
    JOIN requested ON requested.item_id = ci.item_id
    JOIN source ON source.id = ci.cargo_id
    WHERE requested.quantity > 0
      AND requested.quantity <= ci.quantity
    FOR UPDATE OF ci
), valid_request AS (
    SELECT COUNT(*) = (SELECT COUNT(*) FROM requested)
               AND COUNT(*) > 0 AS valid
    FROM valid_items
), target_shipment AS (
    SELECT shipment.id
    FROM logistics_shipments shipment
    JOIN source ON source.logistics_batch_id = shipment.logistics_batch_id
    WHERE shipment.id = COALESCE(sqlc.narg(logistics_shipment_id), source.logistics_shipment_id)
      AND shipment.status NOT IN ('received', 'cancelled')
), child AS (
    INSERT INTO cargo (
        cargo_code, origin_station_id, destination_station_id, expedition_id,
        logistics_batch_id, logistics_shipment_id, source_cargo_id, priority, status, notes
    )
    SELECT sqlc.arg(new_cargo_code), source.origin_station_id, source.destination_station_id,
           source.expedition_id, source.logistics_batch_id, target_shipment.id, source.id,
           sqlc.arg(priority), 'draft', sqlc.narg(notes)
    FROM source CROSS JOIN target_shipment CROSS JOIN valid_request
    WHERE valid_request.valid
    RETURNING id
), reduced AS (
    UPDATE cargo_items ci
    SET quantity = ci.quantity - valid_items.quantity
    FROM valid_items
    WHERE ci.id = valid_items.cargo_item_id
      AND EXISTS (SELECT 1 FROM child)
    RETURNING ci.id
), child_items AS (
    INSERT INTO cargo_items (cargo_id, item_id, quantity, declared_weight_kg, notes)
    SELECT child.id, valid_items.item_id, valid_items.quantity,
           CASE WHEN valid_items.source_quantity = 0 THEN NULL
                ELSE source_item.declared_weight_kg * valid_items.quantity / valid_items.source_quantity END,
           source_item.notes
    FROM child
    JOIN valid_items ON TRUE
    JOIN cargo_items source_item ON source_item.id = valid_items.cargo_item_id
    RETURNING id
)
SELECT child.id
FROM child;

-- name: GetCargoForQR :one
SELECT c.id,
       c.cargo_code,
       COALESCE(origin.name, '') AS origin_station_name,
       destination.name AS destination_station_name,
       COALESCE(e.expedition_code, '') AS expedition_code,
       COALESCE(e.name, '') AS expedition_name,
       c.logistics_batch_id,
       COALESCE(b.batch_code, '') AS logistics_batch_code,
       c.logistics_shipment_id,
       COALESCE(shipment.shipment_code, '') AS logistics_shipment_code,
       COALESCE(shipment.transport_mode, '') AS transport_mode,
       c.source_cargo_id,
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
LEFT JOIN logistics_shipments shipment ON shipment.id = c.logistics_shipment_id
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

-- name: RecordCargoQRCheckpoint :one
WITH source AS (
    SELECT logistics_batch_id, logistics_shipment_id
    FROM cargo
    WHERE id = sqlc.arg(cargo_id)
), requested_update AS (
    SELECT sqlc.narg(status)::TEXT AS status
), targets AS (
    SELECT c.id
    FROM cargo c
    CROSS JOIN source s
    WHERE (sqlc.arg(propagation_scope)::TEXT = 'cargo' AND c.id = sqlc.arg(cargo_id))
       OR (sqlc.arg(propagation_scope)::TEXT = 'shipment' AND c.logistics_shipment_id = s.logistics_shipment_id)
       OR (sqlc.arg(propagation_scope)::TEXT = 'batch' AND c.logistics_batch_id = s.logistics_batch_id)
), updated AS (
    UPDATE cargo c
    SET status = COALESCE(requested_update.status, c.status),
        dispatched_at = CASE
            WHEN requested_update.status IN ('dispatched', 'in_transit') THEN COALESCE(c.dispatched_at, NOW())
            ELSE c.dispatched_at
        END,
        received_at = CASE
            WHEN requested_update.status = 'received' THEN COALESCE(c.received_at, NOW())
            ELSE c.received_at
        END
    FROM requested_update
    WHERE c.id IN (SELECT id FROM targets)
      AND requested_update.status IS NOT NULL
      AND c.status NOT IN ('received', 'cancelled')
    RETURNING c.id
), scans AS (
    INSERT INTO cargo_qr_scans (
        cargo_id, event_type, station_id, scanned_by_personnel_id,
        latitude, longitude, notes
    )
    SELECT target.id, sqlc.arg(event_type), sqlc.narg(station_id),
           sqlc.narg(scanned_by_personnel_id), sqlc.arg(latitude)::DOUBLE PRECISION,
           sqlc.arg(longitude)::DOUBLE PRECISION, sqlc.narg(notes)
    FROM targets target
    RETURNING id, cargo_id, scanned_at
), updated_shipments AS (
    UPDATE logistics_shipments shipment
    SET status = requested_update.status,
        dispatched_at = CASE WHEN requested_update.status IN ('dispatched', 'in_transit') THEN COALESCE(shipment.dispatched_at, NOW()) ELSE shipment.dispatched_at END,
        received_at = CASE WHEN requested_update.status = 'received' THEN COALESCE(shipment.received_at, NOW()) ELSE shipment.received_at END
    FROM requested_update
    WHERE requested_update.status IS NOT NULL
      AND shipment.id IN (SELECT DISTINCT c.logistics_shipment_id FROM cargo c JOIN targets t ON t.id = c.id WHERE c.logistics_shipment_id IS NOT NULL)
      AND shipment.status NOT IN ('received', 'cancelled')
    RETURNING shipment.id
)
SELECT (SELECT id FROM scans WHERE cargo_id = sqlc.arg(cargo_id) LIMIT 1) AS scan_id,
       (SELECT scanned_at FROM scans WHERE cargo_id = sqlc.arg(cargo_id) LIMIT 1) AS scanned_at,
       (SELECT COUNT(*)::INTEGER FROM scans) AS affected_cargo_count,
       (SELECT COUNT(*)::INTEGER FROM updated) AS status_updated_count;

-- name: GetLogisticsBatchTracking :one
SELECT b.id,
       b.batch_code,
       b.status,
       e.expedition_code,
       e.name AS expedition_name,
       origin.name AS origin_station_name,
       destination.name AS destination_station_name,
       COUNT(c.id)::INTEGER AS cargo_count
FROM logistics_batches b
JOIN expeditions e ON e.id = b.expedition_id
JOIN stations origin ON origin.id = b.origin_station_id
JOIN stations destination ON destination.id = b.destination_station_id
LEFT JOIN cargo c ON c.logistics_batch_id = b.id
WHERE b.id = sqlc.arg(logistics_batch_id)
GROUP BY b.id, e.expedition_code, e.name, origin.name, destination.name;

-- name: ListLogisticsBatchCheckpointPoints :many
SELECT DISTINCT ON (scan.scanned_at, scan.latitude, scan.longitude, scan.event_type)
       scan.id,
       scan.event_type,
       scan.latitude::DOUBLE PRECISION AS latitude,
       scan.longitude::DOUBLE PRECISION AS longitude,
       scan.scanned_at,
       scan.notes,
       COUNT(*) OVER (
           PARTITION BY scan.scanned_at, scan.latitude, scan.longitude, scan.event_type
       )::INTEGER AS affected_cargo_count
FROM cargo_qr_scans scan
JOIN cargo c ON c.id = scan.cargo_id
WHERE c.logistics_batch_id = sqlc.arg(logistics_batch_id)
  AND scan.latitude IS NOT NULL
  AND scan.longitude IS NOT NULL
ORDER BY scan.scanned_at ASC, scan.latitude, scan.longitude, scan.event_type;

-- name: ListLogisticsShipmentsForBatch :many
SELECT shipment.id,
       shipment.shipment_code,
       shipment.transport_mode,
       shipment.priority,
       shipment.status,
       shipment.planned_dispatch_at,
       shipment.estimated_arrival_at,
       COUNT(c.id)::INTEGER AS cargo_count,
       COALESCE(SUM(item_totals.quantity), 0)::DOUBLE PRECISION AS total_quantity
FROM logistics_shipments shipment
LEFT JOIN cargo c ON c.logistics_shipment_id = shipment.id
LEFT JOIN LATERAL (
    SELECT SUM(ci.quantity) AS quantity
    FROM cargo_items ci
    WHERE ci.cargo_id = c.id
) item_totals ON TRUE
WHERE shipment.logistics_batch_id = sqlc.arg(logistics_batch_id)
GROUP BY shipment.id
ORDER BY shipment.is_default_route DESC, shipment.planned_dispatch_at ASC NULLS LAST, shipment.created_at ASC;

-- name: ListLogisticsShipmentCheckpointPoints :many
SELECT DISTINCT ON (scan.scanned_at, scan.latitude, scan.longitude, scan.event_type, scan.cargo_id)
       scan.id,
       scan.cargo_id,
       c.cargo_code,
       scan.event_type,
       scan.latitude::DOUBLE PRECISION AS latitude,
       scan.longitude::DOUBLE PRECISION AS longitude,
       scan.scanned_at,
       scan.notes
FROM cargo_qr_scans scan
JOIN cargo c ON c.id = scan.cargo_id
WHERE c.logistics_shipment_id = sqlc.arg(logistics_shipment_id)
  AND scan.latitude IS NOT NULL
  AND scan.longitude IS NOT NULL
ORDER BY scan.scanned_at ASC, scan.latitude, scan.longitude, scan.event_type, scan.cargo_id;
