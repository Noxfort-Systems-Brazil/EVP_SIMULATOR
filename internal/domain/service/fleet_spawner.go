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

// File: fleet_spawner.go
// Author: Gabriel Moraes
// Date: October 2026

package service

import (
	"math/rand"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/valueobject"
)

// FleetSpawner encapsulates the domain logic of creating and deploying emergency fleets on a road network.
type FleetSpawner struct {
	router *AStarRouter
}

// NewFleetSpawner creates a new FleetSpawner instance.
func NewFleetSpawner(router *AStarRouter) *FleetSpawner {
	return &FleetSpawner{
		router: router,
	}
}

// VehicleConfig defines parameters for a vehicle to be spawned.
type VehicleConfig struct {
	ID   string
	Corp string
	Mode valueobject.VehicleMode
}

// DefaultEmergencyFleetConfigs returns the standard emergency fleet configuration.
// By default, vehicles are configured to spawn in ModeAlert for immediate emergency readiness and EVP preemption.
func DefaultEmergencyFleetConfigs() []VehicleConfig {
	return []VehicleConfig{
		{ID: "SAMU-01", Corp: "SAMU", Mode: valueobject.ModeAlert},
		{ID: "BOMBEIRO-01", Corp: "BOMBEIROS", Mode: valueobject.ModeAlert},
		{ID: "POLICIA-01", Corp: "POLICIA", Mode: valueobject.ModeAlert},
	}
}

// SpawnEmergencyFleet instantiates emergency vehicles on valid edges with initial routes.
func (fs *FleetSpawner) SpawnEmergencyFleet(network *entity.RoadNetwork) map[string]*entity.Vehicle {
	vehicles := make(map[string]*entity.Vehicle)
	if network == nil || len(network.Edges) == 0 {
		return vehicles
	}

	edgeList := make([]*entity.Edge, 0, len(network.Edges))
	for _, edge := range network.Edges {
		if len(edge.Lanes) > 0 && len(edge.Lanes[0].Shape) > 0 {
			edgeList = append(edgeList, edge)
		}
	}
	if len(edgeList) == 0 {
		return vehicles
	}

	configs := DefaultEmergencyFleetConfigs()

	for _, cfg := range configs {
		edge := edgeList[rand.Intn(len(edgeList))]
		lane := edge.Lanes[0]

		v := entity.NewVehicle(cfg.ID, cfg.Corp, lane.Shape[0])

		targetMode := cfg.Mode
		if targetMode == "" {
			targetMode = valueobject.ModeAlert
		}
		v.SetMode(targetMode)

		v.CurrentEdgeID = edge.ID
		v.CurrentLaneIdx = 0
		v.LaneLength = lane.Length
		v.Waypoints = lane.Shape
		v.WaypointIdx = 0

		targetEdge := edgeList[rand.Intn(len(edgeList))]
		route := fs.router.FindRoute(network, edge.ID, targetEdge.ID, v.Mode)
		v.Route = route
		v.RouteIdx = 0

		vehicles[v.ID] = v
	}

	return vehicles
}
