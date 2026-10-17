---
tags: [research, ai, math, neural]
aliases: [神经科学研究, 数学理论推导, PPO-TCN, GOMES]
---

# 🧠 神经网络前沿研究与数学理论推导

本文档详尽规范了支撑 CARINA **基于图的大规模多智能体专家运行系统 (GOMES)** 以及 **DA SILVA** 智能体渐进成熟度课程学习机制的数学理论框架、神经网络架构与优化推导。

⬅️ 返回 [核心技术主文档中心 (MOC)](carina_moc.md) | 🏛️ 查阅 [多进程架构](architecture.md) | 🛡️ 查阅 [安全机制与看门狗](safety_and_watchdog.md) | 🔍 查阅 [可解释人工智能与SAS](xai_and_sas.md)

---

## 1. PPO-TCN：基于时序卷积网络 (TCN) 替代传统循环网络 (LSTM)

传统循环神经网络（如 LSTM/GRU）在长回合的时间反向传播（BPTT）过程中极易遭遇梯度消失与历史时序上下文遗忘问题。CARINA 在 PPO 战术智能体中引入了 **时序卷积网络（Temporal Convolutional Network, TCN）** 作为主干特征提取器。

```text
输入历史张量: X ∈ ℝ^(B × C × T)
  B: 训练批量大小 (Batch size)
  C: 每个时间步的特征维度 (时间占有率、平均车速、排队长度、信号灯相位)
  T: 时序历史上下文滑动窗口 (例如: 10Hz 采样率下 8 秒时长 = 80 帧)
```

### 1.1 空洞因果卷积 (Dilated Causal Convolutions)

为了严格杜绝未来信息的时序泄漏（即时间步 $t$ 处的特征推断严格仅依赖于 $\le t$ 的历史观测），TCN 采用了基于一维空洞因果卷积的结构：

$$y(t) = (x \star_d f)(t) = \sum_{i=0}^{k-1} f(i) \cdot x(t - d \cdot i)$$

其中各项定义如下：
- $k$：一维卷积核尺寸 ($k = 3$)
- $d$：空洞扩张因子 ($d = 2^l$，在网络层索引 $l \in \{0, 1, 2, 3\}$ 处呈指数级扩张)
- 感受野有效跨度：$RF = 1 + \sum_{l=0}^{L-1} (k - 1) \cdot d_l$

TCN 最终输出高密度的时序抽象特征张量 $H_{tcn}$，并将其并行输入至 PPO 的 Actor（策略网络）与 Critic（价值网络）输出头。

### 1.2 PPO 截断代理目标函数与优势估计

PPO 策略网络利用融合了广义优势估计（GAE）的截断代理目标函数进行梯度反向传播优化：

$$L^{CLIP}(\theta) = \hat{\mathbb{E}}_t \left[ \min \left( r_t(\theta)\hat{A}_t, \text{clip}(r_t(\theta), 1-\epsilon, 1+\epsilon)\hat{A}_t \right) \right]$$

其中：
- $r_t(\theta) = \frac{\pi_\theta(a_t | s_t)}{\pi_{\theta_{old}}(a_t | s_t)}$ 代表新旧策略的动作概率比率。
- $\hat{A}_t = \sum_{l=0}^{\infty} (\gamma \lambda)^l \delta_{t+l}^V$ 代表 GAE 广义优势估计量。
- $\epsilon = 0.2$ 代表策略概率更新的截断保护边界。

---

## 2. 预测自编码器 (PAE) 与潜在特征流形 $Z$

**预测自编码器 (Predictive Autoencoder, PAE)** 旨在通过时空连续历史趋势，前瞻性地预测未来短周期内的交通排队溢流动态。

```text
        ┌──────────────┐          ┌───────────────────────┐          ┌──────────────┐
X_hist  │  PAE 编码器  │ ───────> │    潜在瓶颈流形 Z     │ ───────> │  PAE 解码器  │ ───> X_pred
(t-8s..t)└──────────────┘          │  Z ∈ ℝ^16             │          └──────────────┘      (t+1s..t+30s)
                                   └───────────────────────┘
```

### 2.1 PAE 联合正则化损失函数

PAE 在离线预训练及在线微调中采用联合均方误差（MSE）与 Kullback-Leibler（KL）散度正则化约束：

