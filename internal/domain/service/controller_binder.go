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

// File: controller_binder.go
// Author: Gabriel Moraes
// Date: October 2026

package service

import (
	"sort"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/port"
)

// ControllerBinder encapsulates the domain logic of binding imported controller network configurations
// to existing traffic light entities in the road network.
type ControllerBinder struct{}

// NewControllerBinder creates a new ControllerBinder instance.
func NewControllerBinder() *ControllerBinder {
	return &ControllerBinder{}
}

// BindControllers maps external controller inventory entries to traffic lights.
// It prioritizes direct ID matching, with a deterministic round-robin fallback when IDs differ (e.g., SUMO junction IDs vs numeric IDs).
func (cb *ControllerBinder) BindControllers(
	trafficLights map[string]*entity.TrafficLight,
	mappings []port.ControllerMapping,
) int {
	if len(trafficLights) == 0 || len(mappings) == 0 {
		return 0
	}

	mappedCount := 0

	// 1. Direct ID match (SUMO ID == CSV IntersectionID)
	for _, m := range mappings {
		if tl, exists := trafficLights[m.IntersectionID]; exists {
			tl.IntersectionID = m.IntersectionID
			tl.IPAddress = m.IPAddress
			mappedCount++
		}
	}

	// 2. Deterministic fallback if no direct ID matched
	if mappedCount == 0 {
		// Sort keys to guarantee deterministic assignment across runs
		keys := make([]string, 0, len(trafficLights))
		for k := range trafficLights {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, m := range mappings {
			tlKey := keys[i%len(keys)]
			targetTL := trafficLights[tlKey]
			targetTL.IntersectionID = m.IntersectionID
			targetTL.IPAddress = m.IPAddress
			mappedCount++
		}
	}

	return mappedCount
}
