-- name: CreateExpedition :one
INSERT INTO expeditions (
    expedition_code,
    name,
    purpose,
    origin_station_id,
    destination_station_id,
    lead_personnel_id,
    planned_start_at,
    planned_end_at,
    status
)
VALUES (
    sqlc.arg(expedition_code),
    sqlc.arg(name),
    sqlc.narg(purpose),
    sqlc.arg(origin_station_id),
    sqlc.arg(destination_station_id),
    sqlc.arg(lead_personnel_id),
    sqlc.arg(planned_start_at),
    sqlc.arg(planned_end_at),
    sqlc.arg(status)
)
RETURNING *;

-- name: ListExpeditions :many
SELECT e.id,
       e.name AS expedition_name,
       e.expedition_code,
       COALESCE(origin.name, '') AS origin_station_name,
       COALESCE(destination.name, '') AS destination_station_name,
       COALESCE(leader.full_name, '') AS team_lead_name,
       COUNT(member.personnel_id)::INTEGER AS crew_assigned,
       e.planned_start_at AS start_date,
       e.planned_end_at AS end_date,
       e.status
FROM expeditions e
         LEFT JOIN stations origin ON origin.id = e.origin_station_id
         LEFT JOIN stations destination ON destination.id = e.destination_station_id
         LEFT JOIN personnel leader ON leader.id = e.lead_personnel_id
         LEFT JOIN expedition_members member ON member.expedition_id = e.id
             AND member.released_at IS NULL
GROUP BY e.id, origin.name, destination.name, leader.full_name
ORDER BY e.planned_start_at DESC NULLS LAST, e.created_at DESC;

-- name: UpdateExpedition :one
UPDATE expeditions
SET expedition_code        = sqlc.arg(expedition_code),
    name                   = sqlc.arg(name),
    purpose                = sqlc.narg(purpose),
    origin_station_id      = sqlc.arg(origin_station_id),
    destination_station_id = sqlc.arg(destination_station_id),
    lead_personnel_id      = sqlc.arg(lead_personnel_id),
    planned_start_at       = sqlc.arg(planned_start_at),
    planned_end_at         = sqlc.arg(planned_end_at),
    status                 = sqlc.arg(status)
WHERE id = sqlc.arg(id)
RETURNING *;
