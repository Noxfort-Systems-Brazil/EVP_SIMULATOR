---
tags: [api, reference, protocols, snmp, http, wails, dto, specifications]
aliases: [API Reference, Communication Protocols, Specifications]
---

# 📖 API Reference, Communication Protocols & Interface Contracts

This document serves as the authoritative interface specification for the **EVP Simulator**.

⬅️ [Documentation Hub](README.md) | 🏛️ [Architecture](architecture.md) | ⚡ [Wails Bridge](wails_ipc_and_bridges.md) | 📡 [CARINA Integration](carina_integration_and_telemetry.md)

---

## 1. Wails Bridge API Contract (JavaScript ↔ Go)

### 1.1 `App.StartSimulation()`
- **Signature:** `StartSimulation() => Promise<void>`
- **Effect:** Initiates 100 ms tick loop and begins emitting `"telemetry_tick"` events.

### 1.2 `App.SetVehicleMode(vehicleID, mode)`
- **Parameters:**
  - `vehicleID` (`string`): Vehicle identifier (e.g., `"AMB_01"`).
  - `mode` (`string`): `"Patrol"`, `"Emergency"`, `"Inactive"`.
- **Effect:** Instantly updates target speed, acceleration envelope, and siren state.

---

## 2. CARINA CORE HTTP Telemetry Contract

- **Method & Route:** `POST /api/v1/evp/telemetry`
- **Headers:** `Content-Type: application/json`
- **Body Schema:**
  ```json
  {
    "vehicle_id": "string",
    "corporation": "string",
    "latitude": -23.31028,
    "longitude": -51.16278,
    "speed_kmh": 65.0,
    "bearing": 180.0,
    "siren_active": true,
    "current_edge": "string",
    "target_intersection": "string",
    "preemption_active": true,
    "timestamp": 1729188000000
  }
  ```
- **Success Response:** `HTTP 204 No Content`

---

## 3. SNMP OID Catalog (UTMC2 / NTCIP 1202 Emulation)

| Numeric OID | Symbol Name | Data Type |
| :--- | :--- | :--- |
| `.1.3.6.1.4.1.1202.1.1.0` | `tlJunctionID` | String |
| `.1.3.6.1.4.1.1202.1.2.0` | `tlActivePhase` | Integer (0-based) |
| `.1.3.6.1.4.1.1202.2.2.0` | `tlPreemptionHold` | Integer (0=Released, 1=Held) |

---

<div align="center">
  <img src="assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Intelligent Mobility Engineering • EVP Simulator v1.0.0</i>
</div>
