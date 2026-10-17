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

// File: simulation_usecase_test.go
// Author: Gabriel Moraes
// Date: October 2026

package usecase

import (
	"context"
	"testing"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/port"
	"evp_simulator/internal/domain/valueobject"
)

type mockNetworkRepo struct {
	net    *entity.RoadNetwork
	lights map[string]*entity.TrafficLight
	err    error
}

func (m *mockNetworkRepo) LoadNetworkFromSUMO(filePath string) (*entity.RoadNetwork, map[string]*entity.TrafficLight, error) {
	return m.net, m.lights, m.err
}

type mockControllerRepo struct {
	mappings []port.ControllerMapping
	err      error
}

func (m *mockControllerRepo) LoadControllersFromCSV(filePath string) ([]port.ControllerMapping, error) {
	return m.mappings, m.err
}

type mockTelemetryExporter struct {
	active bool
	count  int
}

func (m *mockTelemetryExporter) ExportVehicle(ctx context.Context, veh *entity.Vehicle) error {
	m.count++
	return nil
}
func (m *mockTelemetryExporter) IsActive() bool          { return m.active }
func (m *mockTelemetryExporter) SetActive(active bool)   { m.active = active }
func (m *mockTelemetryExporter) SetTargetURL(url string) {}
func (m *mockTelemetryExporter) GetTargetURL() string    { return "http://mock/telemetry" }
func (m *mockTelemetryExporter) GetDeliveryStatus() (bool, string) {
	return true, "http://mock/telemetry"
}

type mockFleetExporter struct {
	exported []*entity.Vehicle
	err      error
}

func (m *mockFleetExporter) ExportFleetToCSV(filePath string, vehicles []*entity.Vehicle) error {
	m.exported = vehicles
	return m.err
}

type mockControllerExporter struct {
	exported []*entity.TrafficLight
	err      error
}

func (m *mockControllerExporter) ExportControllersToCSV(filePath string, trafficLights []*entity.TrafficLight) error {
	m.exported = trafficLights
	return m.err
}

type mockServerManager struct {
	startedCount int
	stopped      bool
}

func (m *mockServerManager) StartServers(lights map[string]*entity.TrafficLight, logger func(string)) error {
	m.startedCount = len(lights)
	return nil
}

func (m *mockServerManager) StopServers() {
	m.stopped = true
}

func (m *mockServerManager) ActiveServersCount() int {
	return m.startedCount
}


func createTestNetwork() (*entity.RoadNetwork, map[string]*entity.TrafficLight) {
	net := entity.NewRoadNetwork()
	net.Nodes["J1"] = &entity.Junction{ID: "J1", Position: valueobject.Coordinate{X: 100, Y: 100}}

	edge := &entity.Edge{
		ID:       "E1",
		FromNode: "J0",
		ToNode:   "J1",
		Length:   100.0,
		Speed:    50.0,
		Lanes: []*entity.Lane{
			{
				ID:     "E1_0",
				Index:  0,
				Length: 100.0,
				Shape: []valueobject.Coordinate{
					{X: 0, Y: 0, Lat: -23.55, Lon: -46.63},
					{X: 100, Y: 100, Lat: -23.54, Lon: -46.62},
				},
			},
		},
	}
	net.Edges["E1"] = edge

	lights := map[string]*entity.TrafficLight{
		"J1": entity.NewTrafficLight("J1", []entity.Phase{{Duration: 10, State: "G"}}),
	}

	return net, lights
}

