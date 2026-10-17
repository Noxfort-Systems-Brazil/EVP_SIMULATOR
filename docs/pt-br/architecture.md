---
tags: [architecture, go, clean-architecture, ddd, ports-and-adapters, concurrency]
aliases: [System Architecture, Arquitetura do Sistema, Arquitetura Go]
---

# 🏛️ Arquitetura do Sistema Go e Modelo de Concorrência

Este documento detalha os fundamentos de engenharia do **EVP Simulator**. A aplicação é construída em **Go 1.22** seguindo os princípios de **Clean Architecture**, **Domain-Driven Design (DDD)** e **Arquitetura Hexagonal (Ports & Adapters)**, garantindo desacoplamento total entre as regras cinemáticas de tráfego e as tecnologias de renderização e rede.

⬅️ [Central de Documentação](README.md) | 🗺️ [Master MOC](evp_moc.md) | 🚨 [Walkthrough EVP](evp_walkthrough.md) | ⚡ [Ponte Wails](wails_ipc_and_bridges.md)

---

## 1. Princípios de Arquitetura Hexagonal

O núcleo da simulação não possui dependência de bibliotecas externas, frameworks de interface gráfica ou protocolos de rede:

```mermaid
graph TD
    subgraph DrivingAdapters ["🔌 Adaptadores Primários (Driving)"]
        WailsUI["Wails v2 Desktop GUI (HTML5/JS)"]
        CmdLine["Testes Automatizados / CLI"]
    end

    subgraph AppCore ["⚙️ Núcleo da Aplicação"]
        AppBridge["internal/infrastructure/wails/App"]
        SimUseCase["usecase.SimulationUseCase"]
        TelDisp["usecase.TelemetryDispatcher"]
    end

    subgraph DomainCore ["🧠 Domínio Puro (Sem Dependências)"]
        Entities["Entidades: RoadNetwork, Vehicle, TrafficLight"]
        Services["Serviços: KinematicsEngine, SignalEvaluator, AStarRouter"]
        Ports["Portas (Interfaces): NetworkRepo, ControllerRepo, TelemetryExporter"]
    end

    subgraph DrivenAdapters ["🔌 Adaptadores Secundários (Driven)"]
        SumoParser["Parser SUMO (sumo.Parser)"]
        SNMPServers["Servidores SNMP UTMC2 (utmc.ServerManager)"]
        CarinaClient["Cliente HTTP CARINA (httpclient.CarinaExporter)"]
        CSVFiles["Importador/Exportador CSV (csv.ControllerImporter)"]
    end

    WailsUI --> AppBridge
    CmdLine --> SimUseCase
    AppBridge --> SimUseCase
    SimUseCase --> Entities
    SimUseCase --> Services
    SimUseCase --> Ports
    SumoParser -.->|Implementa| Ports
    SNMPServers -.->|Implementa| Ports
    CarinaClient -.->|Implementa| Ports
    CSVFiles -.->|Implementa| Ports
```

---

## 2. Modelo de Concorrência e Ciclo de Vida da Simulação

A simulação é orientada a passos de tempo discretos (*ticks*), executando em uma goroutine dedicada desacoplada do loop de eventos da interface gráfica:

```mermaid
sequenceDiagram
    autonumber
    actor Operador as 👤 Operador Desktop
    participant UI as 🖥️ Wails Frontend (app.js)
    participant Bridge as 🌉 AppBridge (Go)
    participant UC as ⚙️ SimulationUseCase
    participant Kin as 🚗 KinematicsEngine
    participant Eval as 🚦 SignalEvaluator
    participant CARINA as 📡 CARINA CORE (HTTP)

    Operador->>UI: Clica em "Iniciar Simulação"
    UI->>Bridge: App.StartSimulation()
    Bridge->>UC: uc.Start()
    Note over UC: Dispara loop em background (100 ms / 10 Hz)
    loop Cada Tick (dt = 0.1s)
        UC->>Kin: UpdateVehicle(dt)
        Kin-->>UC: Novas posições e velocidades
        UC->>Eval: EvaluateApproach(vehicle, network)
        Eval-->>UC: Comandos de Preempção (Hold/Advance)
        UC->>UI: runtime.EventsEmit("telemetry_tick", DTO)
        opt Veículo em Emergência Ativa
            UC->>CARINA: POST /api/v1/evp/telemetry (Assíncrono)
        end
    end
```

### 2.1 Proteção de Estado e Thread-Safety
O estado global da simulação (`SimulationState`) é protegido por um mutex de leitura/escrita (`sync.RWMutex`):
- Operações de projeção de telemetria para a UI e exportação de dados utilizam `RLock()`, permitindo leituras concorrentes sem travar a interface.
- Operações de avanço de tick, spawn de veículos e reinicialização utilizam `Lock()`, garantindo mutação atômica e eliminando *data races*.

---

## 3. Desacoplamento de I/O de Rede via Buffer Não-Bloqueante

Para evitar que a latência de requisições HTTP externas para o backend CARINA afete a taxa de quadros (60 FPS) e o passo de física da simulação (10 Hz), o `TelemetryDispatcher` opera com fila delimitada assíncrona:
- As amostras de telemetria são enfileiradas em um canal Go com *buffer* de 1.000 mensagens.
- Uma goroutine trabalhadora dedicada consome do canal e despacha requisições HTTP via `POST /api/v1/evp/telemetry`.
- Caso o canal atinja o limite por instabilidade na rede local, mensagens excedentes não-críticas são descartadas com log de aviso (*load shedding*), mantendo a simulação rigorosamente fluida.

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Engenharia de Mobilidade Inteligente • EVP Simulator v1.0.0</i>
</div>
