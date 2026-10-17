<!--
SPDX-FileCopyrightText: 2026 Noxfort Systems
SPDX-License-Identifier: Apache-2.0
-->

<div align="center">
  <img src="docs/assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="80" /><br/>
  <h1>EVP Simulator</h1>
  <h3>Emergency Vehicle Preemption & Fleet Simulation Platform for Urban Traffic Networks</h3>
  <i>Noxfort Systems — A State Of Art Company</i><br/><br/>

  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go" alt="Go 1.22+" /></a>
  <a href="https://wails.io/"><img src="https://img.shields.io/badge/Wails-v2-DF0000?style=flat&logo=wails" alt="Wails v2" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/SPDX--License--Identifier-Apache--2.0-blue.svg?style=flat" alt="SPDX-License-Identifier: Apache-2.0" /></a>
  <img src="https://img.shields.io/badge/Status-Active-brightgreen?style=flat" alt="Status Active" />
</div>

---

## 🚦 Overview

**EVP Simulator** is a high-performance microscopic simulation platform built in **Go 1.22** with a native desktop interface powered by **Wails v2** and real-time 2D vector rendering in **HTML5 Canvas**. The system is engineered to model, evaluate, and validate **Emergency Vehicle Preemption (EVP)** strategies across complex metropolitan road networks.

Designed in strict compliance with Brazilian safety guidelines (**CONTRAN Resolution nº 960**, **Brazilian Traffic Code CTB - Art. 29, VII**) and international ITS specifications (**FHWA MUTCD Section 4D.26**), the simulator reproduces realistic vehicle kinematics, calculates prioritized green corridors, and integrates seamlessly with the **CARINA CORE** adaptive traffic network intelligence ecosystem.

---

### 🌐 Select Language / Multi-Language Hubs

| Language | Directory | Status | Description |
| :---: | :---: | :---: | :--- |
| **🇺🇸 English** | [**`docs/en/`**](docs/en/README.md) | **Canonical (Source of Truth)** | Full technical documentation suite in English. |
| **🇧🇷 Português (Brasil)** | [**`docs/pt-br/`**](docs/pt-br/README.md) | **Completo (Referência Primária)** | Documentação técnica completa em Português do Brasil. |
| **🇪🇸 Español** | [**`docs/es/`**](docs/es/README.md) | **Completo (Latinoamérica)** | Documentación técnica completa en Español para operaciones internacionales. |
| **🇫🇷 Français** | [**`docs/fr/`**](docs/fr/README.md) | **Complet (Documentation)** | Suite complète de documentation technique en Français. |
| **🇷🇺 Русский** | [**`docs/ru/`**](docs/ru/README.md) | **Полный комплект** | Полная техническая документация и руководства на русском языке. |
| **🇨🇳 简体中文** | [**`docs/zh/`**](docs/zh/README.md) | **完整技术套件** | 面向智能交通系统 (ITS) 工程师的完整中文技术文档库。 |

---

## 🚀 Quick Start

### Prerequisites
- **Go**: 1.22 or higher
- **Wails CLI**: v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- **Node.js**: 18+ (for frontend assets)

### Running in Development
```bash
wails dev
```

### Compiling Standalone Binary
```bash
wails build
```

---

## 📄 License & SPDX Compliance

This project is licensed under the Apache License, Version 2.0.

- **SPDX Identifier**: `Apache-2.0`
- **SPDX License Expression**: `SPDX-License-Identifier: Apache-2.0`
- **Copyright Holder**: `Copyright (C) 2026 Noxfort Systems`

See the full [LICENSE](LICENSE) file for terms and conditions.

---

<div align="center">
  <img src="docs/assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Intelligent Mobility Engineering • EVP Simulator v1.0.0</i><br/>
  <small>SPDX-License-Identifier: <a href="LICENSE">Apache-2.0</a> • &copy; 2026 Noxfort Systems</small>
</div>
