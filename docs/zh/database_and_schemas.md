---
tags: [database, postgresql, sqlite, schema, persistence, delta-compression, telemetry, 12-factor]
aliases: [数据库架构, 关系模式, 持久化层, 步进决策存储]
---

# 🗄️ 数据库架构与关系数据库模式

本文档详尽规范了 CARINA 的数据持久化层、数据库连接池技术、异步非阻塞遥测批量写入机制（[`StepDecisionWorker`](../../src/database/step_decision_worker.py)、[`FluidDynamicsWriter`](../../src/repositories/fluid_dynamics_writer.py)）、12-Factor 原则配置凭据管理，以及面向 PostgreSQL 与 SQLite 的规范关系型数据表模式。

⬅️ 返回 [核心技术主文档中心 (MOC)](carina_moc.md) | 🚨 查阅 [EVP 优先通行](evp_and_preemption.md) | 🔐 查阅 [安全机制与身份认证](security_and_auth.md) | 🚦 查阅 [硬件驱动程序](hardware_drivers.md) | 🛠️ 查阅 [开发者指南](developer_guides.md)

---

## 1. 异步非阻塞数据库写入架构

为了彻底防止同步 SQL `INSERT` 磁盘 I/O 阻塞亚毫秒级的人工智能实时决策主循环（目标延迟 $< 1\text{ ms}$），CARINA 将所有底层运行遥测通过有界内存队列路由至后台批量处理工作进程：

```text
实时决策主循环 (< 0.001 ms RAM 快速入队)
  └──> [有界内存队列] ──> StepDecisionWorker / FluidDynamicsWriter
                                      │
                                      ▼
                           增量差分压缩 (存储减少 97.9%)
                                      │
                                      ▼
                           PostgreSQL 批量 COPY / execute_values 写入
```

- **执行延迟：** 向 Python 内存队列推送数据的开销 $< 0.001\text{ ms}$。
- **批处理规模：** 攒满 50 条记录或达到 3.0 秒最大刷写超时时钟触发落盘。
- **增量差分压缩（Delta Compression）：** 聚合连续相同的时序遥测状态，将海量时序存储占用缩减 **97.9%**（对于包含 200 个路口的都市级路网，每天全量数据仅占用约 380 MB）。

---

## 2. 12-Factor 原则数据库配置

遵循在 [`src/database/db_engine.py`](../../src/database/db_engine.py) 中贯彻的 12-Factor App 架构原则，所有数据库连接敏感凭据均直接由 [`.env`](../../.env.example) 环境变量文件加载解析：

```bash
CARINA_DB_USER=admin
CARINA_DB_PASSWORD=secret_password
CARINA_DB_HOST=localhost
CARINA_DB_PORT=5432
CARINA_DB_NAME=carina_data
CARINA_DB_SCHEMA=schema_carina
```

---

## 3. 关系型数据库模式定义

### 3.1 数据表：`step_decisions`
存储实时智能体提议动作、Guardian 安全否决记录、最终执行决策以及单步执行计时器，采用 1 字节 Smallint 枚举进行极速索引与存储。

```sql
CREATE TABLE IF NOT EXISTS step_decisions (
    id BIGSERIAL PRIMARY KEY,
    simulation_time REAL NOT NULL,
    step_number INTEGER NOT NULL,
    agent_id VARCHAR(64) NOT NULL,
    maturity_stage SMALLINT NOT NULL DEFAULT 2,     -- 0=CHILD 幼年, 1=TEEN 青年, 2=ADULT 成年
    suggested_action SMALLINT NOT NULL DEFAULT 0,   -- 0=KEEP 维持, 1=CHANGE 切换, 2=OVERRIDE 覆盖
    final_decision SMALLINT NOT NULL DEFAULT 0,     -- 0=APPROVED 批准, 1=DENIED 否决
    veto_reason_code SMALLINT NOT NULL DEFAULT 0,   -- 0=NONE 无, 1=MIN_GREEN 最短绿灯, 2=YELLOW 黄灯, 3=SPILLBACK 溢流, 4=GRIDLOCK 锁死
    step_count INTEGER NOT NULL DEFAULT 1,          -- 增量差分压缩连续命中步数
    total_step_time_ms REAL,
    guardian_time_ms REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sd_agent_time ON step_decisions (agent_id, created_at DESC);
CREATE INDEX idx_sd_final_decision ON step_decisions (final_decision, veto_reason_code);
```

