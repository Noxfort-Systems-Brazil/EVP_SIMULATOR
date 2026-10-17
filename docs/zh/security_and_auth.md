---
tags: [security, auth, bcrypt, users, rbac, 12-factor]
aliases: [安全机制与身份认证, 用户管理, 权限控制, 安全管理器]
---

# 🔐 安全架构、身份认证与用户权限管理

本文档详尽规范了 CARINA 位于 [`src/utils/security/`](../../src/utils/security) 与 [`src/utils/security_manager.py`](../../src/utils/security_manager.py) 的企业级系统安全架构。深入阐述了用户身份认证、基于密码学的加盐散列密码存储、防暴力破解锁定机制、数据库模式权限隔离以及遵循 12-Factor 规范的敏感机密管理体系。

⬅️ 返回 [核心技术主文档中心 (MOC)](carina_moc.md) | 🗄️ 查阅 [数据库与模式](database_and_schemas.md) | 🛠️ 查阅 [开发者指南](developer_guides.md)

---

## 1. 系统安全架构全景概览

CARINA 直接管控城市关键市政交通基础设施。任何未经授权的非法入侵或越权提权行为，均可能导致物理交通机失控，造成严重的交通事故碰撞或城市路网瘫痪。为防御各类网络与系统安全威胁，系统强制实施以下安全基线：

1. **确定性强密码哈希：** 基于可配置工作因子的加盐 `bcrypt` 现代密码学散列算法。
2. **防暴力破解防御机制 (Lockdown Engine)：** 针对 IP 与用户名实施多层级速率限制与指数级冷却退避。
3. **基于角色的细粒度访问控制 (RBAC)：** 严格隔离 `OPERATOR`（调度操作员）、`ENGINEER`（配时工程师）与 `ADMIN`（系统管理员）的操作权限。
4. **12-Factor 机密彻底隔离：** 杜绝在任何源代码中硬编码机密凭据，所有敏感信息统一由 `.env` 环境注入。

```text
 客户端终端 (Flet 桌面 UI / CLI 运维工具)
          │
          ▼
 ┌──────────────────┐      校验身份安全凭据       ┌──────────────────┐
 │   AuthService    │ ───────────────────────────> │  Bcrypt Hasher   │
 └────────┬─────────┘                              └──────────────────┘
          │
          ├──> [防爆破安全校验] ──> LockdownManager (最多允许 5 次错误尝试，锁定 15 分钟)
          │
          └──> [用户信息检索]   ──> UserRepository (PostgreSQL / SQLite 隔离数据库模式)
```

---

## 2. `src/utils/security/` 核心安全模块

### 2.1 密码哈希与密码学引擎 (`hasher.py`)
位于 [`src/utils/security/hasher.py`](../../src/utils/security/hasher.py)：
- 基于国际标准 `bcrypt` 算法实现高强度密码散列。
- 自动生成高熵随机盐（默认迭代轮数 $2^{12}$ 轮）。
- 采用恒定时间（Constant-time）字符串比较逻辑，彻底免疫基于时序分析的侧信道旁路攻击（Timing Attacks）。

### 2.2 防暴力破解锁定引擎 (`lockdown.py`)
位于 [`src/utils/security/lockdown.py`](../../src/utils/security/lockdown.py)：
- 独立跟踪每个用户名及客户端来源 IP 的连续认证失败次数。
- 连续失败达到 5 次立即触发自动化账户安全锁定。
- 强制实施阶梯式惩罚冷却期（默认锁定冷却时长：15 分钟）。

### 2.3 用户管理服务与仓储层 (`user_service.py` 与 `user_repository.py`)
- [`src/utils/security/user_repository.py`](../../src/utils/security/user_repository.py)：通过底层原生 SQL 事务，持久化存储用户信息、活跃会话状态以及审计操作日志。
- [`src/utils/security/migrator.py`](../../src/utils/security/migrator.py)：在 SQLite 与 PostgreSQL 数据库中自动初始化并迁移 `users` 及 `audit_log` 基础安全模式表。

---

## 3. 12-Factor 原则配置与环境变量规范

所有生产级通信机密必须且只能在根目录的 [`.env`](../../.env.example) 文件中声明，严禁提交至公共 Git 版本控制仓库：

```bash
# --- 运行环境模式 ---
# 'development': 仅用于开发调试，允许使用默认演示凭据 (admin / Admin@Carina2026!)。
# 'production': 生产环境严格强制禁用默认弱口令，强制校验强哈希机密。
CARINA_ENV=production

# --- 超级管理员初始化 ---
CARINA_SUPERUSER_USER=admin_traffic
CARINA_SUPERUSER_PASSWORD=YourStrongPasswordHere!
CARINA_SUPERUSER_HASH=$2b$12$e8x...

# --- 数据库安全凭据 ---
CARINA_DB_USER=carina_sec
CARINA_DB_PASSWORD=database_secure_password
CARINA_DB_HOST=127.0.0.1
CARINA_DB_PORT=5432
CARINA_DB_NAME=carina_data
CARINA_DB_SCHEMA=schema_carina

# --- SNMP 物理交通机通信机密 ---
CARINA_SNMP_COMMUNITY=YourSnmpCommunityString

# --- 外部遥测与监控通道凭据 ---
CARINA_MQTT_HOST=127.0.0.1                     # 本地 Broker 或 https://tunnel.ngrok-free.dev/api/telemetry
CARINA_MQTT_PORT=1883
CARINA_MQTT_USER=
CARINA_MQTT_PASSWORD=
```

---

## 4. 生产环境安全加固工程指南

1. **轮换 SNMP 默认团体名：** 在现场物理交通机部署中，绝对禁止使用出厂默认字符串 `"public"`。
2. **收紧 IPC 与 gRPC 监听范围：** 在生产运行中，gRPC（`50051` 端口）与 WebSocket（`8080` 端口）应绑定在专用管理私有 VLAN 或 `127.0.0.1` 本地回环接口，并通过具备 TLS 证书终结的反向代理对外暴露。
3. **数据库最小权限原则：** 为 CARINA 业务分配的数据库账号仅应具备在自身专用 schema 下的 `SELECT`、`INSERT` 与 `UPDATE` 权限，杜绝赋予超管或 DDL 破坏性权限。
4. **现场控制器故障安全底线：** 现场部署前必须逐个确认所有物理交通机均已预设并固化了本地自主日时段配时方案，确保在通信中断时能够瞬间无缝降级保底。

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>智能交通系统工程 • CARINA CORE v1.2.0</i><br/>
  <small>© 2026 Noxfort Systems. 基于 AGPLv3 授权。</small>
</div>
