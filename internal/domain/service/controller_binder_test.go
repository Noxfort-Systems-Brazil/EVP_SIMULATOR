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

// File: controller_binder_test.go
// Author: Gabriel Moraes
// Date: October 2026

package service

import (
	"testing"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/port"
)

func TestControllerBinder_DirectMatch(t *testing.T) {
	binder := NewControllerBinder()

	lights := map[string]*entity.TrafficLight{
		"1001": entity.NewTrafficLight("1001", nil),
		"1002": entity.NewTrafficLight("1002", nil),
	}

	mappings := []port.ControllerMapping{
		{IntersectionID: "1001", IPAddress: "192.168.1.10"},
		{IntersectionID: "1002", IPAddress: "192.168.1.20"},
	}

	count := binder.BindControllers(lights, mappings)
	if count != 2 {
		t.Fatalf("expected 2 mapped, got %d", count)
	}

	if lights["1001"].IPAddress != "192.168.1.10" {
		t.Errorf("expected IP 192.168.1.10, got %s", lights["1001"].IPAddress)
	}
	if lights["1002"].IPAddress != "192.168.1.20" {
		t.Errorf("expected IP 192.168.1.20, got %s", lights["1002"].IPAddress)
	}
}

func TestControllerBinder_FallbackDeterministic(t *testing.T) {
	binder := NewControllerBinder()

	lights := map[string]*entity.TrafficLight{
		"J_NORTH": entity.NewTrafficLight("J_NORTH", nil),
		"J_SOUTH": entity.NewTrafficLight("J_SOUTH", nil),
	}

	mappings := []port.ControllerMapping{
		{IntersectionID: "2001", IPAddress: "10.0.0.1"},
		{IntersectionID: "2002", IPAddress: "10.0.0.2"},
	}

	count := binder.BindControllers(lights, mappings)
	if count != 2 {
		t.Fatalf("expected 2 mapped via fallback, got %d", count)
	}

	// Because keys are sorted deterministically: ["J_NORTH", "J_SOUTH"]
	if lights["J_NORTH"].IntersectionID != "2001" || lights["J_NORTH"].IPAddress != "10.0.0.1" {
		t.Errorf("J_NORTH not correctly bound: %+v", lights["J_NORTH"])
	}
	if lights["J_SOUTH"].IntersectionID != "2002" || lights["J_SOUTH"].IPAddress != "10.0.0.2" {
		t.Errorf("J_SOUTH not correctly bound: %+v", lights["J_SOUTH"])
	}
}
