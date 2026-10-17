---
tags: [api, reference, protocols, snmp, http, wails, dto, specifications]
aliases: [Referência de APIs, Protocolos de Comunicação, API Reference]
---

# 📖 Referência de APIs, Protocolos e Especificação de Interfaces

Este documento é a especificação formal e autoritativa de todos os contratos de dados e protocolos do **EVP Simulator**.

⬅️ [Central de Documentação](README.md) | 🏛️ [Arquitetura](architecture.md) | ⚡ [Ponte Wails](wails_ipc_and_bridges.md) | 📡 [Integração CARINA](carina_integration_and_telemetry.md)

---

## 1. Contrato da API Wails Bridge (JavaScript ↔ Go)

### 1.1 `App.StartSimulation()`
- **Assinatura:** `StartSimulation() => Promise<void>`
- **Efeito:** Inicia o ticker em 100 ms e a emissão de eventos `"telemetry_tick"`.

### 1.2 `App.SetVehicleMode(vehicleID, mode)`
- **Parâmetros:**
  - `vehicleID` (`string`): Identificador do veículo (ex: `"AMB_01"`).
  - `mode` (`string`): `"Patrol"`, `"Emergency"`, `"Inactive"`.
- **Efeito:** Atualiza imediatamente as metas de aceleração e sinalização de sirene.

---

## 2. Contrato HTTP de Telemetria com a CARINA CORE

- **Método & Rota:** `POST /api/v1/evp/telemetry`
- **Headers:** `Content-Type: application/json`
- **Corpo da Mensagem:**
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
- **Resposta Sucesso:** `HTTP 204 No Content`

---

## 3. Catálogo de OIDs SNMP (Emulação UTMC2 / NTCIP 1202)

| OID Numérico | Nome Simbólico | Tipo |
| :--- | :--- | :--- |
| `.1.3.6.1.4.1.1202.1.1.0` | `tlJunctionID` | String |
| `.1.3.6.1.4.1.1202.1.2.0` | `tlActivePhase` | Integer (0-based) |
| `.1.3.6.1.4.1.1202.2.2.0` | `tlPreemptionHold` | Integer (0=Livre, 1=Retido) |

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Engenharia de Mobilidade Inteligente • EVP Simulator v1.0.0</i>
</div>
