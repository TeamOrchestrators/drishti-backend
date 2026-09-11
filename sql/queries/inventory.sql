-- name: ListStationInventory :many
SELECT si.id,
       i.id AS item_id,
       i.item_code,
       i.name,
       i.category,
       i.unit,
       i.is_critical,
       si.on_hand_quantity::DOUBLE PRECISION AS on_hand_quantity,
       si.reserved_quantity::DOUBLE PRECISION AS reserved_quantity,
       si.available_quantity::DOUBLE PRECISION AS available_quantity,
       si.minimum_quantity::DOUBLE PRECISION AS minimum_quantity,
       COALESCE(si.maximum_quantity, 0)::DOUBLE PRECISION AS maximum_quantity
FROM station_inventory si
JOIN items i ON i.id = si.item_id
WHERE si.station_id = sqlc.arg(station_id)
ORDER BY i.is_critical DESC, i.name ASC;

-- name: ListStationInventoryHistory :many
SELECT it.id,
       i.name AS item_name,
       i.unit,
       it.transaction_type,
       it.on_hand_delta::DOUBLE PRECISION AS on_hand_delta,
       it.reserved_delta::DOUBLE PRECISION AS reserved_delta,
       it.occurred_at,
       it.notes
FROM inventory_transactions it
JOIN station_inventory si ON si.id = it.station_inventory_id
JOIN items i ON i.id = si.item_id
WHERE si.station_id = sqlc.arg(station_id)
ORDER BY it.occurred_at DESC
LIMIT sqlc.arg(history_limit);

-- name: ListInventoryTransactionsForAnalytics :many
SELECT transaction_type,
       on_hand_delta::DOUBLE PRECISION AS quantity_delta,
       occurred_at
FROM inventory_transactions
WHERE station_inventory_id = sqlc.arg(station_inventory_id)
ORDER BY occurred_at ASC;

-- name: CreateStationInventoryItem :one
WITH new_item AS (
    INSERT INTO items (item_code, name, category, unit, is_critical)
    VALUES (sqlc.arg(item_code), sqlc.arg(name), sqlc.arg(category), sqlc.arg(unit), sqlc.arg(is_critical))
    RETURNING id
), new_inventory AS (
    INSERT INTO station_inventory (station_id, item_id, on_hand_quantity, minimum_quantity, maximum_quantity)
    SELECT sqlc.arg(station_id), id, sqlc.arg(opening_quantity)::NUMERIC,
           sqlc.arg(minimum_quantity)::NUMERIC, sqlc.narg(maximum_quantity)::NUMERIC
    FROM new_item
    RETURNING id
), opening_transaction AS (
    INSERT INTO inventory_transactions (station_inventory_id, transaction_type, on_hand_delta, notes)
    SELECT id, 'receipt', sqlc.arg(opening_quantity)::NUMERIC, sqlc.narg(notes)
    FROM new_inventory
    WHERE sqlc.arg(opening_quantity)::NUMERIC > 0
)
SELECT id FROM new_inventory;

-- name: AdjustStationInventoryStock :one
WITH updated_inventory AS (
    UPDATE station_inventory
    SET on_hand_quantity = on_hand_quantity + sqlc.arg(on_hand_delta)::NUMERIC
    WHERE station_inventory.id = sqlc.arg(station_inventory_id)
      AND station_inventory.station_id = sqlc.arg(station_id)
      AND station_inventory.on_hand_quantity + sqlc.arg(on_hand_delta)::NUMERIC >= 0
    RETURNING id
), transaction AS (
    INSERT INTO inventory_transactions (station_inventory_id, transaction_type, on_hand_delta, notes)
    SELECT id, sqlc.arg(transaction_type), sqlc.arg(on_hand_delta)::NUMERIC, sqlc.narg(notes)
    FROM updated_inventory
    RETURNING id
)
SELECT id FROM transaction;
