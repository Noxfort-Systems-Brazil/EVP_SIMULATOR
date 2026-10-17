---
tags: [sumo, parser, xml, gzip, projection, coordinates, road-network]
aliases: [Parser SUMO, Ingestão de Rede, SUMO Network Parser]
---

# 🗺️ Ingestão e Parser de Redes Viárias SUMO (`.net.xml`)

Este documento especifica o subsistema de ingestão de malhas viárias urbanas do EVP Simulator, implementado em `internal/infrastructure/sumo/parser.go`.

⬅️ [Central de Documentação](README.md) | 🏛️ [Arquitetura](architecture.md) | 🚗 [Cinemática & Roteamento](kinematics_and_routing.md) | 📊 [CSV Exchanger](csv_data_exchanger.md)

---

## 1. Suporte a Arquivos Descompactados e Comprimidos (`.net.xml.gz`)

O parser detecta automaticamente o formato do arquivo por meio dos *magic bytes* iniciais (`0x1F 0x8B`) ou da extensão do arquivo. Caso seja detectado um arquivo comprimido, a leitura é canalizada de forma transparente através do pacote nativo `compress/gzip`:

```go
func (p *Parser) LoadNetwork(filePath string) (*entity.RoadNetwork, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var reader io.Reader = file
    if strings.HasSuffix(filePath, ".gz") {
        gzReader, err := gzip.NewReader(file)
        if err != nil {
            return nil, err
        }
        defer gzReader.Close()
        reader = gzReader
    }
    return p.parseXML(reader)
}
```

---

## 2. Extração dos Elementos Topológicos

O parser processa o schema XML nativo do SUMO, mapeando as seguintes estruturas:

| Tag XML SUMO | Estrutura Go (`entity`) | Descrição |
| :--- | :--- | :--- |
| `<location>` | `BoundingBox`, `netOffset` | Caixa delimitadora cartesiana e geodésica do mapa. |
| `<junction>` | `entity.Junction` | Cruzamentos viários e interseções semaforizadas. |
| `<edge>` | `entity.Edge` | Trechos de vias conectando junções, contendo faixas. |
| `<lane>` | `entity.Lane` | Faixas de rolamento com velocidade máxima e geometria em linha poligonal. |
| `<connection>` | `entity.Connection` | Movimentos permitidos entre faixas de entrada e saída. |
| `<tlLogic>` | `entity.TrafficLight` | Lógica semafórica, definição de planos e fases. |

### 2.1 Conversão de Strings de Forma (*Shape Polylines*)
As geometrias de faixas vêm serializadas no atributo `shape="x1,y1 x2,y2 x3,y3"`. O parser realiza *string splitting* em alta performance gerando fatias contíguas de `valueobject.Coordinate`:

```go
func parseShape(raw string) []valueobject.Coordinate {
    points := strings.Fields(raw)
    coords := make([]valueobject.Coordinate, 0, len(points))
    for _, pt := range points {
        parts := strings.Split(pt, ",")
        if len(parts) == 2 {
            x, _ := strconv.ParseFloat(parts[0], 64)
            y, _ := strconv.ParseFloat(parts[1], 64)
            coords = append(coords, valueobject.Coordinate{X: x, Y: y})
        }
    }
    return coords
}
```

---

## 3. Normalização e Projeção de Coordenadas

Para permitir tanto a renderização em canvas cartesiano $(X, Y)$ em metros quanto o envio de coordenadas geodésicas (Latitude, Longitude em WGS84) para sistemas de despacho de emergência e para a CARINA CORE, o parser armazena a matriz de projeção `convBoundary` e `origBoundary`:

$$\text{lat} = \text{lat}_0 + \frac{Y}{111\,139}, \quad \text{lon} = \text{lon}_0 + \frac{X}{111\,139 \cdot \cos(\text{lat}_0)}$$

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Engenharia de Mobilidade Inteligente • EVP Simulator v1.0.0</i>
  <small>SPDX-License-Identifier: <a href="../../LICENSE">Apache-2.0</a> • &copy; 2026 Noxfort Systems</small>
</div>
