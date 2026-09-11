-- Idempotent demo data for Bharti inventory and the inventory analytics UI.
-- Run manually after migrations: psql "$env:DB_URL" -f sql/seeds/001_inventory_ai_demo.sql

INSERT INTO items (item_code, name, category, unit, is_critical)
VALUES
    ('DEMO-AI-FUEL', 'Aircraft fuel (demo drums)', 'fuel', 'drums', TRUE),
    ('DEMO-AI-OXYGEN', 'Medical oxygen (demo tanks)', 'medical', 'tanks', TRUE),
    ('DEMO-AI-FLUID', 'Vehicle hydraulic fluid (demo)', 'maintenance', 'litres', FALSE),
    ('DEMO-AI-RATIONS', 'Food ration packs (demo)', 'food', 'packs', FALSE)
ON CONFLICT (item_code) DO NOTHING;

INSERT INTO station_inventory (station_id, item_id, on_hand_quantity, minimum_quantity, maximum_quantity)
SELECT s.id,
       i.id,
       CASE i.item_code
           WHEN 'DEMO-AI-FUEL' THEN 14
           WHEN 'DEMO-AI-OXYGEN' THEN 12
           WHEN 'DEMO-AI-FLUID' THEN 180
           WHEN 'DEMO-AI-RATIONS' THEN 450
       END,
       CASE i.item_code
           WHEN 'DEMO-AI-FUEL' THEN 50
           WHEN 'DEMO-AI-OXYGEN' THEN 20
           WHEN 'DEMO-AI-FLUID' THEN 200
           WHEN 'DEMO-AI-RATIONS' THEN 200
       END,
       CASE i.item_code
           WHEN 'DEMO-AI-FUEL' THEN 200
           WHEN 'DEMO-AI-OXYGEN' THEN 100
           WHEN 'DEMO-AI-FLUID' THEN 600
           WHEN 'DEMO-AI-RATIONS' THEN 1000
       END
FROM stations s
JOIN items i ON i.item_code IN ('DEMO-AI-FUEL', 'DEMO-AI-OXYGEN', 'DEMO-AI-FLUID', 'DEMO-AI-RATIONS')
WHERE s.code = 'BHARTI'
ON CONFLICT (station_id, item_id) DO NOTHING;

INSERT INTO inventory_transactions (station_inventory_id, transaction_type, on_hand_delta, occurred_at, notes)
SELECT si.id,
       'receipt',
       CASE i.item_code
           WHEN 'DEMO-AI-FUEL' THEN 50
           WHEN 'DEMO-AI-OXYGEN' THEN 42
           WHEN 'DEMO-AI-FLUID' THEN 360
           WHEN 'DEMO-AI-RATIONS' THEN 600
       END,
       NOW() - INTERVAL '14 days',
       'demo-ai-seed-opening-receipt'
FROM station_inventory si
JOIN stations s ON s.id = si.station_id
JOIN items i ON i.id = si.item_id
WHERE s.code = 'BHARTI'
  AND i.item_code IN ('DEMO-AI-FUEL', 'DEMO-AI-OXYGEN', 'DEMO-AI-FLUID', 'DEMO-AI-RATIONS')
  AND NOT EXISTS (
      SELECT 1 FROM inventory_transactions it
      WHERE it.station_inventory_id = si.id AND it.notes = 'demo-ai-seed-opening-receipt'
  );

-- Fuel contains one unusually high consumption day; it is useful once the anomaly service
-- accepts Drishti transaction data. The remaining series produce predictable stock-days values.
INSERT INTO inventory_transactions (station_inventory_id, transaction_type, on_hand_delta, occurred_at, notes)
SELECT si.id,
       'consumption',
       CASE
           WHEN i.item_code = 'DEMO-AI-FUEL' AND day_number = 7 THEN -10
           WHEN i.item_code = 'DEMO-AI-FUEL' THEN -2
           WHEN i.item_code = 'DEMO-AI-OXYGEN' THEN -3
           WHEN i.item_code = 'DEMO-AI-FLUID' THEN -12
           WHEN i.item_code = 'DEMO-AI-RATIONS' THEN -15
       END,
       NOW() - (15 - day_number) * INTERVAL '1 day',
       'demo-ai-seed-consumption'
FROM station_inventory si
JOIN stations s ON s.id = si.station_id
JOIN items i ON i.id = si.item_id
CROSS JOIN generate_series(1, 14) AS day_number
WHERE s.code = 'BHARTI'
  AND i.item_code IN ('DEMO-AI-FUEL', 'DEMO-AI-OXYGEN', 'DEMO-AI-FLUID', 'DEMO-AI-RATIONS')
  AND NOT EXISTS (
      SELECT 1 FROM inventory_transactions it
      WHERE it.station_inventory_id = si.id AND it.notes = 'demo-ai-seed-consumption'
  );
