---
tags: [justificaciones, mutcd, fhwa, ingenieria-trafico, seguridad, analisis, sas]
aliases: [Criterios de Ingeniería de Tráfico, Justificaciones Semafóricas, Análisis MUTCD, Evaluador de Criterios]
---

# 🚦 Criterios de Ingeniería de Tráfico Semafórico y Análisis de Infraestructura

Este documento especifica el subsistema de análisis de justificaciones técnicas de CARINA ubicado en [`src/analysis/`](../../src/analysis). Detalla la formulación matemática y la evaluación automatizada de los criterios formales de semaforización basados en el **FHWA MUTCD** (Manual on Uniform Traffic Control Devices) y normativas de ingeniería de tránsito municipal.

⬅️ Volver al [Hub Maestro de Documentación](carina_moc.md) | 📈 Ver [MFD y Analítica](mfd_and_analytics.md) | 🔍 Ver [IA Explicable y SAS](xai_and_sas.md)

---

## 1. Evaluación Automatizada de Criterios Semafóricos

Antes de instalar o modificar semáforos, los ingenieros municipales están obligados legalmente a demostrar que una intersección cumple con criterios técnicos formales de justificación. El **Smart Analysis System (SAS)** evalúa los datos históricos acumulados en PostgreSQL para determinar si la semaforización está justificada o si requiere modificaciones en sus planes de tiempos.

```text
 Datos Históricos de Tráfico (PostgreSQL `synapse_fluid_dynamics`)
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
   │(Estrategias Criterio)│      │(Matemática Vol y Vel)│
   └──────────┬───────────┘      └──────────────────────┘
              │
              ├──> Criterio 1: Volumen Vehicular en Ocho Horas
              ├──> Criterio 2: Volumen Vehicular en Cuatro Horas
              ├──> Criterio 3: Demora y Volumen en Hora Pico
              ├──> Criterio 7: Experiencia de Siniestralidad
              └──> Criterio 8: Coordinación de la Red Vial
```

---

## 2. Criterios MUTCD Implementados en `src/analysis/`

### 2.1 Criterio 1: Volumen Vehicular en Ocho Horas (`warrant_strategies.py`)
Evalúa si el volumen de tráfico en las calles intersecantes cumple los umbrales mínimos durante cualquiera de las 8 horas de un día promedio:
- **Condición A (Volumen Vehicular Mínimo):** Los volúmenes de aproximación en la calle principal y secundaria superan los volúmenes horarios críticos $V_{principal} \ge 500\text{ veh/h}$ y $V_{secundaria} \ge 150\text{ veh/h}$.
- **Condición B (Interrupción del Tráfico Continuo):** El volumen en la vía principal es tan intenso que el tráfico en la vía secundaria sufre demoras excesivas ($V_{principal} \ge 750\text{ veh/h}$ y $V_{secundaria} \ge 75\text{ veh/h}$).

### 2.2 Criterio 2: Volumen Vehicular en Cuatro Horas
Aplica cuando el volumen en las vías principal y secundaria satisface las curvas de 4 horas del MUTCD trazadas en [`src/analysis/warrant_math.py`](../../src/analysis/warrant_math.py).

### 2.3 Criterio 3: Demora y Volumen en Hora Pico
Evalúa si las condiciones de tráfico durante la hora de mayor demanda del día generan riesgos o demoras extremas:
- La demora total por detención sufrida por el tráfico en una aproximación secundaria es igual o superior a 4 vehículos-hora para una aproximación de un solo carril.
- El volumen en dicha aproximación es igual o superior a 100 vehículos por hora.

### 2.4 Criterio 7: Experiencia de Siniestralidad (Accidentes)
Evalúa si la instalación semafórica reducirá la frecuencia y gravedad de siniestros:
- La aplicación adecuada de medidas alternativas no ha logrado mitigar la frecuencia de colisiones.
- Han ocurrido cinco o más siniestros reportados durante un período de 12 meses susceptibles de prevención semafórica (ej., colisiones en ángulo recto).

### 2.5 Criterio 8: Red Vial Coordinada
Justifica la semaforización para fomentar la concentración y organización del flujo vehicular en una red arterial coordinada.

---

## 3. Utilidades Matemáticas (`warrant_math.py`)

Ubicadas en [`src/analysis/warrant_math.py`](../../src/analysis/warrant_math.py):
- **Cálculo de Velocidad Percentil 85:** Evalúa si la velocidad operativa permite reducir los volúmenes umbral en un 70% (aplicable si la velocidad supera $70\text{ km/h}$ o en comunidades rurales aisladas).
- **Ratio de Flujo Crítico:** Calcula flujos de saturación carril por carril y el grado de saturación ($X = v/c$).
- **Acumulación por Ventana Móvil:** Agrupa muestras de series temporales en contenedores horarios robustos.

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Ingeniería de Movilidad Inteligente • CARINA CORE v1.2.0</i><br/>
  <small>© 2026 Noxfort Systems. Licenciado bajo AGPLv3.</small>
</div>
