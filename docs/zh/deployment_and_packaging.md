---
tags: [deployment, packaging, wails-build, cross-compile, release, linux, windows]
aliases: [Deployment, Production Packaging, Wails Build]
---

# 🚀 原生编译、打包与单文件桌面部署

This document covers the production build process for standalone Linux and Windows desktop binaries.

⬅️ [技术文档中心](README.md) | 🏛️ [Architecture](architecture.md) | 🛠️ [Developer Guides](developer_guides.md)

---

## 1. Standalone Native Production Build

Wails embeds the frontend assets directly into the Go executable via `//go:embed`:

```bash
# Build optimized binary for Linux
wails build -clean -platform linux/amd64

# Build standalone executable for Windows
wails build -clean -platform windows/amd64
```

---

## 2. Runtime Requirements

| Operating System | Runtime Requirements | Architecture |
| :--- | :--- | :---: |
| **Linux (Ubuntu / Debian / Fedora)** | WebKitGTK 4.0+ (`libwebkit2gtk-4.0-37`) | `x86_64` |
| **Windows 10 / 11** | Microsoft Edge WebView2 Runtime (Built-in) | `x86_64` |

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>智能交通工程 • EVP Simulator v1.0.0</i>
  <small>SPDX-License-Identifier: <a href="../../LICENSE">Apache-2.0</a> • &copy; 2026 Noxfort Systems</small>
</div>
