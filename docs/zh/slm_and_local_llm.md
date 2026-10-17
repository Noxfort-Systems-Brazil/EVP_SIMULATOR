---
tags: [slm, llm, qwen, llama-cpp, gguf, natural-language, xai]
aliases: [端侧小语言模型, 本地大模型, 语义转换器, 离线人工智能]
---

# 🤖 端侧轻量大语言模型 (SLM) 与本地大模型集成

本文档详尽规范了 CARINA 位于 [`src/slm/`](../../src/slm) 的端侧轻量大语言模型（Small Language Model, SLM）子系统。深入阐述了 CARINA 如何利用本地离线开源大语言模型（如 **Qwen3 1.7B**、**Qwen2.5 1.5B/3B** 或 **Llama-3.2 3B**），在完全无需调用任何外部商业云端 API、零外部网络依赖的前提下，将底层神经网络敏感归因张量转化为符合法庭质证标准的出版级自然语言文本阐述。

⬅️ 返回 [核心技术主文档中心 (MOC)](carina_moc.md) | 🔍 查阅 [可解释人工智能与SAS](xai_and_sas.md) | 📄 查阅 [报告生成引擎](report_blocks_and_templates.md)

---

## 1. 架构核心理念：100% 离线私有化与司法数据主权

在市政公共交通治理与机要指挥调度领域，将城市道路路况遥测、交通硬件故障日志或信号控制决策数据发送至第三方公有商业云平台（如 OpenAI、Anthropic、Google Cloud 等），严重违反了公共数据主权治理法规、个人隐私保护法（LGPD / GDPR）以及国家关键基础设施安全准则。

CARINA 将所有报告自然语言文字论述的推理生成任务 **100% 约束在现场边缘物理服务器本地执行**：
- 基于经过高度汇编优化的 C++ 推理运行时（`llama-cpp-python`）。
- 全面支持 **4-bit 与 8-bit GGUF 模型量化技术**（`Q4_K_M`、`Q8_0`）。
- 仅需占用低至 **约 1.5 GB 至 3.5 GB 的显存 (VRAM)**（或在无独立显卡时通过 AVX-512 指令集直接在 CPU 内存中极速推理）。
- 采用低采样温度（$\tau \le 0.2$）实现高度确定性、严谨客观的工程语言输出。

```text
 Captum 积分梯度归因张量与微观遥测
                  │
                  ▼
      ┌───────────────────────┐
      │  semantic_transducer  │
      └───────────┬───────────┘
                  │
       ┌──────────┴──────────┐
       ▼                     ▼
┌──────────────┐      ┌──────────────┐
│prompt_builder│      │device_manager│ (动态计算显存占用 / 评估 GPU 卸载层数)
└──────┬───────┘      └──────┬───────┘
       │                     │
       └──────────┬──────────┘
                  ▼
      ┌───────────────────────┐
      │local_llama_transducer │ (通过 llama.cpp 执行 Qwen GGUF 本地推理)
      └───────────┬───────────┘
                  │
       ┌──────────┴──────────┐
       ▼                     ▼
┌───────────────┐     ┌───────────────┐
│output_sanitizer│    │revision_engine│ (交叉核验大模型文本是否与 SQL 遥测数值完全吻合)
└───────────────┘     └───────────────┘
```

---

## 2. `src/slm/` 核心功能模块

### 2.1 交通领域语义转换器 (`semantic_transducer.py`)
位于 [`src/slm/semantic_transducer.py`](../../src/slm/semantic_transducer.py)：
- 作为自然语言文本生成的主调度统一门面。
- 将微观数值元组 `(queue_length=18, speed=12.4 km/h, veto=True)` 转化为交通工程领域专用的结构化上下文提示词。
- 内置保底启发式引擎：若现场服务器显存不足或未挂载大模型权重文件，系统自动降级启用确定性规则专家库生成规范文本，绝不抛出异常。

### 2.2 本地 Llama 运行时转换器 (`local_llama_transducer.py`)
位于 [`src/slm/local_llama_transducer.py`](../../src/slm/local_llama_transducer.py)：
- 封装带线程安全执行互斥锁的 `llama_cpp.Llama` 运行时。
- 管理 GPU 层卸载策略（`n_gpu_layers = -1` 实现纯 CUDA 硬件加速全卸载）。
- 严格限制最大 Token 输出长度与生成温度，根除模型幻觉。

### 2.3 硬件设备与显存资源管理器 (`device_manager.py` 与 `resource_manager.py`)
- [`src/slm/device_manager.py`](../../src/slm/device_manager.py)：自动分析可用物理硬件环境（NVIDIA CUDA、AMD ROCm、Apple Metal 或纯 CPU 多线程）。
- [`src/slm/resource_manager.py`](../../src/slm/resource_manager.py)：自适应管理 Token 上下文窗口（$2048$ 至 $4096$ tokens），并在单次报告生成完毕后主动刷写释放 KV 缓存。

### 2.4 事实一致性核查与输出净化器 (`revision_engine.py` 与 `output_sanitizer.py`)
- [`src/slm/revision_engine.py`](../../src/slm/revision_engine.py)：提取生成段落中的数值实体，交叉校验其与底层 SQL 数据库实测记录的严格一致性。
- [`src/slm/output_sanitizer.py`](../../src/slm/output_sanitizer.py)：剔除 Markdown 残留代码标签、多余不可见字符与循环冗余废话。

---

## 3. 配置管理与模型金库挂载

预训练开源大模型权重文件存放于本地 `Model_Vault/` 目录：

```text
Model_Vault/
└── qwen3-1.7b-instruct-q4_k_m.gguf
```

在 `config/settings.ini` 中进行参数配置：
```ini
[XAI]
enable_xai = true
model_name = Model_Vault/qwen3-1.7b-instruct-q4_k_m.gguf
vram_allocation_gb = 3.5
temperature = 0.15
context_window = 3072
```

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>智能交通系统工程 • CARINA CORE v1.2.0</i><br/>
  <small>© 2026 Noxfort Systems. 基于 AGPLv3 授权。</small>
</div>
