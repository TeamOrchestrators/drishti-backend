# DRISHTI Backend

**DRISHTI — Data Reporting & Intelligent Surveillance for High-impact Tracking Interface** is a prototype backend for coordinated polar-expedition operations. It is being developed for Smart India Hackathon (SIH) and an internal hackathon demonstration.

The service brings expedition planning, personnel movement, station inventory, QR-oriented cargo logistics, and simulated emergency reporting into a single PostgreSQL-backed API.

> This prototype does not claim to provide satellite connectivity or vessel/aircraft tracking. The emergency device flow is an HTTP mobile-web simulation used to demonstrate SOS reporting and GPS heartbeat handling.

## Core capabilities

| Area | What DRISHTI supports |
| --- | --- |
| Stations | Research-station catalogue and station-aware operations |
| Expeditions | Create, update, list, and prepare expeditions with valid station and personnel options |
| Personnel | Availability, movement history, active movement state, and expedition assignment validation |
| Inventory | Per-station inventory, immutable stock transactions, available quantity, and configurable stock thresholds |
| Inventory analytics | Optional integration with an inventory analytics microservice for alerts, stock days, depletion, anomalies, and consumption trends |
| Logistics | Cargo, manifests, QR-ready cargo UUIDs, logistics batches, and expedition-linked dispatch flows |
| Emergencies | Mobile-device registration, GPS heartbeats, two-step SOS confirmation, live emergency status, and resolution |

## Architecture

```text
Frontend / mobile simulator
            |
            v
      Go net/http API
            |
   Handlers -> Services -> sqlc queries
            |                 |
            |                 v
            |             PostgreSQL
            |
            +--> Optional inventory analytics microservice
```

The backend uses Go's standard `net/http` router, `pgx` for PostgreSQL connectivity, Goose for migrations, and SQLC for type-safe query generation.

## Project structure

```text
.
├── db/
│   ├── sqlc.yaml                 # SQLC configuration
│   └── db/generated/             # Generated SQLC code
├── internal/
│   ├── station/
│   ├── personnel/
│   ├── expedition/
│   ├── inventory/
│   ├── logistics/
│   └── emergency/
├── sql/
│   ├── migrations/               # Ordered Goose migrations
│   ├── queries/                  # SQLC query definitions
│   ├── schemas/                  # Canonical database schema by module
│   └── seeds/                    # Optional demo datasets
├── main.go
└── go.mod
```

Each domain module follows a consistent structure:

```text
handler.go  -> HTTP request/response handling
service.go  -> validation and business logic
request.go  -> API request contracts
response.go -> API response contracts
routes.go   -> route registration
```

## Prerequisites

