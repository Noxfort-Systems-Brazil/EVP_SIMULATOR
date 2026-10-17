---
tags: [developer, guides, build, contribution, solid, go122, wails]
aliases: [Guia do Desenvolvedor, Developer Guides, Contribuição]
---

# 🛠️ Guia do Desenvolvedor & Ambiente de Compilação

Este documento orienta engenheiros de software e pesquisadores sobre como configurar o ambiente de desenvolvimento, compilar e estender o **EVP Simulator**.

⬅️ [Central de Documentação](README.md) | 🏛️ [Arquitetura](architecture.md) | 🧪 [Testes](testing_and_validation.md) | 🚀 [Deploy](deployment_and_packaging.md)

---

## 1. Pré-Requisitos de Sistema

- **Go:** Versão `1.22+` com suporte a módulos Go.
- **Node.js:** Versão `18.0+` ou `20.0+` (LTS) e gerenciador `npm`.
- **Wails CLI:** Versão `v2.8.0+`:
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- **Dependências de Sistema (Linux/Debian/Ubuntu):**
  ```bash
  sudo apt-get update && sudo apt-get install -y \
      libgtk-3-dev \
      libwebkit2gtk-4.0-dev \
      build-essential \
      pkg-config
  ```

---

## 2. Executando em Modo de Desenvolvimento com Live Reload

O Wails permite compilar o backend Go e o frontend JS em modo de desenvolvimento integrado com *hot reload*:

```bash
cd /home/gabriel-moraes/Documentos/EVP_SIMULATOR
wails dev
```

Alterações em arquivos Go recompilam o binário automaticamente, e alterações no HTML/CSS/JS atualizam a janela desktop em tempo real.

---

## 3. Diretrizes de Arquitetura e Extensão

Ao implementar novos recursos, respeite as fronteiras de arquitetura:
1. **Regras de Negócio no Domínio:** Regras de tráfego, cálculos de aceleração e avaliação semafórica devem residir estritamente em `internal/domain/`.
2. **Sem Dependências Externas no Domínio:** Não importe bibliotecas de terceiros ou pacotes do Wails dentro de `domain`. Toda interação externa deve ser intermediada por interfaces em `port/`.
3. **Novos Adaptadores de Infraestrutura:** Ao integrar novos protocolos semafóricos (ex: SCATS, SCOOT, OASIS), implemente a interface `port.ControllerServerManager` em `internal/infrastructure/`.

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Engenharia de Mobilidade Inteligente • EVP Simulator v1.0.0</i>
  <small>SPDX-License-Identifier: <a href="../../LICENSE">Apache-2.0</a> • &copy; 2026 Noxfort Systems</small>
</div>
