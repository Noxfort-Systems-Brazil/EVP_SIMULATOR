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

// File: simulation_dto.go
// Author: Gabriel Moraes
// Date: October 2026

package dto

// VehicleStateDTO represents the live state of a vehicle for UI display.
type VehicleStateDTO struct {
	ID             string  `json:"id"`
	Corporation    string  `json:"corporation"`
	Mode           string  `json:"mode"`
	SpeedKmh       float64 `json:"speed_kmh"`
	HeadingDeg     float64 `json:"heading_deg"`
	PosX           float64 `json:"pos_x"`
	PosY           float64 `json:"pos_y"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	WaitingAtRed   bool    `json:"waiting_at_red"`
	CurrentEdgeID  string  `json:"current_edge_id"`
}

// TrafficLightStateDTO represents the live state of a traffic light for UI display.
type TrafficLightStateDTO struct {
	ID              string  `json:"id"`
	IntersectionID  string  `json:"intersection_id"`
	IPAddress       string  `json:"ip_address"`
	CurrentPhaseIdx int     `json:"current_phase_idx"`
	TransitionState string  `json:"transition_state"`
	StateString     string  `json:"state_string"`
	OverrideState   string  `json:"override_state"`
	PosX            float64 `json:"pos_x"`
	PosY            float64 `json:"pos_y"`
}

// SimulationStateDTO represents the overall state sent to the UI.
type SimulationStateDTO struct {
	IsRunning          bool                   `json:"is_running"`
	IsExporting        bool                   `json:"is_exporting"`
	MapLoaded          bool                   `json:"map_loaded"`
	TotalEdges         int                    `json:"total_edges"`
	TotalLanes         int                    `json:"total_lanes"`
	TotalLights        int                    `json:"total_lights"`
	MappedControllers  int                    `json:"mapped_controllers"`
	Vehicles           []VehicleStateDTO      `json:"vehicles"`
	TrafficLights      []TrafficLightStateDTO `json:"traffic_lights"`
	RecentLogs         []string               `json:"recent_logs"`
	TelemetryURL       string                 `json:"telemetry_url"`
}
