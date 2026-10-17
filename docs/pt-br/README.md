<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="70" /><br/>
  <h1>EVP Simulator — Central de Documentação</h1>
  <h3>Suíte Técnica Multilíngue de Guias de Engenharia</h3>
  <i>Noxfort Systems — A State Of Art Company</i><br/><br/>

  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go" alt="Go 1.22+" /></a>
  <a href="https://wails.io/"><img src="https://img.shields.io/badge/Wails-v2-DF0000?style=flat&logo=wails" alt="Wails v2" /></a>
  <a href="../LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=flat" alt="License Apache 2.0" /></a>
  <img src="https://img.shields.io/badge/Status-Active-brightgreen?style=flat" alt="Status Active" />
</div>

---

### 🌐 Selecione o Idioma / Multi-Language Hubs

| Idioma | Diretório | Status | Descrição |
| :---: | :---: | :---: | :--- |
| **🇧🇷 Português (Brasil)** | [**`docs/pt-br/`**](README.md) | **Canônico (Referência Primária)** | Suíte completa em Português do Brasil para engenheiros e operadores. |
| **🇺🇸 English** | [**`docs/en/`**](../en/README.md) | **Complete (Canonical Suite)** | Full technical documentation suite in English. |
| **🇪🇸 Español** | [**`docs/es/`**](../es/README.md) | **Completo (Latinoamérica)** | Documentación técnica completa en Español para operaciones internacionales. |
| **🇫🇷 Français** | [**`docs/fr/`**](../fr/README.md) | **Complet (Documentation)** | Suite complète de documentation technique en Français. |
| **🇷🇺 Русский** | [**`docs/ru/`**](../ru/README.md) | **Полный комплект** | Полная техническая документация и руководства на русском языке. |
| **🇨🇳 简体中文** | [**`docs/zh/`**](../zh/README.md) | **完整技术套件** | 面向智能交通系统 (ITS) 工程师的完整中文技术文档库。 |

---

## 📚 Índice de Guias Especializados

| Documento | Domínio | Descrição Técnica |
| :--- | :--- | :--- |
| 🗺️ **[Master MOC](evp_moc.md)** | Índice Geral | Grafo Obsidian e taxonomia completa dos pacotes Go. |
| 🏛️ **[Arquitetura do Sistema](architecture.md)** | Core do Sistema | Arquitetura Hexagonal, Ports & Adapters e concorrência com goroutines. |
| 🚨 **[Walkthrough de Preempção](evp_walkthrough.md)** | Preempção EVP | Algoritmo de prioridade semafórica, gatilhos espaciais e checkout. |
| 🚗 **[Cinemática & Roteamento A\*](kinematics_and_routing.md)** | Física e Grafos | Integração de Euler, aceleração/frenagem e busca em grafo viário. |
| 🗺️ **[Parser de Redes SUMO](sumo_network_parser.md)** | Ingestão Viária | Leitor nativo de arquivos `.net.xml` e `.net.xml.gz` com projeção espacial. |
| 🔌 **[Emulação SNMP / UTMC](utmc_and_snmp_emulation.md)** | Hardware Virtual | Servidores UDP SNMP em Go para controle semafórico NTCIP 1202. |
| 📡 **[Integração CARINA CORE](carina_integration_and_telemetry.md)** | Telemetria Externa | Cliente HTTP Fire-and-Forget `POST /api/v1/evp/telemetry`. |
| 🚑 **[Gerenciamento de Frotas](fleet_management_and_spawner.md)** | Modelagem de Frota | Spawn de ambulâncias (SAMU), bombeiros e patrulhas policiais. |
| 🛡️ **[Segurança & Avaliação](signal_evaluation_and_safety.md)** | Trânsito & Segurança | Auditoria de segurança, tempos de amarelo/vermelho total e vetos. |
| 📊 **[Importação / Exportação CSV](csv_data_exchanger.md)** | Dados Relacionais | Schemas de inventário de controladores e logs de frotas. |
| 🖥️ **[Interface & Canvas 2D](ui_and_canvas_rendering.md)** | Frontend Desktop | Wails v2, renderização a 60 FPS e controles de câmera interativos. |
| ⚡ **[Ponte Wails IPC](wails_ipc_and_bridges.md)** | Comunicação Go/JS | Métodos RPC, DTOs tipados e barramento de eventos assíncronos. |
| 🛠️ **[Guias do Desenvolvedor](developer_guides.md)** | Engenharia | Ambiente de build, convenções de código e extensão de serviços. |
| 🧪 **[Testes & Validação](testing_and_validation.md)** | Qualidade de Software | Cobertura de testes unitários, benchmarks e testes de integração. |
| 🚀 **[Deploy & Empacotamento](deployment_and_packaging.md)** | Operações | Build nativo para Linux e Windows com empacotamento de assets. |
| 📖 **[Referência de APIs](api_and_protocols_reference.md)** | Especificações | Contratos Wails, schemas JSON de telemetria e catálogo de OIDs. |

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Engenharia de Mobilidade Inteligente • EVP Simulator v1.0.0</i>
</div>
