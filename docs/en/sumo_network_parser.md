---
tags: [sumo, parser, xml, gzip, projection, coordinates, road-network]
aliases: [SUMO Parser, Network Ingestion, SUMO Network Parser]
---

# 🗺️ SUMO Road Network Parser & Geometry Ingestion (`.net.xml`)

This document specifies the road network ingestion subsystem of the EVP Simulator, implemented in `internal/infrastructure/sumo/parser.go`.

⬅️ [Documentation Hub](README.md) | 🏛️ [Architecture](architecture.md) | 🚗 [Kinematics](kinematics_and_routing.md) | 📊 [CSV Exchanger](csv_data_exchanger.md)

---

## 1. Transparent Compressed & Plain XML Parsing (`.net.xml.gz`)

The parser inspects file headers for gzip magic bytes (`0x1F 0x8B`) and streams decompressed XML transparently:

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

## 2. Topological Element Mapping

| SUMO XML Tag | Go Struct (`entity`) | Description |
| :--- | :--- | :--- |
| `<location>` | `BoundingBox`, `netOffset` | Cartesian and geodetic bounding box boundaries. |
| `<junction>` | `entity.Junction` | Intersection geometry and signal controller anchors. |
| `<edge>` | `entity.Edge` | Directional arterial road segments containing lanes. |
| `<lane>` | `entity.Lane` | Individual vehicular lanes with speed limits and polyline shape. |
| `<connection>` | `entity.Connection` | Allowed turning movements between incoming and outgoing lanes. |
| `<tlLogic>` | `entity.TrafficLight` | Traffic signal plan logic and phase sequence definitions. |

---

## 3. Spatial Projection & Bounding Box Calculation

Cartesian meters $(X, Y)$ are mapped to WGS84 coordinates using `convBoundary` offsets:

$$\text{lat} = \text{lat}_0 + \frac{Y}{111\,139}, \quad \text{lon} = \text{lon}_0 + \frac{X}{111\,139 \cdot \cos(\text{lat}_0)}$$

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Intelligent Mobility Engineering • EVP Simulator v1.0.0</i>
</div>
