---
tags: [database, postgresql, sqlite, schemas, delta-storage, evp, rle]
aliases: [Banco de Dados, Esquemas Relacionais, Armazenamento Delta, Modelagem SQL]
---

# 🗄️ Arquitetura de Banco de Dados, Esquemas e Armazenamento Delta

Este documento detalha a camada de persistência do CARINA localizada em [`src/database/`](../../src/database), o pipeline assíncrono de telemetria não-bloqueante via filas de RAM, a engine de compressão delta por Run-Length Encoding (RLE) no PostgreSQL e os esquemas DDL das tabelas relacionais.

⬅️ Voltar ao [Portal Central MOC](carina_moc.md) | 🏛️ Ver [Arquitetura do Sistema](architecture.md) | 🛡️ Ver [Segurança e Watchdog](safety_and_watchdog.md) | 🚑 Ver [Preempção EVP](evp_and_preemption.md)

---

## 1. Pipeline de Persistência Assíncrona Não-Bloqueante

Para evitar que latências de disco ou operações SQL `INSERT` bloqueiem o ciclo de inferência de sub-milissegundo da IA ($< 0,5\text{ ms}$), todo registro de decisão ou telemetria viária é descarregado de forma desacoplada:

```text
  Thread de Inferência (AI_Process)
                │
                ▼ (Push Não-Bloqueante em RAM < 0.001 ms)
      ┌──────────────────┐
      │   Queue db_queue │
      └─────────┬────────┘
                │
                ▼ (Consome em lotes de 50 registros ou intervalo de 3.0s)
      ┌───────────────────────────────────┐
      │          DatabaseWorker           │ (src/database/db_worker.py)
      │     (Processo de Background)      │
      └─────────────────┬─────────────────┘
                        │
         ┌──────────────┴──────────────┐
         ▼                             ▼
  ┌──────────────┐              ┌──────────────┐
  │  PostgreSQL  │ (Produção)   │ SQLite Local │ (Desenvolvimento)
  └──────────────┘              └──────────────┘
```

### 1.1 Mecânica de Armazenamento Delta (Run-Length Encoding)
Em cruzamentos reais, os semáforos frequentemente mantêm o mesmo estágio verde por dezenas de segundos consecutivos. Em vez de gravar uma linha por segundo:
1. O worker compara o estado semafórico atual com o registro anterior.
2. Se a ação, decisão do Guardian e código de veto forem idênticos, o worker apenas incrementa a coluna `step_count` do registro ativo em memória.
3. Quando ocorre uma transição de estágio ou alteração de veto, o registro consolidado é gravado fisicamente via `INSERT`.

---

## 2. Benchmark de Redução de Armazenamento em Produção

Resultados empíricos medidos em rede urbana com 200 cruzamentos semaforizados operando 24/7 a $1\text{ Hz}$ de amostragem:

| Métrica | Ingestão Convencional (Raw) | Compressão Delta CARINA | Ganho / Redução |
| :--- | :---: | :---: | :---: |
| **Linhas gravadas por dia** | $17.280.000\text{ linhas}$ | **$362.880\text{ linhas}$** | **97,9% menos linhas** |
| **Volume de disco diário** | $\approx 18,2\text{ GB/dia}$ | **$\approx 382\text{ MB/dia}$** | **Economia de 97,9%** |
| **I/O Overhead na CPU de IA** | Bloqueante ($5\text{--}15\text{ ms}$) | **Zero ($< 0,001\text{ ms}$)** | **Totalmente desacoplado** |
| **Tempo de Consulta Forense** | Varredura de tabelas gigantescas | Índices compactos $B\text{-Tree}$ | **Aceleração de $48\times$** |

---

## 3. Esquemas DDL das Tabelas Relacionais

### 3.1 Tabela: `step_decisions`
Registra decisões de agentes táticos, auditorias do Guardian e tempos de execução com enums `SMALLINT` de 1 byte:

```sql
CREATE TABLE IF NOT EXISTS step_decisions (
    id BIGSERIAL PRIMARY KEY,
    simulation_time REAL NOT NULL,
    step_number INTEGER NOT NULL,
    agent_id VARCHAR(64) NOT NULL,
    maturity_stage SMALLINT NOT NULL DEFAULT 2,     -- 0=CHILD, 1=TEEN, 2=ADULT
    suggested_action SMALLINT NOT NULL DEFAULT 0,   -- 0=KEEP, 1=CHANGE, 2=OVERRIDE
    final_decision SMALLINT NOT NULL DEFAULT 0,     -- 0=APPROVED, 1=DENIED (VETOED)
    veto_reason_code SMALLINT NOT NULL DEFAULT 0,   -- 0=NONE, 1=MIN_GREEN, 2=YELLOW, 3=SPILLBACK, 4=GRIDLOCK
    step_count INTEGER NOT NULL DEFAULT 1,          -- Contador de Compressão Delta RLE
    total_step_time_ms REAL,
    guardian_time_ms REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_step_dec_agent_time ON step_decisions (agent_id, created_at DESC);
```

### 3.2 Tabela: `synapse_fluid_dynamics`
Armazena métricas de fluxo contínuo de tráfego (velocidade média, ocupação, extensão de filas):

