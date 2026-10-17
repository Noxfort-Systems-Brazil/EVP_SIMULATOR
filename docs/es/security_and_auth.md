---
tags: [seguridad, autenticacion, bcrypt, usuarios, rbac, 12-factor]
aliases: [Seguridad y Autenticación, Gestión de Usuarios, Control de Acceso, Gestor de Seguridad]
---

# 🔐 Arquitectura de Seguridad, Autenticación y Gestión de Usuarios

Este documento especifica la arquitectura de seguridad empresarial de CARINA ubicada en [`src/utils/security/`](../../src/utils/security) y [`src/utils/security_manager.py`](../../src/utils/security_manager.py). Detalla la autenticación de usuarios, el hashing criptográfico de contraseñas, la defensa contra ataques de fuerza bruta, el aislamiento de esquemas de base de datos y el paradigma de configuración de aplicaciones de 12 factores.

⬅️ Volver al [Hub Maestro de Documentación](carina_moc.md) | 🗄️ Ver [Base de Datos y Esquemas](database_and_schemas.md) | 🛠️ Ver [Guías para Desarrolladores](developer_guides.md)

---

## 1. Visión General de la Arquitectura de Seguridad

CARINA gobierna infraestructura municipal crítica para la seguridad ciudadana. El acceso no autorizado o la elevación indebida de privilegios podrían comprometer los controladores semafóricos, ocasionando siniestros o colapsos viales. Para mitigar estas amenazas, el sistema aplica:

1. **Hashing Determinista de Contraseñas:** Hashing criptográfico `bcrypt` con sal (salt) y factores de coste computacional configurables.
2. **Defensa contra Fuerza Bruta (Motor de Bloqueo):** Limitación de tasa por IP y nombre de usuario con enfriamientos exponenciales.
3. **Control de Acceso Basado en Roles (RBAC):** Permisos diferenciados para perfiles de `OPERATOR`, `ENGINEER` y `ADMIN`.
4. **Aislamiento de Secretos según 12 Factores:** Desacoplamiento absoluto de credenciales sensibles del código fuente mediante archivos `.env`.

```text
 Cliente (UI Flet / CLI)
          │
          ▼
 ┌──────────────────┐      Valida Credenciales         ┌──────────────────┐
 │   AuthService    │ ───────────────────────────────> │  Generador Bcrypt│
 └────────┬─────────┘                                  └──────────────────┘
          │
          ├──> [Control de Fuerza Bruta] ──> LockdownManager (Máx 5 intentos, 15m bloqueo)
          │
          └──> [Consulta de Usuario]     ──> UserRepository (Esquema Aislado PostgreSQL / SQLite)
```

---

## 2. Desglose de Módulos Centrales

### 2.1 Hashing de Contraseñas y Criptografía (`hasher.py`)
Ubicado en [`src/utils/security/hasher.py`](../../src/utils/security/hasher.py):
- Implementa hashing seguro con `bcrypt`.
- Generación automática de sal ($2^{12}$ rondas por defecto).
- Comparación de cadenas resistente a ataques de temporización (*timing attacks*) para la verificación de contraseñas.

### 2.2 Protección contra Fuerza Bruta (`lockdown.py`)
Ubicado en [`src/utils/security/lockdown.py`](../../src/utils/security/lockdown.py):
- Rastrea intentos fallidos de inicio de sesión por nombre de usuario y origen del cliente.
- Dispara el bloqueo automático de la cuenta tras 5 intentos fallidos consecutivos.
- Aplica un período de bloqueo progresivo (predeterminado: 15 minutos).

### 2.3 Servicio y Repositorio de Usuarios (`user_service.py` y `user_repository.py`)
- [`src/utils/security/user_repository.py`](../../src/utils/security/user_repository.py): Persistencia SQL directa para registros de usuarios, sesiones activas y registros de auditoría.
- [`src/utils/security/migrator.py`](../../src/utils/security/migrator.py): Auto-inicializa y migra las tablas `users` y `audit_log` en esquemas de SQLite y PostgreSQL.

---

## 3. Configuración de 12 Factores y Variables de Entorno

Todas las credenciales sensibles deben definirse en el archivo [`.env`](../../.env.example) y nunca deben incluirse en el control de versiones de Git:

```bash
# --- Modo de Ejecución ---
# 'development': permite credenciales de demostración de GitHub (admin / Admin@Carina2026!).
# 'production': BLOQUEA ESTRICTAMENTE las credenciales por defecto; requiere secretos hasheados.
CARINA_ENV=production

# --- Inicialización de Superusuario ---
CARINA_SUPERUSER_USER=admin_traffic
CARINA_SUPERUSER_PASSWORD=YourStrongPasswordHere!
CARINA_SUPERUSER_HASH=$2b$12$e8x...

# --- Credenciales de Base de Datos ---
CARINA_DB_USER=carina_sec
CARINA_DB_PASSWORD=database_secure_password
CARINA_DB_HOST=127.0.0.1
CARINA_DB_PORT=5432
CARINA_DB_NAME=carina_data
CARINA_DB_SCHEMA=schema_carina

# --- Secreto para SNMP / Controladores Físicos ---
CARINA_SNMP_COMMUNITY=YourSnmpCommunityString

# --- Credenciales de Telemetría Externa para Monitor ---
CARINA_MQTT_HOST=127.0.0.1                     # Broker local o https://tunnel.ngrok-free.dev/api/telemetry
CARINA_MQTT_PORT=1883
CARINA_MQTT_USER=
CARINA_MQTT_PASSWORD=
```

---

## 4. Directrices de Fortalecimiento de Seguridad en Producción

1. **Rotar Cadenas de Comunidad SNMP:** Nunca utilice la cadena por defecto `"public"` en despliegues en campo.
2. **Restringir Enlaces de IPC y gRPC:** En entornos de producción, vincule las interfaces gRPC (`50051`) y WebSocket (`8080`) a VLANs privadas locales o a `127.0.0.1`, salvo que se encuentren detrás de un proxy inverso con terminación TLS.
3. **Mínimo Privilegio en Base de Datos:** Conceda al usuario de la aplicación en la base de datos únicamente permisos de `SELECT`, `INSERT` y `UPDATE` sobre los esquemas operacionales.
4. **Fail-Safe en Controladores Físicos:** Verifique que todos los controladores en campo posean modos internos activos de destello por horario del día en caso de interrupción de comunicaciones.

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Ingeniería de Movilidad Inteligente • CARINA CORE v1.2.0</i><br/>
  <small>© 2026 Noxfort Systems. Licenciado bajo AGPLv3.</small>
</div>
