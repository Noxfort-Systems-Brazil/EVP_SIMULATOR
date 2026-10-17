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

// File: vehicle.go
// Author: Gabriel Moraes
// Date: October 2026

package entity

import (
	"math"

	"evp_simulator/internal/domain/valueobject"
)

// Vehicle represents an emergency fleet vehicle moving across the road network.
type Vehicle struct {
	ID          string                  `json:"id"`
	Corporation string                  `json:"corporation"`
	Mode        valueobject.VehicleMode `json:"mode"`

	Position   valueobject.Coordinate `json:"position"`
	Speed      valueobject.SpeedKmh   `json:"speed"`
	HeadingDeg float64                `json:"heading_deg"`

	TargetSpeed valueobject.SpeedKmh `json:"target_speed"`
	NormalLimit valueobject.SpeedKmh `json:"normal_limit"`
	AlertLimit  valueobject.SpeedKmh `json:"alert_limit"`
	AccelRate   float64              `json:"accel_rate"` // m/s^2
	DecelRate   float64              `json:"decel_rate"` // m/s^2

	CurrentEdgeID   string `json:"current_edge_id"`
	CurrentLaneIdx  int    `json:"current_lane_idx"`
	ProgressOnLane  float64 `json:"progress_on_lane"` // meters from start of lane
	LaneLength      float64 `json:"lane_length"`

	Route        []string                 `json:"route"` // sequence of Edge IDs
	RouteIdx     int                      `json:"route_idx"`
	Waypoints    []valueobject.Coordinate `json:"waypoints"` // path points on the current lane
	WaypointIdx  int                      `json:"waypoint_idx"`

	WaitingAtRed bool `json:"waiting_at_red"`

	// Fleet Inventory & Connectivity fields (matching CSV export schema)
	Plate     string `json:"plate"`
	DeviceID  string `json:"device_id"`
	IPAddress string `json:"ip_address"`
	Port      int    `json:"port"`
	Protocol  string `json:"protocol"`
	AuthToken string `json:"auth_token"`
	IsActive  bool   `json:"is_active"`
}

// NewVehicle creates a new Vehicle with default emergency specs based on its corporation.
func NewVehicle(id, corp string, initialPos valueobject.Coordinate) *Vehicle {
	var normalLim, alertLim valueobject.SpeedKmh
	var accel, decel float64
	var plate, devID, ip, token string

	switch corp {
	case "BOMBEIRO", "BOMBEIROS":
		normalLim = 45.0
		alertLim = 70.0
		accel = 1.2
		decel = 2.5
		plate = "SPB1930"
		devID = "IMEI-864209041234569"
		ip = "10.8.0.20"
		token = "sec_token_bomb_01_xyz"
	case "SAMU":
		normalLim = 50.0
		alertLim = 80.0
		accel = 2.0
		decel = 3.5
		plate = "BRA2E19"
		devID = "IMEI-864209041234567"
		ip = "10.8.0.15"
		token = "sec_token_samu_01_xyz"
	default: // "POLICIA"
		normalLim = 50.0
		alertLim = 90.0
		accel = 2.8
		decel = 4.0
		plate = "PMM1900"
		devID = "IMEI-864209041234570"
		ip = "10.8.0.30"
		token = "sec_token_pol_01_xyz"
	}

	return &Vehicle{
		ID:             id,
		Corporation:    corp,
		Mode:           valueobject.ModeNormal,
		Position:       initialPos,
		Speed:          0.0,
		TargetSpeed:    normalLim,
		NormalLimit:    normalLim,
		AlertLimit:     alertLim,
		AccelRate:      accel,
		DecelRate:      decel,
		Route:          make([]string, 0),
		Waypoints:      make([]valueobject.Coordinate, 0),
		Plate:          plate,
		DeviceID:       devID,
		IPAddress:      ip,
		Port:           8085,
		Protocol:       "HTTP",
		AuthToken:      token,
		IsActive:       true,
	}
}

// SetMode updates the vehicle mode and recalculates target speed limit.
func (v *Vehicle) SetMode(m valueobject.VehicleMode) {
	v.Mode = m
	if m.IsAlert() {
		v.TargetSpeed = v.AlertLimit
	} else {
		v.TargetSpeed = v.NormalLimit
	}
}

// ToggleMode switches between Normal and Alert modes.
func (v *Vehicle) ToggleMode() {
	if v.Mode.IsAlert() {
		v.SetMode(valueobject.ModeNormal)
	} else {
		v.SetMode(valueobject.ModeAlert)
	}
}

// UpdateKinematics calculates speed change towards targetSpeed using acceleration limits.
func (v *Vehicle) UpdateKinematics(dt float64, desiredSpeed valueobject.SpeedKmh) {
	currMps := v.Speed.ToMps()
	targetMps := desiredSpeed.ToMps()

	if currMps < targetMps {
		currMps += v.AccelRate * dt
		if currMps > targetMps {
			currMps = targetMps
		}
	} else if currMps > targetMps {
		currMps -= v.DecelRate * dt
		if currMps < targetMps {
			currMps = targetMps
		}
	}

	if currMps < 0 {
		currMps = 0
	}
	v.Speed = valueobject.FromMps(currMps)
}

// StepAlongWaypoints advances position along waypoints by the traveled distance.
func (v *Vehicle) StepAlongWaypoints(distMeters float64) bool {
	if len(v.Waypoints) == 0 || v.WaypointIdx >= len(v.Waypoints)-1 {
		return true // reached end of current lane waypoints
	}

	remainingDist := distMeters
	for remainingDist > 0 && v.WaypointIdx < len(v.Waypoints)-1 {
		currTarget := v.Waypoints[v.WaypointIdx+1]
		distToNext := v.Position.DistanceToCartesian(currTarget)

		if distToNext <= remainingDist {
			v.Position = currTarget
			v.WaypointIdx++
			remainingDist -= distToNext
			v.ProgressOnLane += distToNext
		} else {
			// Interpolate position along segment
			fraction := remainingDist / distToNext
			dx := currTarget.X - v.Position.X
			dy := currTarget.Y - v.Position.Y
			v.Position.X += dx * fraction
			v.Position.Y += dy * fraction
			v.ProgressOnLane += remainingDist

			// GPS linear offset approximation (WGS84 approx 111,320m per deg lat)
			dLat := (currTarget.Lat - v.Position.Lat) * fraction
			dLon := (currTarget.Lon - v.Position.Lon) * fraction
			v.Position.Lat += dLat
			v.Position.Lon += dLon

			// Update heading
			v.HeadingDeg = math.Round(v.Position.BearingTo(currTarget))
			remainingDist = 0
		}
	}

	return v.WaypointIdx >= len(v.Waypoints)-1
}