$$\mathcal{L}_{PAE} = \underbrace{\| X_{pred} - X_{future} \|_2^2}_{\text{未来时空预测重构损失}} + \beta \cdot D_{KL}\Big( q_\phi(Z|X) \,\parallel\, \mathcal{N}(0, I) \Big)$$

生成的 16 维潜在物理特征向量 $Z \in \mathbb{R}^{16}$ 紧凑刻画了道路排队车流的动量演化，并与基础环境观测拼接注入 PPO 输入层。

---

## 3. Dueling DQN-TCN 守卫安全智能体

**Guardian 守卫智能体** 通过在 TCN 时序特征上运行 Dueling Deep Q-Network，持续评估每个进口道车流的**溢流回堵风险指数**：

$$Q(s, a; \theta, \alpha, \beta) = V(s; \theta, \beta) + \left( A(s, a; \theta, \alpha) - \frac{1}{|\mathcal{A}|} \sum_{a' \in \mathcal{A}} A(s, a'; \theta, \alpha) \right)$$

其中：
- $V(s)$：状态 $s$ 的标量基准状态价值。
- $A(s, a)$：采取特定信号相位动作 $a$ 的相对优势值。

若 $\max_a Q(s, a_{spillback}) > \tau_{risk}$（其中风险阈值 $\tau_{risk} = 0.8$），Guardian 智能体将强行覆盖战术智能体的提议动作，直接注入紧急清空相位。

---

## 4. 宏观战略协调层：图注意力网络 (ST-GATv2 Lite)

为了在主干道相邻路口之间自主协调形成大范围“动态绿波带”，CARINA 将整座城市的道路拓扑建模为有向图 $\mathcal{G} = (\mathcal{V}, \mathcal{E})$。

### 4.1 GATv2 动态注意力系数计算

针对物理连通的相邻交叉口节点 $i, j \in \mathcal{V}$，GATv2 在线计算动态空间注意力权重 $\alpha_{ij}$：

$$\alpha_{ij} = \frac{\exp\left( \mathbf{a}^T \text{LeakyReLU}\left( \mathbf{W} [h_i \,\|\, h_j] \right) \right)}{\sum_{k \in \mathcal{N}_i} \exp\left( \mathbf{a}^T \text{LeakyReLU}\left( \mathbf{W} [h_i \,\|\, h_k] \right) \right)}$$

其中：
- $h_i, h_j$：路口节点 $i$ 与其邻居节点 $j$ 的高维嵌入特征。
- $\mathbf{W}$：全局共享的线性变换参数矩阵。
- $\mathbf{a}$：可学习的空间注意力权重向量。

---

## 5. DA SILVA 智能体演化与成熟度课程体系

**自主性验证动态智能体安全集成学习课程体系（DA SILVA）** 引入**策略香农熵（Policy Shannon Entropy）** 作为评估智能体策略成熟度的核心数学测度：

$$\mathcal{H}(\pi_\theta(\cdot | s_t)) = -\sum_{a \in \mathcal{A}} \pi_\theta(a | s_t) \log \pi_\theta(a | s_t)$$

### 5.1 晋升评估准则与动态熵衰减边界

为彻底杜绝盲目随机探索策略过早获得硬件控制权，系统设定了动作空间理论最大熵上限 $E_{max} = \ln(|\mathcal{A}|)$（例如包含 3 种动作时 $E_{max} = \ln(3) \approx 1.0986$），或归一化熵 $\mathcal{H}_{norm} = \mathcal{H} / \ln(|\mathcal{A}|) \in [0, 1]$。`DynamicEntropyCalculator` 依据训练经历的迭代回合数（$t$ 回合）动态衰减准入熵阈值：

$$E(t) = E_{ideal} + (E_{max} - E_{ideal}) \cdot e^{-k \cdot t}$$

其中，理想目标熵 $E_{ideal} = 0.10$。幼年 $\to$ 青年阶段衰减系数取 $k = 0.05$，青年 $\to$ 成年阶段衰减系数取 $k = 0.15$。渐进的目标晋升区间如下：

