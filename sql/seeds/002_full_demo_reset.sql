-- Full operational-data reset for the hackathon demo.
-- Preserves only stations and personnel records; all other operational data is replaced.

TRUNCATE TABLE
    emergency_timeline_events,
    emergency_resource_allocations,
    emergency_resource_requests,
    emergency_personnel,
    emergency_signals,
    emergency_sos_confirmations,
    emergencies,
    emergency_devices,
    cargo_qr_scans,
    cargo_items,
    cargo,
    logistics_batches,
    expedition_resource_allocations,
    expedition_resource_requirements,
    expedition_members,
    personnel_movements,
    expeditions,
    inventory_transactions,
    station_inventory,
    items;

-- Ensure the standard demo roster exists while preserving any additional personnel.
INSERT INTO personnel (personnel_code, full_name, role, medical_clearance_status, current_station_id, status)
VALUES
    ('PERS-001', 'Dr. Ananya Rao', 'Station Commander', 'cleared', (SELECT id FROM stations WHERE code = 'BHARTI'), 'available'),
    ('PERS-002', 'Dr. Arjun Mehta', 'Glaciologist', 'cleared', (SELECT id FROM stations WHERE code = 'BHARTI'), 'available'),
    ('PERS-003', 'Kavya Nair', 'Logistics Officer', 'cleared', (SELECT id FROM stations WHERE code = 'BHARTI'), 'available'),
    ('PERS-004', 'Dr. Rohan Iyer', 'Station Commander', 'cleared', (SELECT id FROM stations WHERE code = 'MAITRI'), 'available'),
    ('PERS-005', 'Neel Sharma', 'Communications Engineer', 'cleared', (SELECT id FROM stations WHERE code = 'MAITRI'), 'available')
ON CONFLICT (personnel_code) DO UPDATE
SET full_name = EXCLUDED.full_name,
    role = EXCLUDED.role,
    medical_clearance_status = EXCLUDED.medical_clearance_status;

-- Reset standard-demo personnel operational state for this scenario.
UPDATE personnel p
SET current_station_id = CASE p.personnel_code
        WHEN 'PERS-002' THEN (SELECT id FROM stations WHERE code = 'INDIA-HQ')
        WHEN 'PERS-004' THEN (SELECT id FROM stations WHERE code = 'MAITRI')
        WHEN 'PERS-005' THEN (SELECT id FROM stations WHERE code = 'MAITRI')
        ELSE (SELECT id FROM stations WHERE code = 'BHARTI')
    END,
    status = CASE p.personnel_code
        WHEN 'PERS-001' THEN 'assigned'
        WHEN 'PERS-002' THEN 'in_transit'
        WHEN 'PERS-003' THEN 'assigned'
        ELSE 'available'
    END
WHERE p.personnel_code IN ('PERS-001', 'PERS-002', 'PERS-003', 'PERS-004', 'PERS-005');

-- Keep the two additional test personnel available at their intended stations.
UPDATE personnel p
SET medical_clearance_status = 'cleared',
    status = 'available',
    current_station_id = CASE p.personnel_code
        WHEN 'PERS-006' THEN (SELECT id FROM stations WHERE code = 'BHARTI')
        WHEN 'PERS-007' THEN (SELECT id FROM stations WHERE code = 'MAITRI')
    END
WHERE p.personnel_code IN ('PERS-006', 'PERS-007');

INSERT INTO expeditions (
    expedition_code, name, purpose, origin_station_id, destination_station_id,
    lead_personnel_id, planned_start_at, planned_end_at, actual_start_at, status
)
SELECT 'EXP-ANT-001', 'Maitri Ice Shelf Survey',
       'Field survey and ice-core sample collection between Bharti and Maitri.',
       (SELECT id FROM stations WHERE code = 'BHARTI'),
       (SELECT id FROM stations WHERE code = 'MAITRI'),
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-001'),
       NOW() - INTERVAL '2 days', NOW() + INTERVAL '5 days', NOW() - INTERVAL '1 day', 'active'
UNION ALL
SELECT 'EXP-ANT-002', 'Bharti Resupply Deployment',
       'Personnel and priority logistics deployment from India HQ to Bharti.',
       (SELECT id FROM stations WHERE code = 'INDIA-HQ'),
       (SELECT id FROM stations WHERE code = 'BHARTI'),
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-003'),
       NOW() - INTERVAL '1 day', NOW() + INTERVAL '4 days', NULL, 'ready';

INSERT INTO expedition_members (expedition_id, personnel_id, assignment_role)
SELECT e.id, p.id,
       CASE p.personnel_code WHEN 'PERS-001' THEN 'Field lead' ELSE 'Logistics coordinator' END
FROM expeditions e
JOIN personnel p ON p.personnel_code IN ('PERS-001', 'PERS-003')
WHERE e.expedition_code = 'EXP-ANT-001';