### 3.2 数据表：`edge_dictionary`
将冗长的路段全称字符串（如 `"topolondrina_via_jk_norte_12"`）映射为紧凑的 4 字节整数，实现高密度外键存储。

```sql
CREATE TABLE IF NOT EXISTS edge_dictionary (
    edge_int_id SERIAL PRIMARY KEY,
    edge_str_id VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_edge_dict_str ON edge_dictionary (edge_str_id);
```

### 3.3 数据表：`synapse_fluid_dynamics`
存储高频微观传感器在各道路断面采集的时序流体动力学指标，采用游程增量差分算法压缩。

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

### 3.4 数据表：`hardware_controller_connections`
安全持久化现场物理信号机硬件通信连接参数（NTCIP/UTMC 端点）。

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

### 3.5 数据表：`users` (系统安全权限层)
存储运维调度人员账户、密码哈希与角色权限，由 [`src/utils/security/user_repository.py`](../../src/utils/security/user_repository.py) 统一管理。

```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(64) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'OPERATOR',   -- 'ADMIN' 管理员, 'ENGINEER' 工程师, 'OPERATOR' 操作员
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3.6 数据表：`evp_registered_fleet` (特勤应急车队准入库)
存储经官方认证授权接入系统的应急特种车辆，由 [`src/evp/evp_fleet_repository.py`](../../src/evp/evp_fleet_repository.py) 统一管理。

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

### 3.7 数据表：`evp_telemetry_records` (特勤遥测司法级取证库)
存储应急特勤车辆上报的历史高精 GPS 定位轨迹与优先控制事件记录，由 [`src/evp/evp_telemetry_repository.py`](../../src/evp/evp_telemetry_repository.py) 统一管理。

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

### 3.8 数据表：`evp_denied_fleet` (未授权特勤终端黑名单)
维护未经授权或被系统拉黑的非法车载终端标识，由 [`src/evp/evp_denied_repository.py`](../../src/evp/evp_denied_repository.py) 统一管理。

```sql
CREATE TABLE IF NOT EXISTS evp_denied_fleet (
    vehicle_id VARCHAR(64) PRIMARY KEY,
    denied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 4. 存储效率基准测试

| 数据表 / 业务板块 | 未压缩原始体积 | 增量压缩后体积 | 存储空间缩减比 |
| :--- | :--- | :--- | :--- |
| **`step_decisions`** | $\approx 3.45\text{ GB / 天}$ | **$\approx 0.02\text{ GB / 天}$** | **$-99.4\%$** |
| **`synapse_fluid_dynamics`** | $\approx 14.00\text{ GB / 天}$ | **$\approx 0.15\text{ GB / 天}$** | **$-98.9\%$** |
| **`evp_telemetry_records`** | $\approx 0.50\text{ GB / 天}$ | **$\approx 0.03\text{ GB / 天}$** | **$-94.0\%$** |
| **`hardware_controller_connections`** | $\approx 0.20\text{ GB / 天}$ | **$\approx 0.01\text{ GB / 天}$** | **$-95.0\%$** |
| **全系统综合数据库** | **$\approx 18.65\text{ GB / 天}$** | **$\approx 0.41\text{ GB / 天}$** | **$-97.8\%$!** |

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>智能交通系统工程 • CARINA CORE v1.2.0</i><br/>
  <small>© 2026 Noxfort Systems. 基于 AGPLv3 授权。</small>
</div>
