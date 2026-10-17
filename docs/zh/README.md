<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="70" /><br/>
  <h1>EVP Simulator — 技术文档中心</h1>
  <h3>Multi-语言 Technical Engineering Suite</h3>
  <i>Noxfort Systems — A State Of Art Company</i><br/><br/>

  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go" alt="Go 1.22+" /></a>
  <a href="https://wails.io/"><img src="https://img.shields.io/badge/Wails-v2-DF0000?style=flat&logo=wails" alt="Wails v2" /></a>
  <a href="../LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=flat" alt="License Apache 2.0" /></a>
  <img src="https://img.shields.io/badge/状态-Active-brightgreen?style=flat" alt="状态 Active" />
</div>

---

### 🌐 选择语言 / 多语言文档库

| 语言 | 目录 | 状态 | 说明 |
| :---: | :---: | :---: | :--- |
| **🇺🇸 English** | [**`docs/en/`**](README.md) | **Canonical (Source of Truth)** | Full technical documentation suite in English. |
| **🇧🇷 Português (Brasil)** | [**`docs/pt-br/`**](../pt-br/README.md) | **Completo (Referência Primária)** | Documentação técnica completa em Português do Brasil. |
| **🇪🇸 Español** | [**`docs/es/`**](../es/README.md) | **Completo (Latinoamérica)** | Documentación técnica completa en Español para operaciones internacionales. |
| **🇫🇷 Français** | [**`docs/fr/`**](../fr/README.md) | **Complet (Documentation)** | Suite complète de documentation technique en Français. |
| **🇷🇺 Русский** | [**`docs/ru/`**](../ru/README.md) | **Полный комплект** | Полная техническая документация и руководства на русском языке. |
| **🇨🇳 简体中文** | [**`docs/zh/`**](../zh/README.md) | **完整技术套件** | 面向智能交通系统 (ITS) 工程师的完整中文技术文档库。 |

---

## 📚 Technical Guide 目录

| Document | 领域 | 核心技术摘要 |
| :--- | :--- | :--- |
| 🗺️ **[Master MOC](evp_moc.md)** | General Index | Obsidian knowledge graph and Go package taxonomy. |
| 🏛️ **[System Architecture](architecture.md)** | Core Architecture | Hexagonal architecture, Ports & Adapters, and goroutine concurrency. |
| 🚨 **[Preemption Walkthrough](evp_walkthrough.md)** | EVP Priority | Traffic signal preemption algorithms, spatial thresholds, and checkout. |
| 🚗 **[Kinematics & A\* Routing](kinematics_and_routing.md)** | Physics & Graphs | Euler kinematics, acceleration curves, and A\* road graph routing. |
| 🗺️ **[SUMO Network Parser](sumo_network_parser.md)** | Network Ingestion | Native streaming reader for `.net.xml` and `.net.xml.gz` road topologies. |
| 🔌 **[SNMP / UTMC Emulation](utmc_and_snmp_emulation.md)** | Virtual Hardware | In-memory UDP SNMP agent servers in Go for NTCIP 1202 controller emulation. |
| 📡 **[CARINA CORE Integration](carina_integration_and_telemetry.md)** | External Telemetry | Asynchronous fire-and-forget client for `POST /api/v1/evp/telemetry`. |
| 🚑 **[Fleet Management](fleet_management_and_spawner.md)** | Fleet Modeling | Procedural spawner for SAMU ambulances, fire engines, and police cruisers. |
| 🛡️ **[Safety & Signal Evaluation](signal_evaluation_and_safety.md)** | Safety Auditing | Signal phase clearance, yellow/all-red intervals, and conflict vetoes. |
| 📊 **[CSV Data Exchange](csv_data_exchanger.md)** | Relational Data | Schemas for controller inventory CSV and fleet mission logging. |
| 🖥️ **[Desktop UI & 2D Canvas](ui_and_canvas_rendering.md)** | Frontend Desktop | Wails v2, 60 FPS HTML5 Canvas rendering, and interactive pan/zoom camera. |
| ⚡ **[Wails IPC Bridge](wails_ipc_and_bridges.md)** | Go/JS Communication | RPC methods, typed DTOs, and asynchronous event streaming. |
| 🛠️ **[Developer Guides](developer_guides.md)** | Engineering | Build setup, architecture rules, and extending simulation services. |
| 🧪 **[Testing & Validation](testing_and_validation.md)** | Software QA | Unit test suites, kinematics benchmarks, and network integration tests. |
| 🚀 **[Deployment & Packaging](deployment_and_packaging.md)** | Operations | Native binary builds for Linux and Windows with bundled assets. |
| 📖 **[API & Protocol Reference](api_and_protocols_reference.md)** | Specifications | Complete Wails contracts, telemetry JSON schemas, and SNMP OIDs. |

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>智能交通工程 • EVP Simulator v1.0.0</i>
</div>
