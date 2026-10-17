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

// File: coordinate.go
// Author: Gabriel Moraes
// Date: October 2026

package valueobject

import (
	"math"
)

// Coordinate represents a spatial position in both Cartesian (meters) and Geodetic (WGS84 GPS).
type Coordinate struct {
	X   float64 `json:"x"`
	Y   float64 `json:"y"`
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// DistanceToCartesian returns the Euclidean distance in meters between two Cartesian points.
func (c Coordinate) DistanceToCartesian(target Coordinate) float64 {
	dx := target.X - c.X
	dy := target.Y - c.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// BearingTo calculates the heading/angle in degrees (0 to 360, where 0 is North, 90 is East) from c to target.
func (c Coordinate) BearingTo(target Coordinate) float64 {
	dx := target.X - c.X
	dy := target.Y - c.Y

	// In Cartesian systems: 0 deg North -> dy > 0, dx = 0.
	rad := math.Atan2(dx, dy)
	deg := rad * (180.0 / math.Pi)
	if deg < 0 {
		deg += 360.0
	}
	return deg
}
