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

// File: kinematics_engine.go
// Author: Gabriel Moraes
// Date: October 2026

package service

import (
	"math/rand"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/valueobject"
)

// KinematicsEngine orchestrates vehicle movement, stops at red lights, and transitions between edges.
type KinematicsEngine struct {
	router          *AStarRouter
	signalEvaluator *SignalEvaluator
}

// NewKinematicsEngine creates a new KinematicsEngine instance.
func NewKinematicsEngine(router *AStarRouter, evaluator *SignalEvaluator) *KinematicsEngine {
	return &KinematicsEngine{
		router:          router,
		signalEvaluator: evaluator,
	}
}

// UpdateVehicle processes one simulation step of duration dt for a vehicle.
func (ke *KinematicsEngine) UpdateVehicle(
	veh *entity.Vehicle,
	network *entity.RoadNetwork,
	trafficLights map[string]*entity.TrafficLight,
	dt float64,
) {
	if veh.CurrentEdgeID == "" {
		return
	}

	currEdge := network.Edges[veh.CurrentEdgeID]
	if currEdge == nil || len(currEdge.Lanes) == 0 {
		return
	}

	// Determine next edge in route
	var nextEdgeID string
	if veh.RouteIdx+1 < len(veh.Route) {
		nextEdgeID = veh.Route[veh.RouteIdx+1]
	}

	// 1. Evaluate Traffic Light state for upcoming intersection
	distToEndOfLane := veh.LaneLength - veh.ProgressOnLane
	if distToEndOfLane < 0 {
		distToEndOfLane = 0
	}

	sigState, _, _ := ke.signalEvaluator.EvaluateApproach(veh, network, trafficLights, nextEdgeID)

	desiredSpeed := veh.TargetSpeed

	// 2. Behavioral Response to Traffic Signal
	if sigState == valueobject.SignalRed {
		if !veh.Mode.IsAlert() {
			// Normal Mode: Strict compliance. Stop before crossing line.
			if distToEndOfLane < 35.0 {
				desiredSpeed = 0.0
				if distToEndOfLane < 3.0 {
					veh.Speed = 0.0
					veh.WaitingAtRed = true
				}
			}
		} else {
			// Alert Mode: Caution deceleration if signal has not turned green yet.
			veh.WaitingAtRed = false
			if distToEndOfLane < 30.0 {
				desiredSpeed = valueobject.SpeedKmh(25.0) // 25 km/h caution approach
			}
		}
	} else if sigState == valueobject.SignalYellow {
		if !veh.Mode.IsAlert() && distToEndOfLane > 20.0 {
			desiredSpeed = 0.0 // Slow down if not already inside junction
		}
	} else {
		// Signal is Green or unmonitored
		veh.WaitingAtRed = false
	}

	// 3. Update speed using vehicle inertia (acceleration/deceleration)
	veh.UpdateKinematics(dt, desiredSpeed)

	// 4. Advance spatial position along waypoints
	distTraveled := veh.Speed.ToMps() * dt
	reachedEndOfLane := veh.StepAlongWaypoints(distTraveled)

	// 5. Handle Edge Transition when reaching the end of the current lane
	if reachedEndOfLane && (!veh.WaitingAtRed || veh.Speed > 0) {
		ke.advanceToNextEdge(veh, network)
	}
}

// advanceToNextEdge switches the vehicle to the next edge on its route or generates a new mission.
func (ke *KinematicsEngine) advanceToNextEdge(veh *entity.Vehicle, network *entity.RoadNetwork) {
	if veh.RouteIdx+1 < len(veh.Route) {
		veh.RouteIdx++
		nextEdgeID := veh.Route[veh.RouteIdx]
		nextEdge := network.Edges[nextEdgeID]
		if nextEdge != nil && len(nextEdge.Lanes) > 0 {
			lane := nextEdge.Lanes[0] // pick primary lane
			veh.CurrentEdgeID = nextEdgeID
			veh.CurrentLaneIdx = 0
			veh.ProgressOnLane = 0
			veh.LaneLength = lane.Length
			veh.Waypoints = lane.Shape
			veh.WaypointIdx = 0
			if len(lane.Shape) > 0 {
				veh.Position = lane.Shape[0]
			}
			return
		}
	}

	// Reached end of route: pick a random edge and compute a new mission with A*
	ke.assignNewRandomRoute(veh, network)
}

// assignNewRandomRoute picks a random destination edge and calculates a route using A*.
func (ke *KinematicsEngine) assignNewRandomRoute(veh *entity.Vehicle, network *entity.RoadNetwork) {
	if len(network.Edges) == 0 {
		return
	}

	allEdges := make([]string, 0, len(network.Edges))
	for edgeID := range network.Edges {
		allEdges = append(allEdges, edgeID)
	}

	randomTarget := allEdges[rand.Intn(len(allEdges))]
	newRoute := ke.router.FindRoute(network, veh.CurrentEdgeID, randomTarget, veh.Mode)

	if len(newRoute) > 1 {
		veh.Route = newRoute
		veh.RouteIdx = 0
	}
}