- Go 1.27+
- PostgreSQL 14+ with `pgcrypto` enabled
- [Goose](https://github.com/pressly/goose) CLI
- [SQLC](https://docs.sqlc.dev/en/latest/overview/install.html) CLI

Enable UUID generation once in the target database:

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```

## Quick start

1. Create a `.env` file in the project root:

   ```env
   DB_URL=postgres://postgres:password@localhost:5432/drishti?sslmode=disable

   # Optional: enables inventory analytics enrichment.
   INVENTORY_MICROSERVICE_URL=https://your-inventory-service.example
   ```

2. Apply migrations:

   ```powershell
   goose -dir sql/migrations up
   ```

3. Generate SQLC code after changing a schema or query:

   ```powershell
   sqlc generate -f db/sqlc.yaml
   ```

4. Run the API:

   ```powershell
   go run .
   ```

The service starts on `http://localhost:8080`.

5. Verify the build:

   ```powershell
   go test ./...
   ```

## API overview

### Stations

| Method | Route | Description |
| --- | --- | --- |
| `GET` | `/api/stations` | List stations |

### Expeditions

| Method | Route | Description |
| --- | --- | --- |
| `GET` | `/api/expeditions` | List expeditions |
| `GET` | `/api/expeditions/form-options` | List valid station and personnel options |
| `POST` | `/api/expeditions` | Create an expedition |
| `PUT` | `/api/expeditions/{id}` | Update an expedition |

### Personnel

| Method | Route | Description |
| --- | --- | --- |
| `GET` | `/api/personnel` | Return available, moving, and historical personnel views |
| `GET` | `/api/personnel/assignment-form-options` | List eligible personnel and expeditions |
| `POST` | `/api/personnel/expedition-assignment` | Assign personnel and create their movement record |

Personnel assignment validates medical clearance, availability, origin station, existing active movement, and the selected expedition's planned dates.

### Inventory

| Method | Route | Description |
| --- | --- | --- |
| `GET` | `/api/inventory?station_id={uuid}` | Return station inventory, recent history, and optional analytics |
| `POST` | `/api/inventory/items?station_id={uuid}` | Create an item and opening stock record |
| `GET` | `/api/stations/{stationID}/inventory` | Return station inventory |
| `POST` | `/api/stations/{stationID}/inventory/items` | Create a station item |
| `POST` | `/api/stations/{stationID}/inventory/{inventoryID}/stock` | Add or remove stock |

Inventory uses three quantities:

```text
on_hand_quantity - reserved_quantity = available_quantity
```

Every adjustment creates an immutable `inventory_transactions` record.

### Cargo and logistics

| Method | Route | Description |
| --- | --- | --- |
| `GET` | `/api/cargo` | List cargo and logistics batches |
| `POST` | `/api/cargo` | Create cargo |
| `PUT` | `/api/cargo/{id}/logistics-batch` | Assign cargo to a logistics batch |
| `POST` | `/api/logistics-batches` | Create an expedition-linked batch |

Cargo UUIDs act as QR tokens. A batch is assigned to an expedition, and cargo is assigned to a batch.

```text
Expedition -> Logistics batch -> Cargo -> Cargo items / QR scans
```

### Emergencies and mobile device simulation

| Method | Route | Description |
| --- | --- | --- |
| `GET` | `/api/emergencies` | List all emergencies |
| `GET` | `/api/emergencies/active` | List active, acknowledged, and responding emergencies |
| `PATCH` | `/api/emergencies/{emergencyID}/status` | Change an emergency status |
| `POST` | `/api/emergencies/{emergencyID}/resolve` | Resolve an emergency |
| `GET` | `/api/emergency/device/simulate/form-options` | List people and registered simulated devices |
| `POST` | `/api/emergency/device/simulate/register` | Register or reactivate a simulated device |
| `POST` | `/api/emergency/device/simulate/{deviceID}/heartbeat` | Record a GPS heartbeat |
| `POST` | `/api/emergency/device/simulate/{deviceID}/sos/initiate` | Begin an SOS confirmation window |
| `POST` | `/api/emergency/device/simulate/{deviceID}/sos/{confirmationID}/confirm` | Confirm SOS and create the emergency |

The SOS flow is deliberately two-step:

```text
Register device -> GPS heartbeat -> Initiate SOS -> Confirm within 30 seconds -> Emergency active
```

The confirmation must be explicitly completed by the client. A single accidental tap cannot create an emergency.

## Inventory analytics integration

When `INVENTORY_MICROSERVICE_URL` is configured, the backend calls the analytics service server-to-server. The stock-days request includes live Drishti inventory and consumption history:

```json
{
  "inventory_id": "uuid",
  "current_quantity": 142,
  "inventory_transactions": [
    {
      "transaction_type": "consumption",
      "quantity_delta": -13,
      "occurred_at": "2026-09-11T00:00:00Z"
    }
  ]
}
```

Analytics are optional: inventory data remains available if the external service is unavailable. For per-item anomaly, depletion, or average-consumption results, the analytics service should return the corresponding `inventory_id`.

## Demo data

Demo seeds are intended for local or hackathon environments only.

| File | Purpose |
| --- | --- |
| `sql/seeds/001_inventory_ai_demo.sql` | Adds idempotent Bharti inventory and consumption examples |
| `sql/seeds/002_full_demo_reset.sql` | Clears all operational data except stations/personnel and loads a complete demo scenario |

Run the full reset seed only when you intend to remove operational data:

```powershell
psql "<your DB_URL value>" -v ON_ERROR_STOP=1 -f sql/seeds/002_full_demo_reset.sql
```

It preserves `stations` and `personnel`, while replacing expeditions, movements, inventory, logistics, cargo, QR scans, and emergency records.

## Development notes

- All primary identifiers are PostgreSQL-generated UUIDs.
- No authentication layer is included in this prototype; government identity-provider requirements are outside the current hackathon scope.
- Device heartbeats are HTTP simulation data, not satellite uplinks.
- No vessel, aircraft, or other vehicle tracking is implemented.
- Do not commit `.env` files or production credentials.

## License

Copyright 2026 Team Orchestrators.

This project is licensed under the Apache License, Version 2.0. You may use, modify, and distribute it in accordance with the terms of the license. See [LICENSE](LICENSE) for the full text.

SPDX identifier: `Apache-2.0`
