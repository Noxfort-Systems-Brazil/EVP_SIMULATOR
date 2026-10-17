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

// File: signal_evaluator_test.go
// Author: Gabriel Moraes
// Date: October 2026

package service

import (
	"testing"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/valueobject"
)

func TestSignalEvaluator_EvaluateApproach(t *testing.T) {
	evaluator := NewSignalEvaluator()

	// Build network with connection on linkIndex 1
	network := entity.NewRoadNetwork()
	conn := &entity.Connection{
		FromEdge:       "edge_A",
		ToEdge:         "edge_B",
		FromLane:       0,
		ToLane:         0,
		TrafficLightID: "TL_1",
		LinkIndex:      1,
	}
	network.AddConnection(conn)

	// Create Traffic Light with 2 links: link 0 is Red, link 1 is Green
	phases := []entity.Phase{
		{Duration: 30, State: "rG"},
		{Duration: 30, State: "Gr"},
	}
	tl := entity.NewTrafficLight("TL_1", phases)
	trafficLights := map[string]*entity.TrafficLight{"TL_1": tl}

	// Vehicle on edge_A lane 0
	veh := entity.NewVehicle("SAMU-01", "SAMU", valueobject.Coordinate{})
	veh.CurrentEdgeID = "edge_A"
	veh.CurrentLaneIdx = 0

	// 1. In phase 0 ("rG"): link 1 is 'G' -> Green!
	sigState, foundTL, linkIdx := evaluator.EvaluateApproach(veh, network, trafficLights, "edge_B")
	if foundTL == nil || linkIdx != 1 {
		t.Fatalf("expected TL_1 and linkIndex 1, got %v, %d", foundTL, linkIdx)
	}
	if sigState != valueobject.SignalGreen {
		t.Errorf("expected SignalGreen, got %v", sigState)
	}

	// 2. Change phase to 1 ("Gr"): link 1 is 'r' -> Red!
	tl.CurrentPhaseIdx = 1
	sigStateRed, _, _ := evaluator.EvaluateApproach(veh, network, trafficLights, "edge_B")
	if sigStateRed != valueobject.SignalRed {
		t.Errorf("expected SignalRed, got %v", sigStateRed)
	}
}
