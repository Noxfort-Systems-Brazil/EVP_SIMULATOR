// EVP Simulator is an open-source emergency vehicle preemption and fleet simulation platform for urban traffic networks.
// Copyright (C) 2026 Noxfort Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// File: parser.go
// Author: Gabriel Moraes
// Date: October 2026

package sumo

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/valueobject"
)

// XML structs matching SUMO .net.xml schema
type xmlNet struct {
	XMLName     xml.Name        `xml:"net"`
	Location    xmlLocation     `xml:"location"`
	Edges       []xmlEdge       `xml:"edge"`
	Junctions   []xmlJunction   `xml:"junction"`
	Connections []xmlConnection `xml:"connection"`
	TLLogics    []xmlTLLogic    `xml:"tlLogic"`
}

type xmlLocation struct {
	NetOffset    string `xml:"netOffset,attr"`
	ConvBoundary string `xml:"convBoundary,attr"` // minX,minY,maxX,maxY
	OrigBoundary string `xml:"origBoundary,attr"` // minLon,minLat,maxLon,maxLat
}

type xmlEdge struct {
	ID       string    `xml:"id,attr"`
	From     string    `xml:"from,attr"`
	To       string    `xml:"to,attr"`
	Priority int       `xml:"priority,attr"`
	Function string    `xml:"function,attr"` // skip internal edges (function="internal")
	Lanes    []xmlLane `xml:"lane"`
}

type xmlLane struct {
	ID     string  `xml:"id,attr"`
	Index  int     `xml:"index,attr"`
	Speed  float64 `xml:"speed,attr"`  // meters per second in SUMO
	Length float64 `xml:"length,attr"` // meters
	Shape  string  `xml:"shape,attr"`  // "x1,y1 x2,y2 ..."
}

type xmlJunction struct {
	ID       string  `xml:"id,attr"`
	Type     string  `xml:"type,attr"`
	X        float64 `xml:"x,attr"`
	Y        float64 `xml:"y,attr"`
	IncLanes string  `xml:"incLanes,attr"`
}

type xmlConnection struct {
	From      string `xml:"from,attr"`
	To        string `xml:"to,attr"`
	FromLane  int    `xml:"fromLane,attr"`
	ToLane    int    `xml:"toLane,attr"`
	TL        string `xml:"tl,attr"`
	LinkIndex int    `xml:"linkIndex,attr"`
	Dir       string `xml:"dir,attr"`
}

type xmlTLLogic struct {
	ID     string     `xml:"id,attr"`
	Phases []xmlPhase `xml:"phase"`
}

type xmlPhase struct {
	Duration float64 `xml:"duration,attr"`
	State    string  `xml:"state,attr"`
}

// Parser reads and streams SUMO networks.
type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

