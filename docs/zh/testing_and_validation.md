---
tags: [testing, validation, unit-tests, benchmarks, qa, go-test]
aliases: [Testing & QA, Test Suite, Validation Guidelines]
---

# 🧪 自动化测试规范与质量保证

This document outlines testing practices, unit test suites, and integration tests for the **EVP Simulator**.

⬅️ [技术文档中心](README.md) | 🏛️ [Architecture](architecture.md) | 🚗 [Kinematics](kinematics_and_routing.md) | 🛠️ [Developer Guides](developer_guides.md)

---

## 1. Running the Test Suite

```bash
# Execute all tests with coverage analysis
go test -v -cover ./...
```

---

## 2. Table-Driven Test Pattern

```go
func TestKinematicsEulerIntegration(t *testing.T) {
    cases := []struct {
        name          string
        initialSpeed  valueobject.SpeedKmh
        dt            float64
        expectedDist  float64
    }{
        {"Zero speed", 0, 0.1, 0.0},
        {"60 km/h for 1 second", 60, 1.0, 16.666},
        {"90 km/h for 100 ms", 90, 0.1, 2.5},
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            dist := tc.initialSpeed.ToMps() * tc.dt
            if math.Abs(dist - tc.expectedDist) > 0.01 {
                t.Errorf("Expected %.3f m, got %.3f m", tc.expectedDist, dist)
            }
        })
    }
}
```

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>智能交通工程 • EVP Simulator v1.0.0</i>
</div>