func TestSimulationUseCase_LifecycleAndMap(t *testing.T) {
	net, lights := createTestNetwork()
	netRepo := &mockNetworkRepo{net: net, lights: lights}
	ctrlRepo := &mockControllerRepo{
		mappings: []port.ControllerMapping{
			{IntersectionID: "1001", IPAddress: "192.168.1.10"},
		},
	}
	telExp := &mockTelemetryExporter{}
	fleetExp := &mockFleetExporter{}
	ctrlExp := &mockControllerExporter{}
	serverMgr := &mockServerManager{}

	uc := NewSimulationUseCase(netRepo, ctrlRepo, ctrlExp, serverMgr, telExp, fleetExp)

	// 1. Initial State
	state := uc.GetState()
	if state.MapLoaded {
		t.Fatalf("expected map not loaded initially")
	}
	if state.IsRunning {
		t.Fatalf("expected simulation paused initially")
	}

	// 2. Load Map - invalid extension should fail
	if err := uc.LoadMap("dummy.txt"); err == nil {
		t.Fatalf("expected error loading map without valid extension (.net.xml.gz or .net.xml)")
	}

	// 2a. TogglePlay before map loaded should return false
	if running := uc.TogglePlay(); running {
		t.Fatalf("expected TogglePlay to return false before map is loaded")
	}

	// 2b. Load Map - valid .net.xml.gz should succeed and immediately start servers
	if err := uc.LoadMap("dummy.net.xml.gz"); err != nil {
		t.Fatalf("LoadMap failed: %v", err)
	}

	if serverMgr.startedCount != 1 {
		t.Fatalf("expected 1 server started on map load, got %d", serverMgr.startedCount)
	}


	state = uc.GetState()
	if !state.MapLoaded {
		t.Fatalf("expected map to be loaded")
	}
	if !state.IsRunning {
		t.Fatalf("expected simulation to be running automatically on map load")
	}
	if len(state.Vehicles) != 3 {
		t.Fatalf("expected 3 vehicles, got %d", len(state.Vehicles))
	}
	if state.TotalLights != 1 {
		t.Fatalf("expected 1 traffic light, got %d", state.TotalLights)
	}

	// 3. Import Controllers
	count, err := uc.ImportControllers("dummy.csv")
	if err != nil {
		t.Fatalf("ImportControllers failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 controller mapped, got %d", count)
	}

	// 4. Toggle Play (pauses the automatically running simulation)
	running := uc.TogglePlay()
	if running {
		t.Fatalf("expected simulation to be paused after toggle")
	}
	// Resume simulation for subsequent steps
	if running = uc.TogglePlay(); !running {
		t.Fatalf("expected simulation to be running after resuming")
	}

	// 5. Verify Initial Alert Mode and Toggle Vehicle Mode
	firstVehID := state.Vehicles[0].ID
	if state.Vehicles[0].Mode != "ALERT" {
		t.Fatalf("expected initial vehicle mode ALERT, got %s", state.Vehicles[0].Mode)
	}

	// Toggle vehicle to Normal
	if err := uc.ToggleVehicleMode(firstVehID); err != nil {
		t.Fatalf("ToggleVehicleMode failed: %v", err)
	}

	state = uc.GetState()
	var updatedVehMode string
	for _, v := range state.Vehicles {
		if v.ID == firstVehID {
			updatedVehMode = v.Mode
			break
		}
	}
	if updatedVehMode != "NORMAL" {
		t.Fatalf("expected vehicle mode NORMAL after toggle, got %s", updatedVehMode)
	}

	// Toggle vehicle back to Alert
	if err := uc.ToggleVehicleMode(firstVehID); err != nil {
		t.Fatalf("ToggleVehicleMode back failed: %v", err)
	}
	state = uc.GetState()
	for _, v := range state.Vehicles {
		if v.ID == firstVehID {
			updatedVehMode = v.Mode
			break
		}
	}
	if updatedVehMode != "ALERT" {
		t.Fatalf("expected vehicle mode ALERT after second toggle, got %s", updatedVehMode)
	}

	// 6. Tick physics
	uc.Tick(0.1)

	// 7. Telemetry activation
	exporting := uc.SetExportActive(true)
	if !exporting {
		t.Fatalf("expected telemetry active")
	}

	// 8. Export fleet CSV
	if err := uc.ExportFleetCSV("fleet.csv"); err != nil {
		t.Fatalf("ExportFleetCSV failed: %v", err)
	}
	if len(fleetExp.exported) != 3 {
		t.Fatalf("expected 3 vehicles exported, got %d", len(fleetExp.exported))
	}

	// 8b. Export controllers CSV
	if err := uc.ExportControllersCSV("controllers.csv"); err != nil {
		t.Fatalf("ExportControllersCSV failed: %v", err)
	}
	if len(ctrlExp.exported) != 1 {
		t.Fatalf("expected 1 controller exported, got %d", len(ctrlExp.exported))
	}
	if ctrlExp.exported[0].IntersectionID != "1001" {
		t.Errorf("expected IntersectionID '1001', got '%s'", ctrlExp.exported[0].IntersectionID)
	}

	// 9. Logs
	if len(uc.logger.GetLogs()) == 0 {
		t.Fatalf("expected logs to be recorded")
	}
}

func TestSimulationUseCase_ToggleInvalidVehicle(t *testing.T) {
	uc := NewSimulationUseCase(&mockNetworkRepo{}, &mockControllerRepo{}, &mockControllerExporter{}, &mockServerManager{}, &mockTelemetryExporter{}, &mockFleetExporter{})
	err := uc.ToggleVehicleMode("NON_EXISTENT")
	if err == nil {
		t.Fatalf("expected error toggling non-existent vehicle")
	}
}

