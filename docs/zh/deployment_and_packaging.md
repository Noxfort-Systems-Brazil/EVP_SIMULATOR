---
tags: [deployment, packaging, docker, debian, systemd]
aliases: [部署指南, 系统打包, Docker容器化, Systemd服务]
---

# 🚀 系统部署与生产环境服务配置

本文档详尽规范了 CARINA 在生产环境下的部署规范，涵盖网络端口规划、编译构建流程、Docker 容器化、Linux Systemd 守护进程配置以及基于 Tailscale 的安全远程穿透。

⬅️ 返回 [核心技术主文档中心 (MOC)](carina_moc.md) | 🛠️ 查阅 [开发者指南](developer_guides.md) | 🗄️ 查阅 [数据库与模式](database_and_schemas.md) | 🧪 查阅 [测试规范与质量保证](testing.md) | 🚑 查阅 [EVP 优先通行](evp_and_preemption.md)

---

## 1. 网络端口规划与通信参考速查表

CARINA 统筹管理本地多进程通信、现场硬件控制器以及云端大屏展示所使用的专用端口：

| 监听端口 | 传输层协议 | 对应子系统 / 进程 | 功能语义说明 |
| :--- | :--- | :--- | :--- |
| **`50051`** | TCP (gRPC) | CentralController $\leftrightarrow$ 微服务集群 | 内部高频 RPC 总线，用于时序遥测帧与神经控制动作交互。 |
| **`8080`** | TCP (WS) | 智能仪表盘服务 (SDS) | 遥测 WebSocket 服务器，向 Web 客户端广播实时排队数据。 |
| **`2201`** | TCP (HTTP/SSE) | Go Web 网关 (`src_go/web_gateway`) | 边缘综合接入服务：`POST /api/v1/evp/telemetry` (HTTP 204)、`GET /cco/stream` (SSE) 及 CCO 大屏。 |
| **`8085`** | TCP (HTTP) | 独立 EVP 接收器 (`src/evp`) | Go 网关未启用时，用于应急特种车辆 GPS 遥测接入的备用监听器。 |
| **`42123`** | TCP (Local) | 单实例互斥锁 (`single_instance.py`) | 本地操作系统互斥锁，防止应用多开；向下发 `RESTORE` 唤醒信号。 |
| **`161`** | UDP (SNMP) | NTCIP 1202 / UTMC 现场交通机 | 向路口交通信号机发送轮询与相位锁定指令 (`ascPhaseHold`)。 |
| **`162`** | UDP (Trap) | 硬件陷阱监听服务器 (`traplistener`) | 异步摄入现场交通机硬件故障告警陷阱与检测器触发报文。 |

---

## 2. 核心子系统编译构建流程

在生产环境中正式部署 CARINA 之前，需完成原生二进制网关及 Web 前端资产的编译：

### 2.1 编译 Go 硬件与 Web 通信网关
编译生成的 Go 原生二进制守护进程负责现场交通机 SNMP 轮询、微秒级 EVP 相位锁定以及 CCO 大屏 SSE 流分发：
```bash
cd src_go
go build -ldflags="-s -w" -o ../bin/carina-go ./cmd/gateway
chmod +x ../bin/carina-go
```

### 2.2 构建 React CCO Web 运营大屏前端
中央控制调度中心 (CCO) 现代化大屏前端位于 [`src_tsx/`](../../src_tsx)：
```bash
cd src_tsx
npm install
npm run build
```
Vite 构建工具将生产静态包输出至 `src_tsx/dist/`，该静态资源由 Go Web 网关在 `http://localhost:2201/cco/` 路由下自动托管交付。

---

## 3. 生产环境 Docker 容器化构建 (`Dockerfile`)

CARINA 提供了经过多阶段构建优化、支持 GPU 硬件加速的标准化 `Dockerfile`：

### 3.1 构建 Docker 镜像
```bash
docker build -t carina-core:latest .
```

### 3.2 启用 GPU 硬件加速与端口映射运行容器
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

## 4. Linux Systemd 系统守护进程配置

在 Ubuntu / Debian Linux 生产服务器上以系统系统守护进程形式运行 CARINA：

创建服务单元文件 `/etc/systemd/system/carina.service`：

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

加载配置并启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable carina
sudo systemctl start carina
```

---

## 5. 基于 Tailscale Funnel 的零信任安全网格部署

针对无静态公网 IP 的分布式都市交通边缘机房：
1. 在 Tailscale 控制台生成具备 `funnel` 授权的 Auth Key。
2. 配置 `.env` 环境变量：
   ```bash
   CARINA_GATEWAY_ENABLED=true
   CARINA_GATEWAY_PORT=2201
   CARINA_TAILSCALE_AUTH_KEY=tskey-auth-kXXXXX-XXXXXXXX
   ```
3. 当 Go 网关启动时，[`src_go/web_gateway/tailscale.go`](../../src_go/web_gateway/tailscale.go) 将自动完成节点注册并建立基于公网 TLS 的安全 HTTPS Funnel（例如 `https://<node-name>.<tailnet>.ts.net:2201`），在零路由器改动的前提下安全对外暴露 EVP 接入 API 与 CCO 指挥大屏。

---

## 6. 打包与分发说明

旧版基于 PyInstaller 与通过 `build_installer.sh` 构建 `.deb` 的打包脚本已被废弃移除。系统未来版本将推出基于现代化容器编排与跨平台打包的标准化安装交付体系。

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>智能交通系统工程 • CARINA CORE v1.2.0</i><br/>
  <small>© 2026 Noxfort Systems. 基于 AGPLv3 授权。</small>
</div>
