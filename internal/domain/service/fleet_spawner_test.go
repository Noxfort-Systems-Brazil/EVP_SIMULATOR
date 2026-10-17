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

// File: fleet_spawner_test.go
// Author: Gabriel Moraes
// Date: October 2026

package service

import (
	"testing"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/valueobject"
)

func TestFleetSpawner_SpawnEmergencyFleet(t *testing.T) {
	router := NewAStarRouter()
	spawner := NewFleetSpawner(router)

	// Empty network
	vehs := spawner.SpawnEmergencyFleet(nil)
	if len(vehs) != 0 {
		t.Fatalf("expected 0 vehicles on nil network")
	}

	// Network with valid edge
	net := entity.NewRoadNetwork()
	net.Edges["E1"] = &entity.Edge{
		ID:     "E1",
		Lanes: []*entity.Lane{
			{
				ID:     "E1_0",
				Index:  0,
				Length: 100.0,
				Shape: []valueobject.Coordinate{
					{X: 0, Y: 0},
					{X: 100, Y: 0},
				},
			},
		},
	}

	vehs = spawner.SpawnEmergencyFleet(net)
	if len(vehs) != 3 {
		t.Fatalf("expected 3 spawned vehicles, got %d", len(vehs))
	}

	expectedIDs := map[string]string{
		"SAMU-01":     "SAMU",
		"BOMBEIRO-01": "BOMBEIROS",
		"POLICIA-01":  "POLICIA",
	}

	for id, expectedCorp := range expectedIDs {
		v, exists := vehs[id]
		if !exists {
			t.Fatalf("expected vehicle %s to exist", id)
		}
		if v.Corporation != expectedCorp {
			t.Errorf("expected corp %s, got %s", expectedCorp, v.Corporation)
		}
		if v.Mode != valueobject.ModeAlert {
			t.Errorf("expected vehicle %s to spawn in ModeAlert, got %s", id, v.Mode)
		}
		if v.TargetSpeed != v.AlertLimit {
			t.Errorf("expected vehicle %s target speed to be alert limit %.1f, got %.1f", id, v.AlertLimit, v.TargetSpeed)
		}
		if v.CurrentEdgeID != "E1" {
			t.Errorf("expected edge E1, got %s", v.CurrentEdgeID)
		}
		if len(v.Waypoints) != 2 {
			t.Errorf("expected 2 waypoints, got %d", len(v.Waypoints))
		}
	}
}
