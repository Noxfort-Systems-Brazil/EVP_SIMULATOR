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

// File: signal_evaluator.go
// Author: Gabriel Moraes
// Date: October 2026

package service

import (
	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/valueobject"
)

// SignalEvaluator checks whether an approaching vehicle has a green, yellow, or red light on its specific lane.
type SignalEvaluator struct{}

// NewSignalEvaluator creates a new SignalEvaluator instance.
func NewSignalEvaluator() *SignalEvaluator {
	return &SignalEvaluator{}
}

// EvaluateApproach evaluates the signal state for a vehicle approaching an intersection.
func (se *SignalEvaluator) EvaluateApproach(
	veh *entity.Vehicle,
	network *entity.RoadNetwork,
	trafficLights map[string]*entity.TrafficLight,
	nextEdgeID string,
) (valueobject.SignalState, *entity.TrafficLight, int) {
	if veh.CurrentEdgeID == "" || nextEdgeID == "" {
		return valueobject.SignalGreen, nil, -1
	}

	conn := network.FindConnection(veh.CurrentEdgeID, veh.CurrentLaneIdx, nextEdgeID)
	if conn == nil || conn.TrafficLightID == "" {
		return valueobject.SignalGreen, nil, -1 // No traffic light on this movement
	}

	tl, exists := trafficLights[conn.TrafficLightID]
	if !exists {
		return valueobject.SignalGreen, nil, -1
	}

	sigState := tl.GetSignalForLink(conn.LinkIndex)
	return sigState, tl, conn.LinkIndex
}
