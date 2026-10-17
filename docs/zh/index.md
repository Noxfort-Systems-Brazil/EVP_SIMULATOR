---
tags: [home, index, evp, simulator, hub]
aliases: [Documentation Portal, Overview, EVP Simulator Hub]
---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="70" /><br/>
  <h1>EVP Simulator — 官方技术门户与仿真中心</h1>
  <h3>开源城市路网紧急车辆优先通行与车队仿真平台</h3>
  <i>Noxfort Systems — A State Of Art Company</i>
</div>

---

## 🚦 系统核心概述

**EVP Simulator** is a high-performance microscopic simulation platform built in **Go 1.22** with a native desktop interface powered by **Wails v2** and real-time 2D vector rendering in **HTML5 Canvas**. The system is engineered to model, evaluate, and validate **Emergency Vehicle Preemption (EVP)** strategies across complex metropolitan road networks.

Designed in strict compliance with Brazilian safety guidelines (**CONTRAN Resolution nº 960**, **Brazilian Traffic Code CTB - Art. 29, VII**) and international ITS specifications (**FHWA MUTCD Section 4D.26**), the simulator reproduces realistic vehicle kinematics, calculates prioritized green corridors, and integrates seamlessly with the **CARINA CORE** adaptive traffic network intelligence ecosystem.

---

## 🗺️ 主导航地图与专业子系统

| 子系统 / 维度 | 技术范围 | 文档链接 |
| :--- | :--- | :---: |
| 📚 **Master Map of Content** | Obsidian Knowledge Graph & Go Codebase Package Hierarchy | [Explore MOC](evp_moc.md) |
| 🏛️ **System Architecture** | Clean Architecture, Hexagonal / Ports & Adapters, and Goroutines | [View Architecture](architecture.md) |
| 🚨 **Preemption Walkthrough** | Approach Detection ($60\text{--}160\text{ m}$), Checkout ($15\text{ m}$), and Watchdog | [View Walkthrough](evp_walkthrough.md) |
| 🚗 **Kinematics & A\* Routing** | Euler Kinematic Engine, Vehicle Modes, and Graph A\* Pathfinding | [View Kinematics](kinematics_and_routing.md) |
| 🗺️ **SUMO Network Parser** | `.net.xml` and `.net.xml.gz` Ingestion, Cartesian/WGS84 Projection | [View SUMO Parser](sumo_network_parser.md) |
| 🔌 **SNMP / UTMC Emulation** | UDP SNMP Servers in Go, UTMC2, and NTCIP 1202 Protocol Agents | [View SNMP Emulation](utmc_and_snmp_emulation.md) |
| 📡 **CARINA CORE Integration** | Fire-and-Forget HTTP Telemetry Dispatch `POST /api/v1/evp/telemetry` | [View Integration](carina_integration_and_telemetry.md) |
| 🚑 **Fleet Management** | Procedural Emergency Fleet Generator (SAMU, Fire, Police) | [View Fleet Spawner](fleet_management_and_spawner.md) |
| 🛡️ **Safety & Signal Evaluation** | Phase Evaluation, Clearance Intervals (Yellow/All-Red), and Vetoes | [View Safety](signal_evaluation_and_safety.md) |
| 📊 **CSV Data Exchange** | Physical Controller Inventory and Trajectory Logging | [View CSV Exchanger](csv_data_exchanger.md) |
| 🖥️ **Desktop UI & 2D Canvas** | Wails v2 Desktop GUI, 60 FPS HTML5 Canvas Engine, Pan & Zoom | [View UI & Rendering](ui_and_canvas_rendering.md) |
| ⚡ **Wails IPC Bridge (Go ↔ JS)** | RPC Methods, Typed DTOs, and Asynchronous Event Bus | [View Wails IPC](wails_ipc_and_bridges.md) |
| 🛠️ **Developer Guides** | Environment Setup, Go 1.22+, Wails CLI, and Architecture Rules | [View Developer Guides](developer_guides.md) |
| 🧪 **Testing & Validation** | Unit Testing, Kinematics Benchmarks, and SNMP Mock Integration | [View Testing](testing_and_validation.md) |
| 🚀 **Deployment & Packaging** | Native Standalone Build (`wails build`), Linux and Windows Bundles | [View Deployment](deployment_and_packaging.md) |
| 📖 **API & Protocol Reference** | Wails RPC Signatures, Telemetry JSON Schemas, and SNMP OIDs | [View API Reference](api_and_protocols_reference.md) |

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>智能交通工程 • EVP Simulator v1.0.0</i>
</div>
