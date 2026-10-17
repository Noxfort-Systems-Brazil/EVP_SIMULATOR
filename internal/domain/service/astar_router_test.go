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

// File: astar_router_test.go
// Author: Gabriel Moraes
// Date: October 2026

package service

import (
	"testing"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/valueobject"
)

func TestAStarRouter_FindRoute(t *testing.T) {
	network := entity.NewRoadNetwork()

	// Create nodes J1, J2, J3, J4
	network.Nodes["J1"] = &entity.Junction{ID: "J1", Position: valueobject.Coordinate{X: 0, Y: 0}}
	network.Nodes["J2"] = &entity.Junction{ID: "J2", Position: valueobject.Coordinate{X: 100, Y: 0}}
	network.Nodes["J3"] = &entity.Junction{ID: "J3", Position: valueobject.Coordinate{X: 200, Y: 0}}
	network.Nodes["J4"] = &entity.Junction{ID: "J4", Position: valueobject.Coordinate{X: 100, Y: 100}}

	// Edge 1: J1 -> J2
	network.Edges["E1"] = &entity.Edge{
		ID: "E1", FromNode: "J1", ToNode: "J2", Length: 100, Speed: 50, NumLanes: 2,
	}
	// Edge 2 (Direct): J2 -> J3 (length 100m)
	network.Edges["E2"] = &entity.Edge{
		ID: "E2", FromNode: "J2", ToNode: "J3", Length: 100, Speed: 50, NumLanes: 2,
	}
	// Edge 3 (Detour via J4): J2 -> J4 (length 100m)
	network.Edges["E3"] = &entity.Edge{
		ID: "E3", FromNode: "J2", ToNode: "J4", Length: 100, Speed: 50, NumLanes: 1,
	}
	// Edge 4: J4 -> J3 (length 100m)
	network.Edges["E4"] = &entity.Edge{
		ID: "E4", FromNode: "J4", ToNode: "J3", Length: 100, Speed: 50, NumLanes: 1,
	}

	network.AddConnection(&entity.Connection{FromEdge: "E1", ToEdge: "E2"})
	network.AddConnection(&entity.Connection{FromEdge: "E1", ToEdge: "E3"})
	network.AddConnection(&entity.Connection{FromEdge: "E3", ToEdge: "E4"})

	router := NewAStarRouter()

	// Should select direct path: E1 -> E2
	route := router.FindRoute(network, "E1", "E2", valueobject.ModeNormal)
	if len(route) != 2 || route[0] != "E1" || route[1] != "E2" {
		t.Fatalf("expected route [E1, E2], got %v", route)
	}

	// Should route via detour when target is E4: E1 -> E3 -> E4
	routeDetour := router.FindRoute(network, "E1", "E4", valueobject.ModeAlert)
	if len(routeDetour) != 3 || routeDetour[0] != "E1" || routeDetour[1] != "E3" || routeDetour[2] != "E4" {
		t.Fatalf("expected route [E1, E3, E4], got %v", routeDetour)
	}
}
