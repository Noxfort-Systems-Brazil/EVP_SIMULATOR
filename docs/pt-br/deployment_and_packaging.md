---
tags: [deployment, packaging, wails-build, cross-compile, release, linux, windows]
aliases: [Deploy e Empacotamento, Compilação de Produção, Wails Build]
---

# 🚀 Compilação, Empacotamento e Distribuição

Este documento aborda o processo de geração de binários de produção standalone otimizados para Linux e Windows.

⬅️ [Central de Documentação](README.md) | 🏛️ [Arquitetura](architecture.md) | 🛠️ [Guia do Desenvolvedor](developer_guides.md)

---

## 1. Compilação de Binários Standalone Nativos

O Wails empacota a interface HTML/CSS/JS diretamente no executável binário usando a diretiva `//go:embed` da biblioteca padrão de Go, eliminando dependências de servidores Node.js externos em produção:

```bash
# Compilação otimizada para Linux (binário em build/bin/evp_simulator)
wails build -clean -platform linux/amd64

# Compilação otimizada para Windows (.exe standalone)
wails build -clean -platform windows/amd64
```

### 1.1 Flags de Otimização de Tamanho e Performance
O compilador de Go remove símbolos de debug para reduzir o binário para ~12 MB:
```bash
wails build -ldflags "-s -w"
```

---

## 2. Requisitos de Execução em Produção

| Sistema Operacional | Requisitos de Runtime | Arquitetura |
| :--- | :--- | :---: |
| **Linux (Ubuntu / Debian / Fedora)** | WebKitGTK 4.0+ (`libwebkit2gtk-4.0-37`) | `x86_64` |
| **Windows 10 / 11** | Microsoft Edge WebView2 Runtime (Nativo) | `x86_64` |

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Engenharia de Mobilidade Inteligente • EVP Simulator v1.0.0</i>
</div>
