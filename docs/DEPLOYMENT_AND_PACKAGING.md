---
tags: [deployment, packaging, docker, debian, systemd]
aliases: [Deployment Guide, Packaging, Docker, Systemd]
---

# 🚀 Deployment & Service Configuration

This document details production deployment options for CARINA, including Systemd service setup and environment configuration.

⬅️ Back to [Main Documentation Hub](CARINA_MOC.md) | 🛠️ See [Developer Guides](DEVELOPER_GUIDES.md) | 🗄️ See [Database & Schemas](DATABASE_AND_SCHEMAS.md) | 🧪 See [Testing & Validation](TESTING.md) | 🚑 See [EVP Preemption](EVP_AND_PREEMPTION.md)

---

## 1. Network Ports & Communication Reference

CARINA coordinates local multiprocessing pipelines, hardware controllers, and cloud videowalls using dedicated ports:

| Port | Protocol | Subsystem / Process | Description |
| :--- | :--- | :--- | :--- |
| **`50051`** | TCP (gRPC) | CentralController $\leftrightarrow$ Microservices | Internal RPC bus for telemetry frames and neural action requests. |
| **`8080`** | TCP (WS) | Smart Dashboard Service (SDS) | Telemetry WebSocket server streaming live queue data to web clients. |
| **`2201`** | TCP (HTTP/SSE) | Go Web Gateway (`src_go/web_gateway`) | Primary edge ingress: `POST /api/v1/evp/telemetry` (HTTP 204), `GET /cco/stream` (SSE), and CCO dashboard. |
| **`8085`** | TCP (HTTP) | Standalone EVP Receiver (`src/evp`) | Fallback emergency vehicle GPS telemetry listener when Go gateway is inactive. |
| **`42123`** | TCP (Local) | SingleInstanceLock (`single_instance.py`) | OS mutex preventing duplicate application instances; sends `RESTORE` signal. |
| **`161`** | UDP (SNMP) | NTCIP 1202 / UTMC Field Controllers | Polling and phase hold control (`ascPhaseHold`) sent to traffic signal heads. |
| **`162`** | UDP (Trap) | Hardware Trap Listener (`traplistener`) | Ingestion of asynchronous hardware fault traps and detector actuations. |

---

## 2. Building Core Subsystems

Before deploying CARINA in production, native binaries and web assets must be built:

### 2.1 Compiling Go Hardware & Web Gateway
The compiled Go daemon handles physical controller SNMP polling, microsecond EVP preemption holds, and CCO videowall streaming:
```bash
cd src_go
go build -ldflags="-s -w" -o ../bin/carina-go ./cmd/gateway
chmod +x ../bin/carina-go
```

### 2.2 Building React CCO Web Dashboard
The Central Control Operator (CCO) web interface resides in [`src_tsx/`](../src_tsx):
```bash
cd src_tsx
npm install
npm run build
```
Vite outputs production bundles into `src_tsx/dist/`, which are automatically served by the Go Web Gateway at `http://localhost:2201/cco/`.

---

## 3. Production Docker Containerization (`Dockerfile`)

CARINA includes an optimized multi-stage build `Dockerfile` for GPU-accelerated container execution:

### 3.1 Build Docker Image
```bash
docker build -t carina-core:latest .
```

### 3.2 Run Container with GPU Acceleration & Network Ports
```bash
docker run -d \
  --name carina_app \
  --gpus all \
  -p 50051:50051 \
  -p 8080:8080 \
  -p 2201:2201 \
  -p 8085:8085 \
  -v /var/log/carina:/app/logs \
  -v /etc/carina:/app/config \
  carina-core:latest
```

---

## 4. Linux Systemd Daemon Configuration

To run CARINA as a system daemon on Ubuntu/Debian Linux:

Create `/etc/systemd/system/carina.service`:

```ini
[Unit]
Description=CARINA AI Real-Time Traffic Orchestrator
After=network.target postgresql.service

[Service]
Type=simple
User=carina
WorkingDirectory=/opt/carina
ExecStart=/usr/bin/python3 /opt/carina/carina.py
Restart=always
RestartSec=5
Environment=PYTHONUNBUFFERED=1

[Install]
WantedBy=multi-user.target
```

Enable and start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable carina
sudo systemctl start carina
```

---

## 5. Zero-Trust Mesh Deployment via Tailscale Funnel

For distributed metropolitan networks without public static IPs:
1. Generate a Tailscale Auth Key with `funnel` capabilities from the Tailscale Admin Console.
2. Configure `.env`:
   ```bash
   CARINA_GATEWAY_ENABLED=true
   CARINA_GATEWAY_PORT=2201
   CARINA_TAILSCALE_AUTH_KEY=tskey-auth-kXXXXX-XXXXXXXX
   ```
3. When the Go Gateway starts, [`src_go/web_gateway/tailscale.go`](../src_go/web_gateway/tailscale.go) automatically registers the node and spins up an HTTPS Funnel at `https://<node-name>.<tailnet>.ts.net:2201`, exposing the EVP API and CCO videowall over public TLS with zero firewall changes.

---

## 6. Packaging & Distribution

Legacy PyInstaller and `.deb` packaging via `build_installer.sh` have been removed. A modern, containerized distribution and installation pipeline is planned for upcoming releases.

---

<div align="center">
  <img src="assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Intelligent Mobility Engineering • CARINA CORE v1.2.0</i><br/>
  <small>© 2026 Noxfort Systems. Licensed under AGPLv3.</small>
</div>
