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

// File: state_projector.go
// Author: Gabriel Moraes
// Date: October 2026

package usecase

import (
	"sort"

	"evp_simulator/internal/application/dto"
	"evp_simulator/internal/domain/entity"
)

// StateProjector converts internal simulation domain entities into presentation DTOs for the UI.
type StateProjector struct{}

// NewStateProjector creates a new StateProjector instance.
func NewStateProjector() *StateProjector {
	return &StateProjector{}
}

// Project compiles the snapshot of the current simulation into a SimulationStateDTO.
func (p *StateProjector) Project(
	network *entity.RoadNetwork,
	trafficLights map[string]*entity.TrafficLight,
	vehicles map[string]*entity.Vehicle,
	isRunning bool,
	isExporting bool,
	logs []string,
	telemetryURL string,
) *dto.SimulationStateDTO {
	vehiclesDTO := make([]dto.VehicleStateDTO, 0, len(vehicles))
	for _, v := range vehicles {
		vehiclesDTO = append(vehiclesDTO, dto.VehicleStateDTO{
			ID:            v.ID,
			Corporation:   v.Corporation,
			Mode:          string(v.Mode),
			SpeedKmh:      v.Speed.Float64(),
			HeadingDeg:    v.HeadingDeg,
			PosX:          v.Position.X,
			PosY:          v.Position.Y,
			Latitude:      v.Position.Lat,
			Longitude:     v.Position.Lon,
			WaitingAtRed:  v.WaitingAtRed,
			CurrentEdgeID: v.CurrentEdgeID,
		})
	}
	sort.Slice(vehiclesDTO, func(i, j int) bool {
		return vehiclesDTO[i].ID < vehiclesDTO[j].ID
	})

	lightsDTO := make([]dto.TrafficLightStateDTO, 0, len(trafficLights))
	for _, tl := range trafficLights {
		var posX, posY float64
		if network != nil {
			if node, ok := network.Nodes[tl.ID]; ok {
				posX = node.Position.X
				posY = node.Position.Y
			}
		}

		lightsDTO = append(lightsDTO, dto.TrafficLightStateDTO{
			ID:              tl.ID,
			IntersectionID:  tl.IntersectionID,
			IPAddress:       tl.IPAddress,
			CurrentPhaseIdx: tl.CurrentPhaseIdx,
			TransitionState: tl.TransitionState,
			StateString:     tl.GetCurrentState(),
			OverrideState:   tl.OverrideState,
			PosX:            posX,
			PosY:            posY,
		})
	}
	sort.Slice(lightsDTO, func(i, j int) bool {
		return lightsDTO[i].ID < lightsDTO[j].ID
	})

	totalEdges := 0
	totalLanes := 0
	if network != nil {
		totalEdges = len(network.Edges)
		for _, e := range network.Edges {
			totalLanes += len(e.Lanes)
		}
	}

	mappedCount := 0
	for _, tl := range trafficLights {
		if tl.IntersectionID != "" {
			mappedCount++
		}
	}

	return &dto.SimulationStateDTO{
		IsRunning:         isRunning,
		IsExporting:       isExporting,
		MapLoaded:         network != nil,
		TotalEdges:        totalEdges,
		TotalLanes:        totalLanes,
		TotalLights:       len(trafficLights),
		MappedControllers: mappedCount,
		Vehicles:          vehiclesDTO,
		TrafficLights:     lightsDTO,
		RecentLogs:        append([]string(nil), logs...),
		TelemetryURL:      telemetryURL,
	}
}
