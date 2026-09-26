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

-- A fresh database contains schema only. Ensure the station master data exists
-- before later seed statements resolve station IDs by code.
INSERT INTO stations (
    code, name, station_type, latitude, longitude, personnel_capacity, notes
)
VALUES
    ('BHARTI', 'Bharti Research Station', 'station', -69.408000, 76.192000, 47,
     'Indian Antarctic research station and primary operations base.'),
    ('MAITRI', 'Maitri Research Station', 'station', -70.766667, 11.733333, 65,
     'Indian Antarctic research station supporting inland operations.'),
    ('INDIA-HQ', 'India HQ', 'hub', 28.613900, 77.209000, 120,
     'India-based coordination and departure hub.')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    station_type = EXCLUDED.station_type,
    latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    personnel_capacity = EXCLUDED.personnel_capacity,
    notes = EXCLUDED.notes;

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
       NOW() - INTERVAL '1 day', NOW() + INTERVAL '4 days', NULL, 'ready'
UNION ALL
SELECT 'EXP-ANT-003', 'Maitri Winter Logistics Support',
       'Completed cross-station logistics support mission used for personnel-history demonstration.',
       (SELECT id FROM stations WHERE code = 'MAITRI'),
       (SELECT id FROM stations WHERE code = 'BHARTI'),
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-004'),
       NOW() - INTERVAL '50 days', NOW() - INTERVAL '36 days',
       NOW() - INTERVAL '50 days', 'completed';

INSERT INTO expedition_members (expedition_id, personnel_id, assignment_role)
SELECT e.id, p.id,
       CASE p.personnel_code WHEN 'PERS-001' THEN 'Field lead' ELSE 'Logistics coordinator' END
FROM expeditions e
JOIN personnel p ON p.personnel_code IN ('PERS-001', 'PERS-003')
WHERE e.expedition_code = 'EXP-ANT-001';

