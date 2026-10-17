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

// File: speed.go
// Author: Gabriel Moraes
// Date: October 2026

package valueobject

// SpeedKmh represents speed in kilometers per hour.
type SpeedKmh float64

// ToMps converts km/h to meters per second.
func (s SpeedKmh) ToMps() float64 {
	return float64(s) / 3.6
}

// FromMps converts meters per second to km/h.
func FromMps(mps float64) SpeedKmh {
	return SpeedKmh(mps * 3.6)
}

// Float64 returns the raw float64 value.
func (s SpeedKmh) Float64() float64 {
	return float64(s)
}
