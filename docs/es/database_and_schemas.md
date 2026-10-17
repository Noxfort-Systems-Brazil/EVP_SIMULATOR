---
tags: [base-de-datos, postgresql, sqlite, esquemas, persistencia, compresion-delta, telemetria, 12-factor]
aliases: [Arquitectura de Base de Datos, Esquemas Relacionales, Capa de Persistencia, Decisiones de Paso]
---

# 🗄️ Arquitectura de Base de Datos y Esquemas Relacionales

Este documento detalla la capa de persistencia de CARINA, el pool de conexiones, el agrupamiento asíncrono no bloqueante de telemetría ([`StepDecisionWorker`](../../src/database/step_decision_worker.py), [`FluidDynamicsWriter`](../../src/repositories/fluid_dynamics_writer.py)), las credenciales según el estándar de aplicaciones de 12 factores y los esquemas de tablas relacionales para PostgreSQL y SQLite.

⬅️ Volver al [Hub Maestro de Documentación](carina_moc.md) | 🚨 Ver [Preempción EVP](evp_and_preemption.md) | 🔐 Ver [Seguridad y Autenticación](security_and_auth.md) | 🚦 Ver [Controladores de Hardware](hardware_drivers.md) | 🛠️ Ver [Guías para Desarrolladores](developer_guides.md)

---

## 1. Arquitectura Asíncrona de Base de Datos No Bloqueante

Para evitar que las operaciones síncronas `INSERT` de SQL bloqueen el bucle de decisión en tiempo real de la IA en sub-milisegundos (meta $< 1\text{ ms}$), CARINA enruta toda la telemetría operativa a través de colas de memoria acotadas y workers de procesamiento por lotes en segundo plano.

```text
Bucle_Decision_TiempoReal (< 0.001 ms Push en RAM)
  └──> [Cola en Memoria] ──> StepDecisionWorker / FluidDynamicsWriter
                                      │
                                      ▼
                           Compresión Delta (Reducción del 97.9%)
                                      │
                                      ▼
                           Carga Masiva COPY / execute_values en PostgreSQL
```

- **Latencia de Ejecución:** Sobrecarga $< 0.001\text{ ms}$ por push en la cola de memoria de Python.
- **Tamaño del Lote:** 50 registros o tiempo máximo de vaciado (flush) de 3.0 segundos.
- **Compresión Delta:** Agrupa estados de telemetría idénticos consecutivos para reducir la huella de almacenamiento en un **97.9%** (~380 MB/día para 200 intersecciones).

---

## 2. Configuración de Base de Datos según App de 12 Factores

De acuerdo con los principios de 12-Factor App implementados en [`src/database/db_engine.py`](../../src/database/db_engine.py), todas las credenciales de base de datos se cargan directamente desde el archivo de entorno [`.env`](../../.env.example):

```bash
CARINA_DB_USER=admin
CARINA_DB_PASSWORD=secret_password
CARINA_DB_HOST=localhost
CARINA_DB_PORT=5432
CARINA_DB_NAME=carina_data
CARINA_DB_SCHEMA=schema_carina
```

---

## 3. Esquemas de Base de Datos Relacional

### 3.1 Tabla: `step_decisions`
Almacena sugerencias de agentes en tiempo real, vetos del Guardian, decisiones y temporizadores de paso con codificación Enum Smallint de 1 byte.

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
    step_count INTEGER NOT NULL DEFAULT 1,          -- Conteo Consecutivo Compresión Delta
    total_step_time_ms REAL,
    guardian_time_ms REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sd_agent_time ON step_decisions (agent_id, created_at DESC);
