---
tags: [slm, llm, qwen, llama-cpp, gguf, lenguaje-natural, xai]
aliases: [Modelos de Lenguaje Reducidos, LLM Local, Transductor Semántico, IA Offline]
---

# 🤖 Modelos de Lenguaje Reducidos (SLM) e Integración de LLM Local

Este documento especifica el subsistema de Modelos de Lenguaje Reducidos (SLM) de CARINA ubicado en [`src/slm/`](../../src/slm). Detalla cómo CARINA genera justificaciones textuales periciales listas para publicación a partir de tensores de atribución neuronal mediante modelos de lenguaje locales y offline (ej., **Qwen3 1.7B**, **Qwen2.5 1.5B/3B** o **Llama-3.2 3B**) sin realizar llamadas a APIs externas ni depender de la nube.

⬅️ Volver al [Hub Maestro de Documentación](carina_moc.md) | 🔍 Ver [IA Explicable y SAS](xai_and_sas.md) | 📄 Ver [Bloques de Informes y Word](report_blocks_and_templates.md)

---

## 1. Filosofía Arquitectónica: Privacidad Forense 100% Offline

En la administración pública municipal, enviar telemetría de tráfico, registros de incidentes o datos de control de intersecciones a APIs comerciales en la nube (OpenAI, Anthropic, Google Cloud) vulnera la gobernanza de datos públicos, normativas de privacidad (LGPD/GDPR) y la soberanía municipal.

CARINA ejecuta toda la síntesis de informes en lenguaje natural **100% localmente en sitio (on-premise)**:
- Se ejecuta mediante enlaces C++ optimizados (`llama-cpp-python`).
- Soporta **cuantización GGUF en 4 y 8 bits** (`Q4_K_M`, `Q8_0`).
- Consume tan solo **~1.5 GB a 3.5 GB de VRAM** (o memoria RAM del sistema con aceleración vectorial AVX-512).
- Salida completamente determinista mediante muestreo a baja temperatura ($\tau \le 0.2$).

```text
 Gradientes Integrados Captum y Telemetría
                   │
                   ▼
       ┌───────────────────────┐
       │  semantic_transducer  │
       └───────────┬───────────┘
                   │
        ┌──────────┴──────────┐
        ▼                     ▼
 ┌──────────────┐      ┌──────────────┐
 │prompt_builder│      │device_manager│ (Calcula VRAM y capas de descarga a GPU)
 └──────┬───────┘      └──────┬───────┘
        │                     │
        └──────────┬──────────┘
                   ▼
       ┌───────────────────────┐
       │local_llama_transducer │ (Ejecuta el modelo Qwen GGUF vía llama.cpp)
       └───────────┬───────────┘
                   │
        ┌──────────┴──────────┐
        ▼                     ▼
 ┌───────────────┐     ┌───────────────┐
 │output_sanitizer│    │revision_engine│ (Cruza y valida texto frente a telemetría)
 └───────────────┘     └───────────────┘
```

---

## 2. Módulos Centrales en `src/slm/`

### 2.1 Transductor Semántico (`semantic_transducer.py`)
Ubicado en [`src/slm/semantic_transducer.py`](../../src/slm/semantic_transducer.py):
- Actúa como la interfaz pública principal para la generación de explicaciones textuales.
- Transforma tuplas numéricas `(queue_length=18, speed=12.4 km/h, veto=True)` en prompts especializados de ingeniería de tráfico.
- Emplea un motor heurístico de contingencia: si el hardware carece de GPU/RAM o faltan los pesos del LLM, genera texto técnico determinista basado en reglas sin interrumpir el servicio.

### 2.2 Transductor Local Llama (`local_llama_transducer.py`)
Ubicado en [`src/slm/local_llama_transducer.py`](../../src/slm/local_llama_transducer.py):
- Encapsula `llama_cpp.Llama` con bloqueos de ejecución thread-safe.
- Administra la descarga de capas a la GPU (`n_gpu_layers = -1` para descarga total por CUDA).
- Restringe la longitud de generación y la temperatura para impedir alucinaciones textuales.

### 2.3 Gestión de Dispositivos y Recursos (`device_manager.py` y `resource_manager.py`)
- [`src/slm/device_manager.py`](../../src/slm/device_manager.py): Perfila el hardware disponible (NVIDIA CUDA, ROCm, Apple Metal o hilos de CPU).
- [`src/slm/resource_manager.py`](../../src/slm/resource_manager.py): Administra automáticamente las ventanas de contexto ($2048$ a $4096$ tokens) y purga la caché KV entre episodios de generación de informes.

### 2.4 Motor de Revisión Factual (`revision_engine.py` y `output_sanitizer.py`)
- [`src/slm/revision_engine.py`](../../src/slm/revision_engine.py): Extrae entidades numéricas de la prosa generada y verifica que coincidan estrictamente con los valores registrados en la base de datos SQL.
- [`src/slm/output_sanitizer.py`](../../src/slm/output_sanitizer.py): Elimina tokens de formato markdown, espacios en blanco superfluos y artefactos de bucles repetitivos.

---

## 3. Configuración y Preparación de la Bóveda de Modelos

Los pesos preentrenados se almacenan en el directorio local `Model_Vault/`:

```text
Model_Vault/
└── qwen3-1.7b-instruct-q4_k_m.gguf
```

Configuración en `config/settings.ini`:
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
  <i>Ingeniería de Movilidad Inteligente • CARINA CORE v1.2.0</i><br/>
  <small>© 2026 Noxfort Systems. Licenciado bajo AGPLv3.</small>
</div>
