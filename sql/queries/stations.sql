-- name: ListStations :many
SELECT id,
       code,
       name,
       station_type,
       latitude,
       longitude,
       personnel_capacity,
       notes,
       created_at,
       updated_at
FROM stations
ORDER BY name ASC;