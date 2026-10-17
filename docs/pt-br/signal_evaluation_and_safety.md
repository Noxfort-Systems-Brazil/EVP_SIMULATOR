---
tags: [safety, signal-evaluation, preemption, clearance, yellow, all-red]
aliases: [Avaliação Semafórica, Segurança Viária, Signal Evaluator]
---

# 🛡️ Avaliação Semafórica & Segurança de Trânsito

Este documento especifica o subsistema de auditoria de segurança e temporização semafórica implementado em `internal/domain/service/signal_evaluator.go`.

⬅️ [Central de Documentação](README.md) | 🏛️ [Arquitetura](architecture.md) | 🚨 [Walkthrough EVP](evp_walkthrough.md) | 🔌 [Emulação SNMP](utmc_and_snmp_emulation.md)

---

## 1. Protocolo de Transição Segura de Estágios

A concessão de prioridade de passagem **jamais** pode colocar os demais motoristas ou pedestres em risco de colisão. Por esse motivo, comutações abruptas de verde conflitante para verde de emergência são terminantemente vetadas.

Toda transição segue rigorosamente os estágios intermediários de segurança:

```text
Estágio Conflitante Ativo (Verde)
              │
              ▼  (Tempo de Amarelo: 3.5s - 4.5s)
Estágio Conflitante em Amarelo de Mudança
              │
              ▼  (Tempo de Vermelho Total / All-Red: 1.0s - 2.0s)
Vermelho Integral em Todas as Aproximações
              │
              ▼  (Acionamento do ACTION_HOLD)
Estágio Prioritário de Emergência (Verde Contínuo para a Viatura)
```

---

## 2. Avaliação de Proximidade e Cálculo do Ponto de Gatilho

O `SignalEvaluator` calcula em cada tick a projeção vetorial da viatura em relação à linha de retenção do semáforo:

```go
func (se *SignalEvaluator) EvaluateApproach(
    veh *entity.Vehicle,
    light *entity.TrafficLight,
    junction *entity.Junction,
) EvaluationResult {
    dist := veh.Position.DistanceToCartesian(junction.Coordinate)
    
    // Condição de Gatilho EVP: 60m <= dist <= 160m e sirene ligada
    if dist >= 60.0 && dist <= 160.0 && veh.SirenActive {
        return EvaluationResult{Action: ActionHoldTargetPhase, TargetPhase: 0}
    }
    
    // Condição de Checkout: dist <= 15m (cruzou o stop line)
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
  <i>Engenharia de Mobilidade Inteligente • EVP Simulator v1.0.0</i>
</div>
