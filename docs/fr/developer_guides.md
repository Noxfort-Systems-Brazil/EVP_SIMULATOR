---
tags: [developer, guides, build, contribution, solid, go122, wails]
aliases: [Developer Guides, Build Environment, Contribution]
---

# 🛠️ Guides du Développeur & Environnement de Compilation

This guide provides setup instructions and engineering standards for building and extending the **EVP Simulator**.

⬅️ [Hub de Documentation](README.md) | 🏛️ [Architecture](architecture.md) | 🧪 [Testing](testing_and_validation.md) | 🚀 [Deployment](deployment_and_packaging.md)

---

## 1. Prerequisites

- **Go:** `1.22+` with Go modules enabled.
- **Node.js:** `18.0+` or `20.0+` (LTS) with `npm`.
- **Wails CLI:**
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- **Linux Packages:** `libgtk-3-dev`, `libwebkit2gtk-4.0-dev`, `build-essential`.

---

## 2. Running in Development Mode

```bash
cd /home/gabriel-moraes/Documentos/EVP_SIMULATOR
wails dev
```

Hot reload monitors both Go files (auto-recompiling) and frontend assets (auto-refreshing desktop view).

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Ingénierie de Mobilité Intelligente • EVP Simulator v1.0.0</i>
  <small>SPDX-License-Identifier: <a href="../../LICENSE">Apache-2.0</a> • &copy; 2026 Noxfort Systems</small>
</div>
