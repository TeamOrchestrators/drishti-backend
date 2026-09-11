-- name: ListActivePersonnel :many
SELECT id,
       personnel_code,
       full_name,
       role
FROM personnel
WHERE status <> 'inactive'
ORDER BY full_name ASC;

-- name: ListMovingPersonnel :many
SELECT p.id AS personnel_id,
       p.full_name AS name,
       p.role,
       COALESCE(current_station.name, '') AS current_station,
       COALESCE(origin.name, '') AS origin_station_name,
       COALESCE(destination.name, '') AS destination_station_name,
       m.departed_at AS departure_time,
       m.estimated_arrival_at AS arrival_time,
       m.status
FROM personnel_movements m
JOIN personnel p ON p.id = m.personnel_id
LEFT JOIN stations current_station ON current_station.id = p.current_station_id
LEFT JOIN stations origin ON origin.id = m.origin_station_id
LEFT JOIN stations destination ON destination.id = m.destination_station_id
WHERE m.status IN ('planned', 'in_transit')
ORDER BY m.departed_at DESC NULLS LAST, m.created_at DESC;

-- name: ListPersonnel :many
SELECT p.id AS personnel_id,
       p.full_name AS name,
       p.role,
       COALESCE(s.name, '') AS current_station,
       p.status
FROM personnel p
LEFT JOIN stations s ON s.id = p.current_station_id
ORDER BY p.full_name ASC;

-- name: ListMovementHistory :many
SELECT p.id AS personnel_id,
       p.full_name AS name,
       COALESCE(origin.name, '') AS origin_station_name,
       COALESCE(destination.name, '') AS destination_station_name,
       m.departed_at AS departure_time,
       m.arrived_at AS arrival_time,
       m.status
FROM personnel_movements m
JOIN personnel p ON p.id = m.personnel_id
LEFT JOIN stations origin ON origin.id = m.origin_station_id
LEFT JOIN stations destination ON destination.id = m.destination_station_id
WHERE m.status = 'arrived'
ORDER BY m.arrived_at DESC NULLS LAST, m.created_at DESC;

-- name: ListAssignableExpeditions :many
SELECT e.id,
       e.name,
       COALESCE(origin.name, '') AS origin_station_name,
       COALESCE(destination.name, '') AS destination_station_name
FROM expeditions e
LEFT JOIN stations origin ON origin.id = e.origin_station_id
LEFT JOIN stations destination ON destination.id = e.destination_station_id
WHERE e.status NOT IN ('completed', 'cancelled')
ORDER BY e.planned_start_at ASC NULLS LAST, e.name ASC;

-- name: ListAssignablePersonnel :many
SELECT p.id,
       p.full_name AS name,
       p.role,
       COALESCE(s.name, '') AS current_station
FROM personnel p
LEFT JOIN stations s ON s.id = p.current_station_id
WHERE p.status = 'available'
  AND p.medical_clearance_status = 'cleared'
ORDER BY p.full_name ASC;

-- name: AssignPersonnelToExpedition :one
WITH valid_assignment AS (
    SELECT p.id AS personnel_id,
           origin.code AS origin_station_code,
           destination.code AS destination_station_code
    FROM expeditions e
    JOIN personnel p ON p.id = sqlc.arg(personnel_id)
    JOIN stations origin ON origin.id = sqlc.arg(origin_station_id)
    JOIN stations destination ON destination.id = sqlc.arg(destination_station_id)
    WHERE e.id = sqlc.arg(expedition_id)
      AND e.status NOT IN ('completed', 'cancelled')
      AND e.planned_start_at IS NOT NULL
      AND e.planned_end_at IS NOT NULL
      AND e.planned_start_at <= sqlc.arg(departure_time)::TIMESTAMPTZ
      AND e.planned_end_at >= sqlc.arg(arrival_time)::TIMESTAMPTZ
      AND p.status = 'available'
      AND p.medical_clearance_status = 'cleared'
      AND p.current_station_id = sqlc.arg(origin_station_id)
      AND NOT EXISTS (
          SELECT 1
          FROM personnel_movements m
          WHERE m.personnel_id = p.id
            AND m.status IN ('planned', 'in_transit')
      )
    FOR UPDATE OF e, p
), membership AS (
    INSERT INTO expedition_members (expedition_id, personnel_id, released_at)
    SELECT sqlc.arg(expedition_id),
           personnel_id,
           CASE WHEN sqlc.arg(movement_status)::TEXT = 'cancelled' THEN NOW() ELSE NULL END
    FROM valid_assignment
    ON CONFLICT (expedition_id, personnel_id) DO UPDATE
        SET released_at = EXCLUDED.released_at
    RETURNING personnel_id
), movement AS (
    INSERT INTO personnel_movements (
        personnel_id,
        expedition_id,
        movement_type,
        origin_station_id,
        destination_station_id,
        status,
        departed_at,
        estimated_arrival_at,
        arrived_at
    )
    SELECT membership.personnel_id,
           sqlc.arg(expedition_id),
           CASE
               WHEN valid_assignment.origin_station_code = 'INDIA-HQ' THEN 'deployment'
               WHEN valid_assignment.destination_station_code = 'INDIA-HQ' THEN 'return_to_nation'
               ELSE 'station_transfer'
           END,
           sqlc.arg(origin_station_id),
           sqlc.arg(destination_station_id),
           sqlc.arg(movement_status)::TEXT,
           sqlc.arg(departure_time)::TIMESTAMPTZ,
           sqlc.arg(arrival_time)::TIMESTAMPTZ,
           CASE
               WHEN sqlc.arg(movement_status)::TEXT = 'arrived' THEN sqlc.arg(arrival_time)::TIMESTAMPTZ
               ELSE NULL::TIMESTAMPTZ
           END
    FROM membership
    CROSS JOIN valid_assignment
    RETURNING id
), updated_personnel AS (
    UPDATE personnel
    SET status = CASE
                     WHEN sqlc.arg(movement_status) = 'in_transit' THEN 'in_transit'
                     WHEN sqlc.arg(movement_status) = 'cancelled' THEN 'available'
                     ELSE 'assigned'
                 END,
        current_station_id = CASE
                                 WHEN sqlc.arg(movement_status) = 'arrived' THEN sqlc.arg(destination_station_id)
                                 ELSE current_station_id
                             END
    WHERE id = sqlc.arg(personnel_id)
    RETURNING id
)
SELECT movement.id
FROM movement
CROSS JOIN updated_personnel;
