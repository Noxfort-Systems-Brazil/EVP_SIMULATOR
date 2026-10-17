---
tags: [safety, signal-evaluation, preemption, clearance, yellow, all-red]
aliases: [Signal Evaluation, Traffic Safety, Signal Evaluator]
---

# 🛡️ Оценка светофорных фаз и аудит безопасности движения

This document specifies the real-time signal phase evaluation and safety auditing engine in `internal/domain/service/signal_evaluator.go`.

⬅️ [Центр документации](README.md) | 🏛️ [Architecture](architecture.md) | 🚨 [EVP Walkthrough](evp_walkthrough.md) | 🔌 [SNMP Emulation](utmc_and_snmp_emulation.md)

---

## 1. Safe Stage Transition Protocol

Granting emergency priority must never compromise vehicular or pedestrian safety:

```text
Active Conflicting Stage (Green)
              │
              ▼  (Yellow Clearance: 3.5s - 4.5s)
Conflicting Stage in Yellow Change Interval
              │
              ▼  (All-Red Clearance: 1.0s - 2.0s)
All-Red Clearance Interval Across All Approaches
              │
              ▼  (Command ACTION_HOLD)
Emergency Target Stage Locked in Green Hold
```

---

## 2. Dual-Threshold Spatial Logic

```go
func (se *SignalEvaluator) EvaluateApproach(
    veh *entity.Vehicle,
    light *entity.TrafficLight,
    junction *entity.Junction,
) EvaluationResult {
    dist := veh.Position.DistanceToCartesian(junction.Coordinate)
    
    // EVP Preemption Trigger Zone: 60m <= dist <= 160m and siren engaged
    if dist >= 60.0 && dist <= 160.0 && veh.SirenActive {
        return EvaluationResult{Action: ActionHoldTargetPhase, TargetPhase: 0}
    }
    
    // Checkout Zone: dist <= 15m (crossed stop line)
    if dist <= 15.0 && light.IsPreempted && light.PreemptingVehicleID == veh.ID {
        return EvaluationResult{Action: ActionReleasePreemption}
    }
    
    return EvaluationResult{Action: ActionNone}
}
```

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Инженерия интеллектуальной мобильности • EVP Simulator v1.0.0</i>
  <small>SPDX-License-Identifier: <a href="../../LICENSE">Apache-2.0</a> • &copy; 2026 Noxfort Systems</small>
</div>