1. **幼年期 CHILD（影子观察模式）：** $\mathcal{H} > E_{child}(t)$（渐进基准 $\mathcal{H} > 0.40$）。动作仅作影子推演（仅采集遥测），底层由传统固定配时控制。由于对所有 $t \ge 1$ 均满足 $E(t) < \ln(3)$，完全随机策略（$\mathcal{H} = \ln 3$）绝对无法跨级晋升。
2. **青年期 TEEN（限制性自主运行）：** $E_{adult}(t) < \mathcal{H} \le E_{child}(t)$（渐进区间 $0.15 < \mathcal{H} \le 0.40$）。智能体仅获准在夜间极低风险时段（凌晨 01:00 至 04:00）尝试驱动物理硬件。
3. **成年期 ADULT（完全自主接管）：** $\mathcal{H} \le E_{adult}(t)$（渐进基准 $\mathcal{H} \le 0.15$，在 $k=0.15$ 下经历 $t \ge 20$ 回合达成）且强化学习平均奖励 $\mu_{reward} > \mu_{baseline} + \delta^*$。正式授予 7×24 小时全天候自主控制权限。

### 5.2 相位内群体进化算法 (Intra-Phase PBT)
为了在多智能体群体进化过程中维护发育阶段的一致性，[`PopulationManager`](../../src/core/population_manager.py) 强制实施严格的生命周期层级隔离：
- **同梯队竞争变异：** 智能体仅在同层级伙伴之间开展权重克隆与超参数变异（$\text{TEEN} \leftrightarrow \text{TEEN}$，$\text{ADULT} \leftrightarrow \text{ADULT}$）。
- **幼年期免疫防护：** 处于 `CHILD` 阶段的智能体被完全屏蔽在变异扰动之外，专注于在纯净状态下被动观测路网物理流动。
- **异构拓扑尺寸安全：** 跨不同车道拓扑进行权重迁移时，内建严格的张量维度动态校验，杜绝形状不匹配崩溃。

### 5.3 动态毕业性能增益裕度计算 ($\delta^*$)
CARINA 摒弃了静态武断的考核标准，智能体从 `TEEN` 晋升至 `ADULT` 所需的超越基准传统控制器的性能增益裕度，完全由历史实测道路流体动力学动态计算求得：

$$\delta^* = K \times \left(1 - \frac{\bar{v}_{\text{child}}}{v_{\text{limit}}}\right) \times (1 - \text{occupancy}_{\text{child}}) \times 100\%$$

其中各项物理意义为：
- $\left(1 - \frac{\bar{v}_{\text{child}}}{v_{\text{limit}}}\right)$ 量化了幼年阶段观测到的传统固定配时方案的运行低效度。
- $(1 - \text{occupancy}_{\text{child}})$ 代表道路尚存的物理时空容量裕度。
- $K = 0.25$ 将提升预期与全球主流自适应系统（如 SCATS / SCOOT 的经典提升幅度 $10\%\text{--}15\%$）相标定。
- $\delta^*$ 被严格钳位在 $[3.0\%, 25.0\%]$ 之间，避免在过饱和路网中设定不可企及的目标，或在空旷道路上过早晋升。

### 5.4 物理奖励函数平衡机制 (HCM 道路通行能力物理学)
CARINA 对多目标强化学习奖励函数进行量纲物理对齐，消除排队延误惩罚与通行放行奖励之间的数量级偏差：

$$R(t) = w_{\text{wait}} \times T_{\text{wait}}(t) + w_{\text{flow}} \times F_{\text{discharged}}(t)$$

依据美国公路通行能力手册（HCM）原理，每顺利放行一辆排队车辆，即可为系统挽回相当于整条路段特征行程时间的延误：

$$\tau = \frac{\bar{L}_{\text{links}}}{\bar{v}_{\text{limit}}}$$

[`RewardWeightsCalculator`](../../src/engine/reward_weights_calculator.py) 计算出的物理平衡参数如下：
- $w_{\text{wait}} = -1.0$（排队车辆每秒累积的归一化延误惩罚）。
- $w_{\text{flow}} = \text{clamp}(2.0, 25.0, \text{round}(\tau \times \alpha, 2))$，确保放行车辆的激励权重与累积排队延误在物理尺度上完全匹配。

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>智能交通系统工程 • CARINA CORE v1.2.0</i><br/>
  <small>© 2026 Noxfort Systems. 基于 AGPLv3 授权。</small>
</div>
