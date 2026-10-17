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

// File: network.go
// Author: Gabriel Moraes
// Date: October 2026

package entity

import (
	"evp_simulator/internal/domain/valueobject"
)

// Lane represents a single physical lane on an edge.
type Lane struct {
	ID     string                   `json:"id"`
	Index  int                      `json:"index"`
	Speed  valueobject.SpeedKmh     `json:"speed"`
	Length float64                  `json:"length"`
	Shape  []valueobject.Coordinate `json:"shape"`
}

// Edge represents a directed street segment with one or more lanes.
type Edge struct {
	ID       string   `json:"id"`
	FromNode string   `json:"from_node"`
	ToNode   string   `json:"to_node"`
	Priority int      `json:"priority"`
	NumLanes int      `json:"num_lanes"`
	Lanes    []*Lane  `json:"lanes"`
	Length   float64  `json:"length"`
	Speed    valueobject.SpeedKmh `json:"speed"`
}

// Connection represents a valid turning movement from an incoming lane to an outgoing lane.
type Connection struct {
	FromEdge       string `json:"from_edge"`
	ToEdge         string `json:"to_edge"`
	FromLane       int    `json:"from_lane"`
	ToLane         int    `json:"to_lane"`
	TrafficLightID string `json:"traffic_light_id"`
	LinkIndex      int    `json:"link_index"` // Index inside the tlLogic phase state string!
	Dir            string `json:"dir"`
}

// Junction represents an intersection (can be traffic light, priority, right-before-left, etc.).
type Junction struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Position valueobject.Coordinate `json:"position"`
	IncLanes []string               `json:"inc_lanes"`
}

// RoadNetwork is the aggregate root containing the entire street topology.
type RoadNetwork struct {
	Nodes       map[string]*Junction               `json:"nodes"`
	Edges       map[string]*Edge                   `json:"edges"`
	Connections map[string][]*Connection           `json:"connections"` // Key: FromEdge
	ConnLookup  map[string]*Connection             // Key: "fromEdge:fromLane:toEdge" -> Connection
}

// NewRoadNetwork creates a blank RoadNetwork instance.
func NewRoadNetwork() *RoadNetwork {
	return &RoadNetwork{
		Nodes:       make(map[string]*Junction),
		Edges:       make(map[string]*Edge),
		Connections: make(map[string][]*Connection),
		ConnLookup:  make(map[string]*Connection),
	}
}

// AddConnection registers a connection in both the adjacency list and lookup map.
func (rn *RoadNetwork) AddConnection(conn *Connection) {
	rn.Connections[conn.FromEdge] = append(rn.Connections[conn.FromEdge], conn)
	key := conn.FromEdge + ":" + string(rune('0'+conn.FromLane)) + ":" + conn.ToEdge
	rn.ConnLookup[key] = conn
}

// FindConnection returns the connection between two edges and lanes, if valid.
func (rn *RoadNetwork) FindConnection(fromEdge string, fromLane int, toEdge string) *Connection {
	key := fromEdge + ":" + string(rune('0'+fromLane)) + ":" + toEdge
	return rn.ConnLookup[key]
}
