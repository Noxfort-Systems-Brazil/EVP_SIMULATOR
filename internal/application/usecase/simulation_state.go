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

// File: simulation_state.go
// Author: Gabriel Moraes
// Date: October 2026

package usecase

import (
	"sync"

	"evp_simulator/internal/domain/entity"
)

// SimulationState holds the in-memory domain state and synchronizes concurrent access.
type SimulationState struct {
	mu            sync.RWMutex
	network       *entity.RoadNetwork
	trafficLights map[string]*entity.TrafficLight
	vehicles      map[string]*entity.Vehicle
	isRunning     bool
}

// NewSimulationState creates an empty simulation state container.
func NewSimulationState() *SimulationState {
	return &SimulationState{
		trafficLights: make(map[string]*entity.TrafficLight),
		vehicles:      make(map[string]*entity.Vehicle),
		isRunning:     false,
	}
}

// Reset initializes the road network, traffic lights, and vehicles.
func (s *SimulationState) Reset(
	net *entity.RoadNetwork,
	lights map[string]*entity.TrafficLight,
	vehicles map[string]*entity.Vehicle,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.network = net
	s.trafficLights = lights
	s.vehicles = vehicles
}

// ToggleRunning inverts the running status and returns the new state.
func (s *SimulationState) ToggleRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRunning = !s.isRunning
	return s.isRunning
}

// SetRunning directly updates the running state.
func (s *SimulationState) SetRunning(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRunning = running
}

// IsRunning returns whether the simulation loop is active.
func (s *SimulationState) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// GetVehicle retrieves a vehicle by ID.
func (s *SimulationState) GetVehicle(id string) (*entity.Vehicle, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, exists := s.vehicles[id]
	return v, exists
}

// GetVehiclesList returns a defensive copy slice of all currently instantiated vehicles.
func (s *SimulationState) GetVehiclesList() []*entity.Vehicle {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*entity.Vehicle, 0, len(s.vehicles))
	for _, v := range s.vehicles {
		list = append(list, v)
	}
	return list
}

// GetTrafficLightsList returns a defensive copy slice of all currently instantiated traffic lights.
func (s *SimulationState) GetTrafficLightsList() []*entity.TrafficLight {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*entity.TrafficLight, 0, len(s.trafficLights))
	for _, tl := range s.trafficLights {
		list = append(list, tl)
	}
	return list
}

// WithLock executes an action holding an exclusive write lock.
func (s *SimulationState) WithLock(fn func(net *entity.RoadNetwork, lights map[string]*entity.TrafficLight, vehicles map[string]*entity.Vehicle, isRunning bool)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s.network, s.trafficLights, s.vehicles, s.isRunning)
}

// WithRLock executes an action holding a shared read lock.
func (s *SimulationState) WithRLock(fn func(net *entity.RoadNetwork, lights map[string]*entity.TrafficLight, vehicles map[string]*entity.Vehicle, isRunning bool)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fn(s.network, s.trafficLights, s.vehicles, s.isRunning)
}
