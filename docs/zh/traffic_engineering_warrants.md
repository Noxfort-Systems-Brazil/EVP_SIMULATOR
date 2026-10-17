---
tags: [warrants, mutcd, fhwa, traffic-engineering, safety, analysis, sas]
aliases: [交通工程信号依据, 信号设置依据, MUTCD分析, 依据评估器]
---

# 🚦 交通工程信号设置依据与基础设施分析

本文档详尽规范了 CARINA 位于 [`src/analysis/`](../../src/analysis) 的交通工程法定设置依据分析子系统。深入阐述了基于美国 **FHWA MUTCD**（统一交通控制设施手册）以及巴西市政交通工程技术标准的信号控制法定设置依据的数学建模与自动化评估流程。

⬅️ 返回 [核心技术主文档中心 (MOC)](carina_moc.md) | 📈 查阅 [宏观基础图 (MFD)](mfd_and_analytics.md) | 🔍 查阅 [可解释人工智能与SAS](xai_and_sas.md)

---

## 1. 信号设置依据自动化评估架构

在城市道路规划中新建或调整交通信号灯之前，市政交通工程师必须依法证明目标交叉口满足法定的工程设置依据。**智能分析系统（Smart Analysis System, SAS）** 自动读取持久化于 PostgreSQL 的长期历史微观流体动力学数据，科学研判该交叉口是否确需设置信号灯或优化当前相位时序配时。

```text
 历史微观交通流数据 (PostgreSQL `synapse_fluid_dynamics`)
                             │
                             ▼
                ┌─────────────────────────┐
                │ InfrastructureAnalyzer  │
                └────────────┬────────────┘
                             │
              ┌──────────────┴──────────────┐
              ▼                             ▼
   ┌──────────────────────┐      ┌──────────────────────┐
   │   WarrantEvaluator   │      │     WarrantMath      │
   │   (法定依据评估策略) │      │ (流量与速度数学计算) │
   └──────────┬───────────┘      └──────────────────────┘
              │
              ├──> 依据 1: 八小时机动车交通量 (Eight-Hour Vehicular Volume)
              ├──> 依据 2: 四小时机动车交通量 (Four-Hour Vehicular Volume)
              ├──> 依据 3: 高峰小时延误与交通量 (Peak Hour Delay & Volume)
              ├──> 依据 7: 交通事故经验与临界率 (Crash Experience & Critical Rate)
              └──> 依据 8: 道路干线网络协调 (Roadway Network Coordination)
```

---

## 2. `src/analysis/` 模块实现的 MUTCD 核心依据

### 2.1 依据 1：八小时机动车交通量 (`warrant_strategies.py`)
评估在普通工作日中任意 8 个小时内，相交道路的交通量是否持续达到法定最小交通量门限：
- **条件 A（最小机动车交通量）：** 主干道与次干道进口道小时流量同时超过临界流量门限 $V_{major} \ge 500\text{ 辆/小时}$ 且 $V_{minor} \ge 150\text{ 辆/小时}$。
- **条件 B（连续车流中断）：** 主干道交通流量极为繁重，导致次干道汇入车辆遭受严重通行延误（$V_{major} \ge 750\text{ 辆/小时}$ 且 $V_{minor} \ge 75\text{ 辆/小时}$）。

### 2.2 依据 2：四小时机动车交通量
当主干道与次干道在任意连续 4 个小时内的交通流量均落在 [`src/analysis/warrant_math.py`](../../src/analysis/warrant_math.py) 绘制的 MUTCD 四小时标准曲线阈值上方时满足此项依据。

### 2.3 依据 3：高峰小时交通量与延误
评估全天单向流量最高峰的小时内，交通拥堵是否造成极端延误或严重安全隐患：
- 单车道次干道进口方向排队停车总延误达到或超过 4 车·小时。
- 该次干道进口方向高峰小时流量达到或超过 100 辆/小时。

### 2.4 依据 7：交通事故经验
评估安装交通信号灯是否能显著改善路口行车安全并降低事故发生率：
- 已经充分尝试了其他替代性交通管控措施，但仍无法有效遏制事故发生频次。
- 在最近连续 12 个月内，发生了 5 起及以上能够通过交通信号灯控制予以有效预防的交通事故（如直行与转弯车辆的垂直角度冲撞事故）。

### 2.5 依据 8：道路干线网络协调
当交叉口位于骨干路网节点时，通过设置信号控制来引导并集聚过境交通流，支撑干线动脉协同绿波控制。

---

## 3. 交通工程数学计算工具 (`warrant_math.py`)

位于 [`src/analysis/warrant_math.py`](../../src/analysis/warrant_math.py)：
- **第 85 百分位车速计算 (85th-Percentile Speed)：** 判定主干道设计或实测车速是否需要将各项依据门限交通量折减至 70%（适用于主干道实测车速超过 $70\text{ km/h}$ 或位于孤立乡村公路环境）。
- **临界流量比 (Critical Flow Ratio)：** 逐车道精密测算饱和流量与饱和度（$X = v/c$）。
- **滑动时间窗口累积统计：** 将微观离散时间序列传感器样本平滑聚合为高鲁棒性的小时交通量分布直方图。

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>智能交通系统工程 • CARINA CORE v1.2.0</i><br/>
  <small>© 2026 Noxfort Systems. 基于 AGPLv3 授权。</small>
</div>
