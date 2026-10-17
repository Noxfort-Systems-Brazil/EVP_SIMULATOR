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

// File: kinematics_test.go
// Author: Gabriel Moraes
// Date: October 2026

package service

import (
	"testing"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/valueobject"
)

func TestKinematicsEngine_StopAtRedLight(t *testing.T) {
	router := NewAStarRouter()
	evaluator := NewSignalEvaluator()
	engine := NewKinematicsEngine(router, evaluator)

	network := entity.NewRoadNetwork()
	// Edge 1 (length 100m) -> Edge 2
	lane1 := &entity.Lane{
		ID:     "E1_0",
		Index:  0,
		Length: 100,
		Speed:  50,
		Shape: []valueobject.Coordinate{
			{X: 0, Y: 0},
			{X: 100, Y: 0},
		},
	}
	network.Edges["E1"] = &entity.Edge{
		ID: "E1", Length: 100, Speed: 50, Lanes: []*entity.Lane{lane1},
	}
	network.Edges["E2"] = &entity.Edge{
		ID: "E2", Length: 100, Speed: 50,
	}

	network.AddConnection(&entity.Connection{
		FromEdge:       "E1",
		ToEdge:         "E2",
		FromLane:       0,
		ToLane:         0,
		TrafficLightID: "TL_TEST",
		LinkIndex:      0,
	})

	// Traffic light in Red state
	tl := entity.NewTrafficLight("TL_TEST", []entity.Phase{{Duration: 30, State: "r"}})
	trafficLights := map[string]*entity.TrafficLight{"TL_TEST": tl}

	// 1. Vehicle in Normal Mode close to signal (at 80m of 100m, remaining: 20m)
	veh := entity.NewVehicle("SAMU-01", "SAMU", valueobject.Coordinate{X: 80, Y: 0})
	veh.CurrentEdgeID = "E1"
	veh.CurrentLaneIdx = 0
	veh.LaneLength = 100
	veh.ProgressOnLane = 80
	veh.Speed = 40.0
	veh.Waypoints = lane1.Shape
	veh.WaypointIdx = 0
	veh.Route = []string{"E1", "E2"}

	// Tick forward (4.0 seconds for complete braking)
	for i := 0; i < 40; i++ {
		engine.UpdateVehicle(veh, network, trafficLights, 0.1)
	}

	// In Normal Mode, vehicle must stop at the red light
	if veh.Speed > 1.0 {
		t.Errorf("expected vehicle to decelerate to stop for red light, got speed %.2f km/h", veh.Speed)
	}
	if !veh.WaitingAtRed {
		t.Errorf("expected vehicle to be waiting at red light")
	}

	// 2. Switch to Alert Mode (Sirene ON) and turn signal Green
	veh.SetMode(valueobject.ModeAlert)
	tl.Phases[0].State = "G"

	for i := 0; i < 20; i++ {
		engine.UpdateVehicle(veh, network, trafficLights, 0.1)
	}

	if veh.WaitingAtRed {
		t.Errorf("expected vehicle to resume movement after signal turns green")
	}
	if veh.Speed <= 0 {
		t.Errorf("expected vehicle to accelerate after signal turns green")
	}
}