INSERT INTO personnel_movements (
    personnel_id, expedition_id, movement_type, origin_station_id, destination_station_id,
    status, departed_at, estimated_arrival_at, arrived_at, notes
)
SELECT (SELECT id FROM personnel WHERE personnel_code = 'PERS-002'),
       (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-002'),
       'deployment',
       (SELECT id FROM stations WHERE code = 'INDIA-HQ'),
       (SELECT id FROM stations WHERE code = 'BHARTI'),
       'in_transit', NOW() - INTERVAL '8 hours', NOW() + INTERVAL '16 hours', NULL,
       'Scheduled deployment to Bharti.'
UNION ALL
SELECT (SELECT id FROM personnel WHERE personnel_code = 'PERS-004'), NULL,
       'deployment',
       (SELECT id FROM stations WHERE code = 'INDIA-HQ'),
       (SELECT id FROM stations WHERE code = 'MAITRI'),
       'arrived', NOW() - INTERVAL '14 days', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days',
       'Completed deployment to Maitri.';

INSERT INTO items (item_code, name, category, unit, is_critical)
VALUES
    ('FUEL-ANT-001', 'Aircraft fuel drums', 'fuel', 'drums', TRUE),
    ('OXY-ANT-001', 'Medical oxygen tanks', 'medical', 'tanks', TRUE),
    ('HYD-ANT-001', 'Vehicle hydraulic fluid', 'maintenance', 'litres', FALSE),
    ('RAT-ANT-001', 'Food ration packs', 'food', 'packs', FALSE),
    ('GEN-ANT-001', 'Generator maintenance kits', 'equipment', 'kits', TRUE),
    ('MED-ANT-001', 'Field medical kits', 'medical', 'kits', TRUE);

INSERT INTO station_inventory (station_id, item_id, on_hand_quantity, minimum_quantity, maximum_quantity)
SELECT (SELECT id FROM stations WHERE code = 'BHARTI'), i.id,
       CASE i.item_code
           WHEN 'FUEL-ANT-001' THEN 142 WHEN 'OXY-ANT-001' THEN 18
           WHEN 'HYD-ANT-001' THEN 640 WHEN 'RAT-ANT-001' THEN 1840
           WHEN 'GEN-ANT-001' THEN 8 ELSE 36
       END,
       CASE i.item_code
           WHEN 'FUEL-ANT-001' THEN 400 WHEN 'OXY-ANT-001' THEN 50
           WHEN 'HYD-ANT-001' THEN 1200 WHEN 'RAT-ANT-001' THEN 900
           WHEN 'GEN-ANT-001' THEN 5 ELSE 40
       END,
       CASE i.item_code
           WHEN 'FUEL-ANT-001' THEN 900 WHEN 'OXY-ANT-001' THEN 120
           WHEN 'HYD-ANT-001' THEN 2000 WHEN 'RAT-ANT-001' THEN 2400
           WHEN 'GEN-ANT-001' THEN 30 ELSE 100
       END
FROM items i;

-- Opening receipts and fourteen days of consumption provide live history for stock-days.
INSERT INTO inventory_transactions (station_inventory_id, transaction_type, on_hand_delta, occurred_at, notes)
SELECT si.id, 'receipt',
       CASE i.item_code
           WHEN 'FUEL-ANT-001' THEN 359 WHEN 'OXY-ANT-001' THEN 60
           WHEN 'HYD-ANT-001' THEN 1004 WHEN 'RAT-ANT-001' THEN 2232
           WHEN 'GEN-ANT-001' THEN 15 ELSE 64
       END,
       NOW() - INTERVAL '15 days', 'Fresh-demo opening receipt'
FROM station_inventory si JOIN items i ON i.id = si.item_id;

INSERT INTO inventory_transactions (station_inventory_id, transaction_type, on_hand_delta, occurred_at, notes)
SELECT si.id, 'consumption',
       CASE
           WHEN i.item_code = 'FUEL-ANT-001' AND day_number = 7 THEN -48
           WHEN i.item_code = 'FUEL-ANT-001' THEN -13
           WHEN i.item_code = 'OXY-ANT-001' THEN -3
           WHEN i.item_code = 'HYD-ANT-001' THEN -26
           WHEN i.item_code = 'RAT-ANT-001' THEN -28
           WHEN i.item_code = 'GEN-ANT-001' THEN -0.5
           ELSE -2
       END,
       NOW() - (15 - day_number) * INTERVAL '1 day', 'Fresh-demo daily consumption'
FROM station_inventory si
JOIN items i ON i.id = si.item_id
CROSS JOIN generate_series(1, 14) AS day_number;

INSERT INTO expedition_resource_requirements (expedition_id, item_id, required_quantity, notes)
SELECT (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-001'), i.id,
       CASE i.item_code WHEN 'FUEL-ANT-001' THEN 40 WHEN 'OXY-ANT-001' THEN 10 ELSE 20 END,
       'Reserved for Maitri Ice Shelf Survey'
FROM items i WHERE i.item_code IN ('FUEL-ANT-001', 'OXY-ANT-001', 'RAT-ANT-001');

INSERT INTO logistics_batches (
    batch_code, expedition_id, origin_station_id, destination_station_id, status,
    planned_dispatch_at, estimated_arrival_at, dispatched_at, notes
)
SELECT 'BATCH-ANT-001', (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-001'),
       (SELECT id FROM stations WHERE code = 'BHARTI'), (SELECT id FROM stations WHERE code = 'MAITRI'),
       'dispatched', NOW() - INTERVAL '1 day', NOW() + INTERVAL '2 days', NOW() - INTERVAL '6 hours',
       'Priority field-support batch currently in transit.'
UNION ALL
SELECT 'BATCH-ANT-002', (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-002'),
       (SELECT id FROM stations WHERE code = 'INDIA-HQ'), (SELECT id FROM stations WHERE code = 'BHARTI'),
       'packed', NOW() + INTERVAL '12 hours', NOW() + INTERVAL '3 days', NULL,
       'Deployment batch ready for dispatch.';

INSERT INTO cargo (
    cargo_code, origin_station_id, destination_station_id, expedition_id, logistics_batch_id,
    priority, status, created_by_personnel_id, dispatched_at, notes
)
SELECT 'CG-ANT-001', (SELECT id FROM stations WHERE code = 'BHARTI'), (SELECT id FROM stations WHERE code = 'MAITRI'),
       (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-001'),
       (SELECT id FROM logistics_batches WHERE batch_code = 'BATCH-ANT-001'),
       'critical', 'in_transit', (SELECT id FROM personnel WHERE personnel_code = 'PERS-003'), NOW() - INTERVAL '6 hours',
       'Medical and fuel support for active field team.'
UNION ALL
SELECT 'CG-ANT-002', (SELECT id FROM stations WHERE code = 'BHARTI'), (SELECT id FROM stations WHERE code = 'MAITRI'),
       (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-001'),
       (SELECT id FROM logistics_batches WHERE batch_code = 'BATCH-ANT-001'),
       'high', 'in_transit', (SELECT id FROM personnel WHERE personnel_code = 'PERS-003'), NOW() - INTERVAL '6 hours',
       'Food and maintenance support.'
UNION ALL
SELECT 'CG-ANT-003', (SELECT id FROM stations WHERE code = 'INDIA-HQ'), (SELECT id FROM stations WHERE code = 'BHARTI'),
       (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-002'),
       (SELECT id FROM logistics_batches WHERE batch_code = 'BATCH-ANT-002'),
       'standard', 'packed', (SELECT id FROM personnel WHERE personnel_code = 'PERS-003'), NULL,
       'Bharti deployment equipment.';

INSERT INTO cargo_items (cargo_id, item_id, quantity, declared_weight_kg, notes)
SELECT c.id, i.id,
       CASE c.cargo_code
           WHEN 'CG-ANT-001' THEN CASE i.item_code WHEN 'OXY-ANT-001' THEN 6 ELSE 20 END
           WHEN 'CG-ANT-002' THEN CASE i.item_code WHEN 'RAT-ANT-001' THEN 120 ELSE 4 END
           ELSE CASE i.item_code WHEN 'MED-ANT-001' THEN 12 ELSE 3 END
       END,
       CASE c.cargo_code WHEN 'CG-ANT-001' THEN 280 WHEN 'CG-ANT-002' THEN 190 ELSE 75 END,
       'Fresh-demo cargo manifest'
FROM cargo c
JOIN items i ON (c.cargo_code = 'CG-ANT-001' AND i.item_code IN ('FUEL-ANT-001', 'OXY-ANT-001'))
             OR (c.cargo_code = 'CG-ANT-002' AND i.item_code IN ('RAT-ANT-001', 'GEN-ANT-001'))
             OR (c.cargo_code = 'CG-ANT-003' AND i.item_code IN ('MED-ANT-001', 'HYD-ANT-001'));

INSERT INTO cargo_qr_scans (cargo_id, event_type, station_id, scanned_by_personnel_id, latitude, longitude, scanned_at, notes)
SELECT c.id,
       CASE WHEN c.status = 'in_transit' THEN 'dispatched' ELSE 'packed' END,
       c.origin_station_id, (SELECT id FROM personnel WHERE personnel_code = 'PERS-003'),
       s.latitude, s.longitude, COALESCE(c.dispatched_at, NOW() - INTERVAL '2 hours'),
       'Fresh-demo QR scan event'
FROM cargo c JOIN stations s ON s.id = c.origin_station_id;