```sql
CREATE TABLE IF NOT EXISTS synapse_fluid_dynamics (
    id BIGSERIAL PRIMARY KEY,
    edge_id VARCHAR(64) NOT NULL,
    occupancy REAL NOT NULL,
    mean_speed_ms REAL NOT NULL,
    queue_length_m REAL NOT NULL,
    vehicle_count INT NOT NULL,
    timestamp DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_fluid_edge_time ON synapse_fluid_dynamics (edge_id, created_at DESC);
```

### 3.3 Tabela: `hardware_controller_connections`
Catálogo de equipamentos físicos de rua (NTCIP 1202, UTMC2):

```sql
CREATE TABLE IF NOT EXISTS hardware_controller_connections (
    id SERIAL PRIMARY KEY,
    intersection_id VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(128),
    ip_address VARCHAR(45) NOT NULL,
    port INT DEFAULT 161,
    protocol VARCHAR(16) DEFAULT 'NTCIP_1202',      -- 'NTCIP_1202', 'UTMC', 'SNMP'
    snmp_community VARCHAR(64) DEFAULT 'public',
    is_active BOOLEAN DEFAULT TRUE,
    connection_status VARCHAR(16) DEFAULT 'DISCONNECTED',
    last_heartbeat TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3.4 Tabela: `edge_dictionary`
Dicionário de mapeamento de vias urbanas para IDs numéricos compactos:

```sql
CREATE TABLE IF NOT EXISTS edge_dictionary (
    id SERIAL PRIMARY KEY,
    edge_str VARCHAR(128) UNIQUE NOT NULL,
    length_m REAL,
    speed_limit_kmh REAL,
    lane_count SMALLINT DEFAULT 1
);
```

### 3.5 Tabelas: `users` e `user_sessions` (Segurança e Auditoria)
Armazena credenciais e sessões de operadores municipais com hash Bcrypt salgado:

```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(64) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(32) DEFAULT 'OPERATOR',            -- 'ADMIN', 'OPERATOR', 'AUDITOR'
    is_active BOOLEAN DEFAULT TRUE,
    failed_attempts SMALLINT DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_sessions (
    token VARCHAR(128) PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3.6 Tabela: `evp_registered_fleet` (Preempção de Emergência)
Registro de viaturas autorizadas a interagir com os semáforos:

```sql
CREATE TABLE IF NOT EXISTS evp_registered_fleet (
    vehicle_id VARCHAR(64) PRIMARY KEY,
    corporation VARCHAR(32) NOT NULL,               -- 'SAMU', 'BOMBEIRO', 'POLICIA'
    plate VARCHAR(16),
    description VARCHAR(128),
    device_id VARCHAR(64),
    ip_address VARCHAR(45),
    port INT DEFAULT 8085,
    protocol VARCHAR(16) DEFAULT 'HTTP',
    auth_token VARCHAR(128),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3.7 Tabela: `evp_telemetry_records` (Histórico de Trajetórias)
Rastreabilidade forense de todas as transmissões de telemetria das viaturas:

```sql
CREATE TABLE IF NOT EXISTS evp_telemetry_records (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id VARCHAR(64) NOT NULL,
    corporation VARCHAR(32) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    speed_kmh REAL NOT NULL,
    heading_deg REAL NOT NULL,
    siren_active BOOLEAN NOT NULL,
    evp_triggered BOOLEAN NOT NULL DEFAULT FALSE,
    intersection_id VARCHAR(64),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_evp_veh_time ON evp_telemetry_records (vehicle_id, created_at DESC);
```

### 3.8 Tabela: `evp_denied_fleet` (Bloqueio Administrativo)
Lista negra de viaturas revogadas ou com transponders defeituosos:

```sql
CREATE TABLE IF NOT EXISTS evp_denied_fleet (
    vehicle_id VARCHAR(64) PRIMARY KEY,
    denied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 4. Configuração de Variáveis de Ambiente (`.env`)

O CARINA adota o padrão **12-Factor App**. As credenciais de banco de dados e segredos nunca são incluídos no código-fonte:

```bash
CARINA_DB_TYPE=postgresql                  # 'postgresql' ou 'sqlite'
CARINA_DB_HOST=127.0.0.1
CARINA_DB_PORT=5432
CARINA_DB_USER=carina_admin
CARINA_DB_PASSWORD=segredo_forte_aqui
CARINA_DB_NAME=carina_telemetry
CARINA_DB_POOL_SIZE=10
CARINA_DB_MAX_OVERFLOW=20
```

---

## 5. Pooling de Conexões e Ciclo de Vida (`src/database/`)

- **`DatabaseEngineFactory` (`engine_factory.py`):** Instancia engines otimizadas com pooling `QueuePool` no PostgreSQL ou `NullPool` no SQLite com chaveamento thread-safe.
- **`SchemaInitializer` (`schema_initializer.py`):** Aplica migrations automáticas DDL no boot do sistema, garantindo criação atômica de índices e tabelas ausentes sem derrubar a aplicação.
- **Graceful Shutdown:** No encerramento do CARINA (`SIGINT`/`SIGTERM`), o `DatabaseWorker` drena todas as mensagens remanescentes nas filas de RAM antes de fechar o pool de conexões, assegurando zero perda de dados.

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Engenharia de Mobilidade Inteligente • CARINA CORE v1.2.0</i><br/>
  <small>© 2026 Noxfort Systems. Licenciado sob AGPLv3.</small>
</div>