func TestSimulationUseCase_ToggleAllVehiclesMode(t *testing.T) {
	net, lights := createTestNetwork()
	netRepo := &mockNetworkRepo{net: net, lights: lights}
	ctrlRepo := &mockControllerRepo{}
	ctrlExp := &mockControllerExporter{}
	serverMgr := &mockServerManager{}
	telExp := &mockTelemetryExporter{}
	fleetExp := &mockFleetExporter{}

	uc := NewSimulationUseCase(netRepo, ctrlRepo, ctrlExp, serverMgr, telExp, fleetExp)
	if err := uc.LoadMap("test.net.xml.gz"); err != nil {
		t.Fatalf("LoadMap failed: %v", err)
	}

	state := uc.GetState()
	if len(state.Vehicles) == 0 {
		t.Fatalf("expected spawned vehicles")
	}

	// 1. Deactivate alert for all vehicles
	initialExportCount := telExp.count
	modified := uc.ToggleAllVehiclesMode(false)
	if modified != len(state.Vehicles) {
		t.Errorf("expected %d modified vehicles, got %d", len(state.Vehicles), modified)
	}

	state = uc.GetState()
	for _, v := range state.Vehicles {
		if v.Mode != "NORMAL" {
			t.Errorf("expected vehicle %s to be in NORMAL mode, got %s", v.ID, v.Mode)
		}
	}
	// Verify immediate export occurred for all vehicles
	if telExp.count != initialExportCount+len(state.Vehicles) {
		t.Errorf("expected %d immediate exports, got %d", initialExportCount+len(state.Vehicles), telExp.count)
	}

	// 2. Reactivate alert for all vehicles
	beforeReactivate := telExp.count
	modified = uc.ToggleAllVehiclesMode(true)
	if modified != len(state.Vehicles) {
		t.Errorf("expected %d modified vehicles on reactivate, got %d", len(state.Vehicles), modified)
	}

	state = uc.GetState()
	for _, v := range state.Vehicles {
		if v.Mode != "ALERT" {
			t.Errorf("expected vehicle %s to be in ALERT mode, got %s", v.ID, v.Mode)
		}
	}
	if telExp.count != beforeReactivate+len(state.Vehicles) {
		t.Errorf("expected %d immediate exports on reactivate, got %d", beforeReactivate+len(state.Vehicles), telExp.count)
	}
}

func TestSimulationUseCase_ImmediateExportOnSingleToggle(t *testing.T) {
	net, lights := createTestNetwork()
	netRepo := &mockNetworkRepo{net: net, lights: lights}
	ctrlRepo := &mockControllerRepo{}
	ctrlExp := &mockControllerExporter{}
	serverMgr := &mockServerManager{}
	telExp := &mockTelemetryExporter{}
	fleetExp := &mockFleetExporter{}

	uc := NewSimulationUseCase(netRepo, ctrlRepo, ctrlExp, serverMgr, telExp, fleetExp)
	if err := uc.LoadMap("test.net.xml.gz"); err != nil {
		t.Fatalf("LoadMap failed: %v", err)
	}

	// Pause simulation to prove export happens independently of running state/Tick
	uc.state.SetRunning(false)

	state := uc.GetState()
	firstVehID := state.Vehicles[0].ID
	beforeCount := telExp.count

	if err := uc.ToggleVehicleMode(firstVehID); err != nil {
		t.Fatalf("ToggleVehicleMode failed: %v", err)
	}

	if telExp.count != beforeCount+1 {
		t.Errorf("expected immediate export count %d, got %d", beforeCount+1, telExp.count)
	}
}

func TestStateProjector_DeterministicOrder(t *testing.T) {
	projector := NewStateProjector()

	vehicles := map[string]*entity.Vehicle{
		"V-Z": {ID: "V-Z", Corporation: "POLICIA"},
		"V-A": {ID: "V-A", Corporation: "SAMU"},
		"V-M": {ID: "V-M", Corporation: "BOMBEIRO"},
	}

	lights := map[string]*entity.TrafficLight{
		"TL-3": {ID: "TL-3"},
		"TL-1": {ID: "TL-1"},
		"TL-2": {ID: "TL-2"},
	}

	for i := 0; i < 20; i++ {
		dto := projector.Project(nil, lights, vehicles, true, true, nil, "")
		if len(dto.Vehicles) != 3 {
			t.Fatalf("expected 3 vehicles")
		}
		if dto.Vehicles[0].ID != "V-A" || dto.Vehicles[1].ID != "V-M" || dto.Vehicles[2].ID != "V-Z" {
			t.Fatalf("iteration %d: non-deterministic vehicle order: %v, %v, %v",
				i, dto.Vehicles[0].ID, dto.Vehicles[1].ID, dto.Vehicles[2].ID)
		}
		if dto.TrafficLights[0].ID != "TL-1" || dto.TrafficLights[1].ID != "TL-2" || dto.TrafficLights[2].ID != "TL-3" {
			t.Fatalf("iteration %d: non-deterministic lights order", i)
		}
	}
}