INSERT INTO expedition_members (expedition_id, personnel_id, assignment_role, assigned_at, released_at)
SELECT (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-003'),
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-004'),
       'Mission lead', NOW() - INTERVAL '50 days', NOW() - INTERVAL '36 days'
UNION ALL
SELECT (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-003'),
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-005'),
       'Communications support', NOW() - INTERVAL '50 days', NOW() - INTERVAL '36 days';

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
       'Completed deployment to Maitri.'
UNION ALL
SELECT (SELECT id FROM personnel WHERE personnel_code = 'PERS-004'),
       (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-003'),
       'station_transfer',
       (SELECT id FROM stations WHERE code = 'MAITRI'),
       (SELECT id FROM stations WHERE code = 'BHARTI'),
       'arrived', NOW() - INTERVAL '50 days', NOW() - INTERVAL '49 days', NOW() - INTERVAL '49 days',
       'Transferred to Bharti to lead winter logistics support.'
UNION ALL
SELECT (SELECT id FROM personnel WHERE personnel_code = 'PERS-004'),
       (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-003'),
       'station_transfer',
       (SELECT id FROM stations WHERE code = 'BHARTI'),
       (SELECT id FROM stations WHERE code = 'MAITRI'),
       'arrived', NOW() - INTERVAL '37 days', NOW() - INTERVAL '36 days', NOW() - INTERVAL '36 days',
       'Returned to Maitri after completed logistics mission.'
UNION ALL
SELECT (SELECT id FROM personnel WHERE personnel_code = 'PERS-001'), NULL,
       'deployment',
       (SELECT id FROM stations WHERE code = 'INDIA-HQ'),
       (SELECT id FROM stations WHERE code = 'BHARTI'),
       'arrived', NOW() - INTERVAL '75 days', NOW() - INTERVAL '73 days', NOW() - INTERVAL '73 days',
       'Completed seasonal deployment to Bharti.'
UNION ALL
SELECT (SELECT id FROM personnel WHERE personnel_code = 'PERS-003'), NULL,
       'deployment',
       (SELECT id FROM stations WHERE code = 'INDIA-HQ'),
       (SELECT id FROM stations WHERE code = 'BHARTI'),
       'arrived', NOW() - INTERVAL '42 days', NOW() - INTERVAL '40 days', NOW() - INTERVAL '40 days',
       'Completed logistics deployment to Bharti.'
UNION ALL
SELECT (SELECT id FROM personnel WHERE personnel_code = 'PERS-005'),
       (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-003'),
       'station_transfer',
       (SELECT id FROM stations WHERE code = 'MAITRI'),
       (SELECT id FROM stations WHERE code = 'BHARTI'),
       'arrived', NOW() - INTERVAL '50 days', NOW() - INTERVAL '49 days', NOW() - INTERVAL '49 days',
       'Travelled to Bharti for winter logistics communications support.'
UNION ALL
SELECT (SELECT id FROM personnel WHERE personnel_code = 'PERS-005'),
       (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-003'),
       'station_transfer',
       (SELECT id FROM stations WHERE code = 'BHARTI'),
       (SELECT id FROM stations WHERE code = 'MAITRI'),
       'arrived', NOW() - INTERVAL '37 days', NOW() - INTERVAL '36 days', NOW() - INTERVAL '36 days',
       'Returned to Maitri after winter logistics support.';

-- Devices, heartbeat trails, and incidents make the personnel-detail drawer demo meaningful.
INSERT INTO emergency_devices (
    personnel_id, device_label, status, last_heartbeat_at,
    last_latitude, last_longitude, last_accuracy_m, last_altitude_m,
    last_heading_deg, last_speed_mps, last_battery_percent
)
SELECT (SELECT id FROM personnel WHERE personnel_code = 'PERS-001'), 'SIM-BHARTI-ANANYA', 'active',
       NOW() - INTERVAL '4 minutes', -69.405200, 76.198300, 8, 42, 118, 1.4, 82
UNION ALL
SELECT (SELECT id FROM personnel WHERE personnel_code = 'PERS-002'), 'SIM-TRANSIT-ARJUN', 'active',
       NOW() - INTERVAL '2 minutes', -45.120000, 62.470000, 12, 5, 152, 9.2, 67
UNION ALL
SELECT (SELECT id FROM personnel WHERE personnel_code = 'PERS-004'), 'SIM-MAITRI-ROHAN', 'active',
       NOW() - INTERVAL '8 minutes', -70.766200, 11.734100, 6, 118, 205, 0.2, 91;

INSERT INTO emergency_signals (
    idempotency_key, personnel_id, device_id, signal_type, transmission_channel,
    latitude, longitude, location_accuracy_m, altitude_m, heading_deg,
    speed_mps, battery_percent, occurred_at, sync_status, payload_notes
)
SELECT gen_random_uuid(), p.id, d.id, 'heartbeat', 'mobile_http_simulator',
       point.latitude, point.longitude, point.accuracy, point.altitude, point.heading,
       point.speed, point.battery, NOW() - point.age, 'processed', 'Fresh-demo personnel tracking point'
FROM (VALUES
    ('PERS-001', -69.407500::NUMERIC, 76.191500::NUMERIC, 9::NUMERIC, 41::NUMERIC, 102::NUMERIC, 1.0::NUMERIC, 86::NUMERIC, INTERVAL '34 minutes'),
    ('PERS-001', -69.406700::NUMERIC, 76.194200::NUMERIC, 8::NUMERIC, 42::NUMERIC, 109::NUMERIC, 1.2::NUMERIC, 84::NUMERIC, INTERVAL '18 minutes'),
    ('PERS-001', -69.405200::NUMERIC, 76.198300::NUMERIC, 8::NUMERIC, 42::NUMERIC, 118::NUMERIC, 1.4::NUMERIC, 82::NUMERIC, INTERVAL '4 minutes'),
    ('PERS-002', -39.800000::NUMERIC, 58.120000::NUMERIC, 15::NUMERIC, 5::NUMERIC, 145::NUMERIC, 10.5::NUMERIC, 73::NUMERIC, INTERVAL '5 hours'),
    ('PERS-002', -42.600000::NUMERIC, 60.250000::NUMERIC, 13::NUMERIC, 5::NUMERIC, 149::NUMERIC, 9.8::NUMERIC, 70::NUMERIC, INTERVAL '2 hours'),
    ('PERS-002', -45.120000::NUMERIC, 62.470000::NUMERIC, 12::NUMERIC, 5::NUMERIC, 152::NUMERIC, 9.2::NUMERIC, 67::NUMERIC, INTERVAL '2 minutes'),
    ('PERS-004', -70.767100::NUMERIC, 11.731400::NUMERIC, 7::NUMERIC, 116::NUMERIC, 198::NUMERIC, 0.3::NUMERIC, 94::NUMERIC, INTERVAL '42 minutes'),
    ('PERS-004', -70.766600::NUMERIC, 11.732800::NUMERIC, 6::NUMERIC, 117::NUMERIC, 201::NUMERIC, 0.2::NUMERIC, 92::NUMERIC, INTERVAL '21 minutes'),
    ('PERS-004', -70.766200::NUMERIC, 11.734100::NUMERIC, 6::NUMERIC, 118::NUMERIC, 205::NUMERIC, 0.2::NUMERIC, 91::NUMERIC, INTERVAL '8 minutes')
) AS point(personnel_code, latitude, longitude, accuracy, altitude, heading, speed, battery, age)
JOIN personnel p ON p.personnel_code = point.personnel_code
JOIN emergency_devices d ON d.personnel_id = p.id;

INSERT INTO emergencies (
    emergency_code, emergency_type, station_id, expedition_id, reported_by_personnel_id,
    emergency_device_id, report_channel, severity, status, latitude, longitude,
    location_accuracy_m, summary, reported_at, resolved_at
)
SELECT 'EMG-DEMO-001', 'whiteout', (SELECT id FROM stations WHERE code = 'MAITRI'),
       (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-003'),
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-004'),
       (SELECT id FROM emergency_devices WHERE device_label = 'SIM-MAITRI-ROHAN'),
       'mobile_sos', 'critical', 'resolved', -70.766900, 11.731900, 9,
       'Whiteout delayed return from the winter logistics support route.',
       NOW() - INTERVAL '40 days', NOW() - INTERVAL '40 days' + INTERVAL '3 hours'
UNION ALL
SELECT 'EMG-DEMO-002', 'equipment_fault', (SELECT id FROM stations WHERE code = 'BHARTI'),
       (SELECT id FROM expeditions WHERE expedition_code = 'EXP-ANT-001'),
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-001'),
       (SELECT id FROM emergency_devices WHERE device_label = 'SIM-BHARTI-ANANYA'),
       'mobile_sos', 'moderate', 'active', -69.405200, 76.198300, 8,
       'Field generator fault reported during ice-shelf survey operations.', NOW() - INTERVAL '35 minutes', NULL;

INSERT INTO emergency_personnel (emergency_id, personnel_id, involvement_type, status)
SELECT (SELECT id FROM emergencies WHERE emergency_code = 'EMG-DEMO-001'),
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-004'), 'affected', 'safe'
UNION ALL
SELECT (SELECT id FROM emergencies WHERE emergency_code = 'EMG-DEMO-001'),
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-005'), 'responder', 'completed'
UNION ALL
SELECT (SELECT id FROM emergencies WHERE emergency_code = 'EMG-DEMO-002'),
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-001'), 'affected', 'awaiting_assessment'
UNION ALL
SELECT (SELECT id FROM emergencies WHERE emergency_code = 'EMG-DEMO-002'),
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-003'), 'coordinator', 'responding';

INSERT INTO emergency_timeline_events (emergency_id, event_type, details, occurred_at, recorded_by_personnel_id)
SELECT (SELECT id FROM emergencies WHERE emergency_code = 'EMG-DEMO-001'),
       'sos_confirmed', 'Two-step SOS confirmed from the personnel device.', NOW() - INTERVAL '40 days',
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-004')
UNION ALL
SELECT (SELECT id FROM emergencies WHERE emergency_code = 'EMG-DEMO-001'),
       'resolved', 'Weather cleared and the team returned safely to Maitri.', NOW() - INTERVAL '40 days' + INTERVAL '3 hours',
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-005')
UNION ALL
SELECT (SELECT id FROM emergencies WHERE emergency_code = 'EMG-DEMO-002'),
       'reported', 'Generator fault logged for the active field survey response.', NOW() - INTERVAL '35 minutes',
       (SELECT id FROM personnel WHERE personnel_code = 'PERS-001');

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