// LoadNetworkFromSUMO parses .net.xml or .net.xml.gz into Domain model.
func (p *Parser) LoadNetworkFromSUMO(filePath string) (*entity.RoadNetwork, map[string]*entity.TrafficLight, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var reader io.Reader = file
	if strings.HasSuffix(filePath, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, nil, err
		}
		defer gzReader.Close()
		reader = gzReader
	}

	var netXML xmlNet
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&netXML); err != nil {
		return nil, nil, err
	}

	origParts := parseFloats(netXML.Location.OrigBoundary) // minLon, minLat, maxLon, maxLat
	convParts := parseFloats(netXML.Location.ConvBoundary) // minX, minY, maxX, maxY

	var coordConverter func(x, y float64) valueobject.Coordinate

	// 1. If both convBoundary and origBoundary are present, map coordinates proportionally
	if len(origParts) >= 4 && len(convParts) >= 4 {
		minLon, minLat, maxLon, maxLat := origParts[0], origParts[1], origParts[2], origParts[3]
		minX, minY, maxX, maxY := convParts[0], convParts[1], convParts[2], convParts[3]
		dx := maxX - minX
		dy := maxY - minY

		if dx > 0 && dy > 0 {
			coordConverter = func(x, y float64) valueobject.Coordinate {
				lat := minLat + ((y-minY)/dy)*(maxLat-minLat)
				lon := minLon + ((x-minX)/dx)*(maxLon-minLon)
				return valueobject.Coordinate{X: x, Y: y, Lat: lat, Lon: lon}
			}
		}
	}

	// 2. Fallback: If only origin point is known, project dynamically using local latitude
	if coordConverter == nil && len(origParts) >= 2 {
		baseLon := origParts[0]
		baseLat := origParts[1]
		latMetersPerDeg := 111320.0
		lonMetersPerDeg := latMetersPerDeg * math.Cos(baseLat*(math.Pi/180.0))
		if lonMetersPerDeg < 1.0 {
			lonMetersPerDeg = latMetersPerDeg
		}
		coordConverter = func(x, y float64) valueobject.Coordinate {
			lat := baseLat + (y / latMetersPerDeg)
			lon := baseLon + (x / lonMetersPerDeg)
			return valueobject.Coordinate{X: x, Y: y, Lat: lat, Lon: lon}
		}
	}

	// 3. Final fallback: purely synthetic network with no GPS boundary
	if coordConverter == nil {
		coordConverter = func(x, y float64) valueobject.Coordinate {
			return valueobject.Coordinate{X: x, Y: y, Lat: 0, Lon: 0}
		}
	}

	network := entity.NewRoadNetwork()

	// 1. Parse Junctions / Nodes
	for _, j := range netXML.Junctions {
		if strings.HasPrefix(j.ID, ":") {
			continue // skip internal junction
		}
		var incLanes []string
		if j.IncLanes != "" {
			incLanes = strings.Fields(j.IncLanes)
		}
		network.Nodes[j.ID] = &entity.Junction{
			ID:       j.ID,
			Type:     j.Type,
			Position: coordConverter(j.X, j.Y),
			IncLanes: incLanes,
		}
	}

	// 2. Parse Edges & Lanes
	for _, e := range netXML.Edges {
		if e.Function == "internal" || strings.HasPrefix(e.ID, ":") {
			continue // skip internal connectors
		}

		edge := &entity.Edge{
			ID:       e.ID,
			FromNode: e.From,
			ToNode:   e.To,
			Priority: e.Priority,
			NumLanes: len(e.Lanes),
			Lanes:    make([]*entity.Lane, 0, len(e.Lanes)),
		}

		for _, l := range e.Lanes {
			shapeCoords := parseShape(l.Shape, coordConverter)
			lane := &entity.Lane{
				ID:     l.ID,
				Index:  l.Index,
				Speed:  valueobject.FromMps(l.Speed),
				Length: l.Length,
				Shape:  shapeCoords,
			}
			edge.Lanes = append(edge.Lanes, lane)
			if l.Length > edge.Length {
				edge.Length = l.Length
			}
			if lane.Speed > edge.Speed {
				edge.Speed = lane.Speed
			}
		}

		network.Edges[edge.ID] = edge
	}

	// 3. Parse Connections
	for _, c := range netXML.Connections {
		if strings.HasPrefix(c.From, ":") || strings.HasPrefix(c.To, ":") {
			continue
		}
		conn := &entity.Connection{
			FromEdge:       c.From,
			ToEdge:         c.To,
			FromLane:       c.FromLane,
			ToLane:         c.ToLane,
			TrafficLightID: c.TL,
			LinkIndex:      c.LinkIndex,
			Dir:            c.Dir,
		}
		network.AddConnection(conn)
	}

	// 4. Parse Traffic Light Logic
	trafficLights := make(map[string]*entity.TrafficLight)
	portOffset := 0
	for _, tl := range netXML.TLLogics {
		if _, exists := trafficLights[tl.ID]; exists {
			continue // keep first definition if duplicate ID
		}
		phases := make([]entity.Phase, 0, len(tl.Phases))
		for _, p := range tl.Phases {
			phases = append(phases, entity.Phase{
				Duration: p.Duration,
				State:    p.State,
			})
		}
		tLight := entity.NewTrafficLight(tl.ID, phases)
		tLight.IntersectionID = tl.ID
		tLight.UTMCPort = 1610 + portOffset
		tLight.IPAddress = fmt.Sprintf("127.0.0.1:%d", 1610+portOffset)
		trafficLights[tl.ID] = tLight
		portOffset++
	}

	return network, trafficLights, nil
}

func parseShape(shapeStr string, conv func(x, y float64) valueobject.Coordinate) []valueobject.Coordinate {
	var res []valueobject.Coordinate
	if shapeStr == "" {
		return res
	}

	pairs := strings.Fields(shapeStr)
	for _, pair := range pairs {
		parts := strings.Split(pair, ",")
		if len(parts) == 2 {
			x, errX := strconv.ParseFloat(parts[0], 64)
			y, errY := strconv.ParseFloat(parts[1], 64)
			if errX == nil && errY == nil {
				res = append(res, conv(x, y))
			}
		}
	}
	return res
}

func parseFloats(str string) []float64 {
	if str == "" {
		return nil
	}
	parts := strings.Split(str, ",")
	res := make([]float64, 0, len(parts))
	for _, p := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil
		}
		res = append(res, val)
	}
	return res
}
