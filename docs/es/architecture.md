---
tags: [architecture, go, clean-architecture, ddd, ports-and-adapters, concurrency]
aliases: [System Architecture, Go Architecture, Hexagonal Architecture]
---

# 🏛️ Arquitectura del Sistema Go y Modelo de Concurrencia

This document details the engineering principles of the **EVP Simulator**. Built in **Go 1.22**, the application strictly adheres to **Clean Architecture**, **Dominio-Driven Design (DDD)**, and **Hexagonal Architecture (Ports & Adapters)**, ensuring complete decoupling between traffic kinematic logic, UI rendering, and external communication protocols.

⬅️ [Centro de Documentación](README.md) | 🗺️ [Master MOC](evp_moc.md) | 🚨 [EVP Walkthrough](evp_walkthrough.md) | ⚡ [Wails Bridge](wails_ipc_and_bridges.md)

---

## 1. Hexagonal Architecture Principles

The core simulation engine has zero external dependencies on graphical frameworks, databases, or networking libraries:

```mermaid
graph TD
    subgraph DrivingAdapters ["🔌 Driving Adapters (Primary)"]
        WailsUI["Wails v2 Desktop GUI (HTML5/JS)"]
        CmdLine["Automated Tests / Headless CLI"]
    end

    subgraph AppCore ["⚙️ Application Core"]
        AppBridge["internal/infrastructure/wails/App"]
        SimUseCase["usecase.SimulationUseCase"]
        TelDisp["usecase.TelemetryDispatcher"]
    end

    subgraph DominioCore ["🧠 Pure Dominio (Zero External Dependencies)"]
        Entities["Entities: RoadNetwork, Vehicle, TrafficLight"]
        Services["Services: KinematicsEngine, SignalEvaluator, AStarRouter"]
        Ports["Ports (Interfaces): NetworkRepo, ControllerRepo, TelemetryExporter"]
    end

    subgraph DrivenAdapters ["🔌 Driven Adapters (Secondary)"]
        SumoParser["SUMO Parser (sumo.Parser)"]
        SNMPServers["UTMC2 SNMP Servers (utmc.ServerManager)"]
        CarinaClient["CARINA HTTP Client (httpclient.CarinaExporter)"]
        CSVFiles["CSV Importer/Exporter (csv.ControllerImporter)"]
    end

    WailsUI --> AppBridge
    CmdLine --> SimUseCase
    AppBridge --> SimUseCase
    SimUseCase --> Entities
    SimUseCase --> Services
    SimUseCase --> Ports
    SumoParser -.->|Implements| Ports
    SNMPServers -.->|Implements| Ports
    CarinaClient -.->|Implements| Ports
    CSVFiles -.->|Implements| Ports
```

---

## 2. Concurrency Model & Simulation Lifecycle

Simulation execution progresses in discrete time steps (*ticks*) inside a dedicated goroutine decoupled from the GUI rendering thread:

```mermaid
sequenceDiagram
    autonumber
    actor Operator as 👤 Desktop Operator
    participant UI as 🖥️ Wails Frontend (app.js)
    participant Bridge as 🌉 AppBridge (Go)
    participant UC as ⚙️ SimulationUseCase
    participant Kin as 🚗 KinematicsEngine
    participant Eval as 🚦 SignalEvaluator
    participant CARINA as 📡 CARINA CORE (HTTP)

    Operator->>UI: Clicks "Start Simulation"
    UI->>Bridge: App.StartSimulation()
    Bridge->>UC: uc.Start()
    Note over UC: Spawns background tick loop (100 ms / 10 Hz)
    loop Every Tick (dt = 0.1s)
        UC->>Kin: UpdateVehicle(dt)
        Kin-->>UC: Updated positions and speeds
        UC->>Eval: EvaluateApproach(vehicle, network)
        Eval-->>UC: Preemption commands (Hold/Advance)
        UC->>UI: runtime.EventsEmit("telemetry_tick", DTO)
        opt Vehicle in Active Emergency
            UC->>CARINA: POST /api/v1/evp/telemetry (Async)
        end
    end
```

### 2.1 State Protection & Thread Safety
The shared simulation state (`SimulationState`) is protected by a read/write mutex (`sync.RWMutex`):
- Projections for UI rendering and CSV logging acquire `RLock()`, enabling concurrent access without UI latency spikes.
- Tick progression, fleet spawning, and scenario reset acquire `Lock()`, ensuring atomic state mutation and eliminating data races.

---

## 3. Non-Blocking Network I/O Decoupling

To prevent external network latency from stalling the 60 FPS UI rendering or the 10 Hz physics step, `TelemetryDispatcher` uses a bounded asynchronous queue:
- Telemetry events are enqueued into a buffered channel of 1,000 items.
- A dedicated background worker routine consumes frames and executes HTTP calls.
- If network congestion occurs, non-critical frames are discarded with rate-limited warning logs (*load shedding*), preserving simulation determinism.

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Ingeniería de Movilidad Inteligente • EVP Simulator v1.0.0</i>
  <small>SPDX-License-Identifier: <a href="../../LICENSE">Apache-2.0</a> • &copy; 2026 Noxfort Systems</small>
</div>
