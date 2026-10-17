---
tags: [home, index, evp, simulator, hub]
aliases: [Portal de Documentação, Overview, EVP Simulator Hub]
---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="70" /><br/>
  <h1>EVP Simulator — Portal Técnico & Hub de Simulação</h1>
  <h3>Simulador Aberto de Preempção Semafórica e Frotas de Emergência em Redes Urbanas</h3>
  <i>Noxfort Systems — A State Of Art Company</i>
</div>

---

## 🚦 Visão Geral do Sistema

O **EVP Simulator** é uma plataforma de simulação microscópica de alto desempenho desenvolvida em **Go 1.22** com interface desktop nativa em **Wails v2** e renderização vetorial 2D em **HTML5 Canvas**. O sistema foi concebido para modelar, avaliar e validar estratégias de **Preempção Semafórica para Veículos de Emergência (EVP - Emergency Vehicle Preemption)** em malhas viárias urbanas complexas.

Projetado em conformidade com as diretrizes do **CONTRAN (Resolução nº 960)**, do **Código de Trânsito Brasileiro (CTB - Art. 29, VII)** e normas internacionais de ITS (**FHWA MUTCD Seção 4D.26**), o simulador reproduz o comportamento cinemático realista de viaturas, calcula corredores de onda verde prioritários e integra-se diretamente com o ecossistema inteligente de controle semafórico **CARINA CORE**.

---

## 🗺️ Mapa de Navegação e Subsistemas

| Subsistema / Dimensão | Foco Técnico | Documento |
| :--- | :--- | :---: |
| 📚 **Master Map of Content** | Grafo Obsidian e Mapa de Módulos da Base de Código Go | [Explorar MOC](evp_moc.md) |
| 🏛️ **Arquitetura do Sistema** | Clean Architecture, Hexagonal / Ports & Adapters e Goroutines | [Ver Arquitetura](architecture.md) |
| 🚨 **Walkthrough de Preempção** | Detecção de Entrada ($60\text{--}160\text{ m}$), Checkout ($15\text{ m}$) e Watchdog | [Ver Preempção](evp_walkthrough.md) |
| 🚗 **Cinemática & Roteamento A\*** | Motor Cinemático de Euler, Modos Veiculares e A\* em Grafos | [Ver Cinemática](kinematics_and_routing.md) |
| 🗺️ **Parser de Redes SUMO** | Ingestão de `.net.xml` e `.net.xml.gz`, Projeção Cartesiana/WGS84 | [Ver Parser SUMO](sumo_network_parser.md) |
| 🔌 **Emulação SNMP / UTMC** | Servidores UDP SNMP em Go, Protocolo UTMC2 e NTCIP 1202 | [Ver Emulador SNMP](utmc_and_snmp_emulation.md) |
| 📡 **Integração com CARINA** | Despacho HTTP Fire-and-Forget `POST /api/v1/evp/telemetry` | [Ver Integração](carina_integration_and_telemetry.md) |
| 🚑 **Gerenciamento de Frotas** | Geração Procedural de Frotas (SAMU, Bombeiros, Polícia) | [Ver Frotas](fleet_management_and_spawner.md) |
| 🛡️ **Segurança & Avaliação Semafórica** | Avaliação de Fases, Tempo de Limpeza (Amarelo/All-Red) e Vetos | [Ver Segurança](signal_evaluation_and_safety.md) |
| 📊 **Importação / Exportação CSV** | Inventário de Controladores e Exportação de Trajetórias | [Ver CSV Exchanger](csv_data_exchanger.md) |
| 🖥️ **Interface & Canvas 2D** | Frontend Desktop Wails v2, Renderizador 60 FPS e Câmera | [Ver Interface](ui_and_canvas_rendering.md) |
| ⚡ **Ponte Wails IPC (Go ↔ JS)** | Métodos RPC, DTOs de Estado e Barramento de Eventos | [Ver Wails IPC](wails_ipc_and_bridges.md) |
| 🛠️ **Guias do Desenvolvedor** | Configuração de Ambiente, Go 1.22+, Wails e Boas Práticas | [Ver Guias](developer_guides.md) |
| 🧪 **Testes & Validação** | Testes Unitários de Cinemática, Roteamento e Mocks SNMP | [Ver Testes](testing_and_validation.md) |
| 🚀 **Deploy & Empacotamento** | Compilação Nativa (`wails build`), Linux e Windows | [Ver Deploy](deployment_and_packaging.md) |
| 📖 **Referência de APIs & Protocolos** | Assinaturas RPC Wails, JSON Telemetria e OIDs SNMP | [Ver APIs](api_and_protocols_reference.md) |

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Engenharia de Mobilidade Inteligente • EVP Simulator v1.0.0</i>
</div>