CREATE INDEX idx_sd_final_decision ON step_decisions (final_decision, veto_reason_code);
```

### 3.2 Tabla: `edge_dictionary`
Asocia identificadores largos de vía en formato string (`"topolondrina_via_jk_norte_12"`) con enteros compactos de 4 bytes para almacenamiento de alta densidad.

```sql
CREATE TABLE IF NOT EXISTS edge_dictionary (
    edge_int_id SERIAL PRIMARY KEY,
    edge_str_id VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_edge_dict_str ON edge_dictionary (edge_str_id);
```

### 3.3 Tabla: `synapse_fluid_dynamics`
Almacena lecturas de sensores de alta frecuencia por segmento vial, comprimidas mediante Compresión Delta.

```sql
CREATE TABLE IF NOT EXISTS synapse_fluid_dynamics (
    id SERIAL PRIMARY KEY,
    collected_at TIMESTAMP NOT NULL DEFAULT NOW(),
    scenario_name TEXT NOT NULL DEFAULT 'default',
    intersection_id TEXT,
    edge_id TEXT NOT NULL,
    edge_int_id INTEGER,
    density REAL NOT NULL,
    mean_speed REAL NOT NULL,
    min_speed REAL,
    queue_length INTEGER NOT NULL,
    max_queue INTEGER,
    occupancy REAL NOT NULL,
    edge_length REAL,
    num_lanes INTEGER,
    speed_limit REAL,
    maturity_stage TEXT NOT NULL DEFAULT 'CHILD',
    sample_count INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_sfd_collected_at ON synapse_fluid_dynamics (collected_at);
CREATE INDEX idx_sfd_edge_id ON synapse_fluid_dynamics (edge_id);
CREATE INDEX idx_sfd_scen_stage_time ON synapse_fluid_dynamics (scenario_name, maturity_stage, collected_at DESC);
```

### 3.4 Tabla: `hardware_controller_connections`
Almacena especificaciones de conexión con controladores físicos (endpoints NTCIP/UTMC).

```sql
CREATE TABLE IF NOT EXISTS hardware_controller_connections (
    id SERIAL PRIMARY KEY,
    intersection_id VARCHAR(64) UNIQUE NOT NULL,
    protocol VARCHAR(32) NOT NULL,                  -- 'NTCIP_1202', 'UTMC', 'SNMP'
    ip_address VARCHAR(45) NOT NULL,
    port INTEGER NOT NULL DEFAULT 161,
    snmp_community VARCHAR(64) DEFAULT 'public',
    timeout_seconds REAL DEFAULT 1.0,
    retry_count INTEGER DEFAULT 3,
    status VARCHAR(32) DEFAULT 'DISCONNECTED',
    last_heartbeat TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3.5 Tabla: `users` (Capa de Seguridad)
Almacena cuentas de operadores y credenciales con hash administradas por [`src/utils/security/user_repository.py`](../../src/utils/security/user_repository.py).

```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(64) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'OPERATOR',   -- 'ADMIN', 'ENGINEER', 'OPERATOR'
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3.6 Tabla: `evp_registered_fleet` (Preempción de Vehículos de Emergencia)
Almacena vehículos de emergencia municipales autorizados administrados por [`src/evp/evp_fleet_repository.py`](../../src/evp/evp_fleet_repository.py).

```sql
CREATE TABLE IF NOT EXISTS evp_registered_fleet (
    vehicle_id VARCHAR(64) PRIMARY KEY,
    corporation VARCHAR(32) NOT NULL,              -- 'SAMU', 'BOMBEIRO', 'POLICIA'
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

### 3.7 Tabla: `evp_telemetry_records` (Auditoría Forense de Telemetría)
Almacena trayectorias históricas de balizas GPS y eventos de preempción emitidos por unidades de emergencia administradas por [`src/evp/evp_telemetry_repository.py`](../../src/evp/evp_telemetry_repository.py).

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

### 3.8 Tabla: `evp_denied_fleet` (Lista de Bloqueo de Auto-Registro)
Mantiene identificadores de transpondedores de vehículos no autorizados o en lista negra administrados por [`src/evp/evp_denied_repository.py`](../../src/evp/evp_denied_repository.py).

```sql
CREATE TABLE IF NOT EXISTS evp_denied_fleet (
    vehicle_id VARCHAR(64) PRIMARY KEY,
    denied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 4. Benchmark de Eficiencia de Almacenamiento

| Tabla / Sector | Tamaño Sin Comprimir | Tamaño Comprimido | Ahorro de Almacenamiento |
| :--- | :--- | :--- | :--- |
| **`step_decisions`** | $\approx 3.45\text{ GB / día}$ | **$\approx 0.02\text{ GB / día}$** | **$-99.4\%$** |
| **`synapse_fluid_dynamics`** | $\approx 14.00\text{ GB / día}$ | **$\approx 0.15\text{ GB / día}$** | **$-98.9\%$** |
| **`evp_telemetry_records`** | $\approx 0.50\text{ GB / día}$ | **$\approx 0.03\text{ GB / día}$** | **$-94.0\%$** |
| **`hardware_controller_connections`** | $\approx 0.20\text{ GB / día}$ | **$\approx 0.01\text{ GB / día}$** | **$-95.0\%$** |
| **BASE DE DATOS TOTAL DEL SISTEMA** | **$\approx 18.65\text{ GB / día}$** | **$\approx 0.41\text{ GB / día}$** | **$-97.8\%$!** |

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Ingeniería de Movilidad Inteligente • CARINA CORE v1.2.0</i><br/>
  <small>© 2026 Noxfort Systems. Licenciado bajo AGPLv3.</small>
</div>
