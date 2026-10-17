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

// File: simulation_usecase.go
// Author: Gabriel Moraes
// Date: October 2026

package usecase

import (
	"fmt"
	"strings"
	"time"

	"evp_simulator/internal/application/dto"
	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/port"
	"evp_simulator/internal/domain/service"
	"evp_simulator/internal/domain/valueobject"
)

// SimulationUseCase orchestrates the simulation lifecycle, vehicles, traffic lights, and external exports.
// It serves as the Clean Architecture Application Coordinator / Facade.
type SimulationUseCase struct {
	networkRepo        port.NetworkRepository
	controllerRepo     port.ControllerRepository
	controllerExporter port.ControllerExporter
	serverManager      port.ControllerServerManager
	fleetExporter      port.FleetExporter

	router           *service.AStarRouter
	kinematicsEng    *service.KinematicsEngine
	fleetSpawner     *service.FleetSpawner
	controllerBinder *service.ControllerBinder

	state          *SimulationState
	logger         *SimulationLogger
	telemetryDisp  *TelemetryDispatcher
	stateProjector *StateProjector
}

// NewSimulationUseCase instantiates the main application coordinator with default domain and application services.
func NewSimulationUseCase(
	netRepo port.NetworkRepository,
	ctrlRepo port.ControllerRepository,
	ctrlExp port.ControllerExporter,
	serverMgr port.ControllerServerManager,
	telExp port.TelemetryExporter,
	fleetExp port.FleetExporter,
) *SimulationUseCase {
	router := service.NewAStarRouter()
	evaluator := service.NewSignalEvaluator()
	kinematics := service.NewKinematicsEngine(router, evaluator)
	fleetSpawner := service.NewFleetSpawner(router)
	controllerBinder := service.NewControllerBinder()

	return NewSimulationUseCaseWithServices(
		netRepo,
		ctrlRepo,
		ctrlExp,
		serverMgr,
		telExp,
		fleetExp,
		router,
		kinematics,
		fleetSpawner,
		controllerBinder,
		NewSimulationState(),
		NewSimulationLogger(300),
		NewTelemetryDispatcher(telExp, time.Second),
		NewStateProjector(),
	)
}

// NewSimulationUseCaseWithServices allows full dependency injection for testing and extensibility (DIP).
func NewSimulationUseCaseWithServices(
	netRepo port.NetworkRepository,
	ctrlRepo port.ControllerRepository,
	ctrlExp port.ControllerExporter,
	serverMgr port.ControllerServerManager,
	telExp port.TelemetryExporter,
	fleetExp port.FleetExporter,
	router *service.AStarRouter,
	kinematics *service.KinematicsEngine,
	fleetSpawner *service.FleetSpawner,
	controllerBinder *service.ControllerBinder,
	state *SimulationState,
	logger *SimulationLogger,
	telemetryDisp *TelemetryDispatcher,
	stateProjector *StateProjector,
) *SimulationUseCase {
	if telemetryDisp != nil && logger != nil {
		telemetryDisp.SetLogger(logger)
	}

	return &SimulationUseCase{
		networkRepo:        netRepo,
		controllerRepo:     ctrlRepo,
		controllerExporter: ctrlExp,
		serverManager:      serverMgr,
		fleetExporter:      fleetExp,
		router:             router,
		kinematicsEng:      kinematics,
		fleetSpawner:       fleetSpawner,
		controllerBinder:   controllerBinder,
		state:              state,
		logger:             logger,
		telemetryDisp:      telemetryDisp,
		stateProjector:     stateProjector,
	}
}

// AddLog logs an event with timestamp for UI consumption.
func (uc *SimulationUseCase) AddLog(msg string) {
	uc.logger.AddLog(msg)
}

// StopServers terminates all running UTMC/SNMP UDP servers.
func (uc *SimulationUseCase) StopServers() {
	if uc.serverManager != nil {
		uc.serverManager.StopServers()
	}
}

// LoadMap parses a SUMO network compressed file (.net.xml.gz), initializes the road topology, and deploys the initial fleet.
func (uc *SimulationUseCase) LoadMap(filePath string) error {
	lower := strings.ToLower(filePath)
	if !strings.HasSuffix(lower, ".net.xml.gz") && !strings.HasSuffix(lower, ".net.xml") {
		err := fmt.Errorf("formato de mapa não suportado: somente arquivos .net.xml.gz ou .net.xml são aceitos")
		uc.AddLog(fmt.Sprintf("❌ Erro ao carregar mapa: %v", err))
		return err
	}

	net, lights, err := uc.networkRepo.LoadNetworkFromSUMO(filePath)
	if err != nil {
		uc.AddLog(fmt.Sprintf("❌ Erro ao processar mapa SUMO: %v", err))
		return err
	}

	vehicles := uc.fleetSpawner.SpawnEmergencyFleet(net)
	uc.state.Reset(net, lights, vehicles)
	uc.state.SetRunning(true)

	// Start live UTMC/SNMP servers immediately upon map load so CARINA can connect
	if uc.serverManager != nil {
		if err := uc.serverManager.StartServers(lights, uc.AddLog); err != nil {
			uc.AddLog(fmt.Sprintf("⚠️ Erro ao iniciar servidores UTMC: %v", err))
		}
	}

	uc.AddLog(fmt.Sprintf("✅ Malha viária importada com sucesso: %d vias, %d semáforos, %d viaturas conectadas.",
		len(net.Edges), len(lights), len(vehicles)))
	uc.AddLog("🚨 Modo Alerta ativado automaticamente: viaturas iniciadas com sirenes ativas, rota prioritária e teto de emergência.")
	uc.AddLog("🚀 Simulação iniciada automaticamente: viaturas em movimento (10 Hz) e telemetria ativa.")
	return nil
}

// ImportControllers loads controller IP addresses and Intersection IDs from CSV and binds them to traffic lights.
func (uc *SimulationUseCase) ImportControllers(filePath string) (int, error) {
	mappings, err := uc.controllerRepo.LoadControllersFromCSV(filePath)
	if err != nil {
		uc.AddLog(fmt.Sprintf("Erro ao importar CSV: %v", err))
		return 0, err
	}

	mappedCount := 0
	uc.state.WithLock(func(net *entity.RoadNetwork, lights map[string]*entity.TrafficLight, vehicles map[string]*entity.Vehicle, isRunning bool) {
		mappedCount = uc.controllerBinder.BindControllers(lights, mappings)
	})

	uc.AddLog(fmt.Sprintf("Importados %d controladores mapeados com os semáforos da malha.", mappedCount))
	return mappedCount, nil
}

// TogglePlay toggles the simulation running state.
func (uc *SimulationUseCase) TogglePlay() bool {
	var hasNetwork bool
	uc.state.WithRLock(func(net *entity.RoadNetwork, lights map[string]*entity.TrafficLight, vehicles map[string]*entity.Vehicle, isRunning bool) {
		hasNetwork = net != nil && len(net.Edges) > 0
	})

	if !hasNetwork {
		uc.AddLog("⚠️ Simulação não pode ser iniciada: nenhuma malha viária importada. Clique em 'Mapa' para importar a malha (.net.xml.gz) e instanciar as viaturas.")
		return false
	}


	running := uc.state.ToggleRunning()
	if running {
		if uc.telemetryDisp.IsActive() {
			uc.AddLog("Simulação INICIADA (Play) • Telemetria ATIVADA (1 Hz para CARINA_CORE)")
		} else {
			uc.AddLog("Simulação INICIADA (Play)")
		}
	} else {
		uc.AddLog("Simulação PAUSADA")
	}
	return running
}


// SetExportActive enforces mandatory continuous telemetry streaming to CARINA_CORE.
func (uc *SimulationUseCase) SetExportActive(_ bool) bool {
	uc.telemetryDisp.SetActive(true)
	return true
}

// ExportFleetCSV exports the simulated fleet inventory to CSV matching CARINA's expected schema.
func (uc *SimulationUseCase) ExportFleetCSV(filePath string) error {
	vList := uc.state.GetVehiclesList()
	if len(vList) == 0 {
		err := fmt.Errorf("nenhuma viatura para exportar: carregue a malha viária SUMO (.net.xml.gz) primeiro clicando em 'Mapa'")
		uc.AddLog(fmt.Sprintf("⚠️ %v", err))
		return err
	}

	if !strings.HasSuffix(strings.ToLower(filePath), ".csv") {
		filePath += ".csv"
	}

	if err := uc.fleetExporter.ExportFleetToCSV(filePath, vList); err != nil {
		uc.AddLog(fmt.Sprintf("❌ Erro ao exportar frota CSV: %v", err))
		return err
	}

	uc.AddLog(fmt.Sprintf("✅ Frota de %d viaturas exportada com sucesso em: %s", len(vList), filePath))
	return nil
}

// ExportControllersCSV exports the traffic lights inventory to CSV matching CARINA's expected schema (Intersection ID, IP Address).
func (uc *SimulationUseCase) ExportControllersCSV(filePath string) error {
	tList := uc.state.GetTrafficLightsList()
	if len(tList) == 0 {
		err := fmt.Errorf("nenhum semáforo para exportar: carregue a malha viária SUMO (.net.xml.gz) primeiro clicando em 'Mapa'")
		uc.AddLog(fmt.Sprintf("⚠️ %v", err))
		return err
	}

	if !strings.HasSuffix(strings.ToLower(filePath), ".csv") {
		filePath += ".csv"
	}

	if err := uc.controllerExporter.ExportControllersToCSV(filePath, tList); err != nil {
		uc.AddLog(fmt.Sprintf("❌ Erro ao exportar controladores CSV: %v", err))
		return err
	}

	uc.AddLog(fmt.Sprintf("✅ Inventário de %d controladores de semáforos exportado com sucesso em: %s", len(tList), filePath))
	return nil
}

// ToggleVehicleMode toggles between Normal and Alert mode for a vehicle.
func (uc *SimulationUseCase) ToggleVehicleMode(vehicleID string) error {
	var targetVehicle *entity.Vehicle

	uc.state.WithLock(func(net *entity.RoadNetwork, lights map[string]*entity.TrafficLight, vehicles map[string]*entity.Vehicle, isRunning bool) {
		v, exists := vehicles[vehicleID]
		if !exists {
			return
		}

		v.ToggleMode()
		// Recalculate route if switched to Alert to leverage emergency prioritization
		if v.Mode.IsAlert() && len(v.Route) > 0 && net != nil {
			targetEdge := v.Route[len(v.Route)-1]
			v.Route = uc.router.FindRoute(net, v.CurrentEdgeID, targetEdge, v.Mode)
			v.RouteIdx = 0
		}
		targetVehicle = v
	})

	if targetVehicle == nil {
		return fmt.Errorf("viatura %s não encontrada", vehicleID)
	}

	if targetVehicle.Mode.IsAlert() {
		uc.AddLog(fmt.Sprintf("🚨 %s entrou em MODO ALERTA! Sirene ativada, rota de emergência calculada.", targetVehicle.ID))
	} else {
		uc.AddLog(fmt.Sprintf("ℹ️ %s retornou ao MODO NORMAL.", targetVehicle.ID))
	}

	// Dispatch immediate telemetry to CARINA so state changes take effect with zero delay
	if uc.telemetryDisp != nil {
		uc.telemetryDisp.ExportVehicleNow(targetVehicle)
	}

	return nil
}

// ToggleAllVehiclesMode sets all vehicles to Alert or Normal mode simultaneously and exports immediately.
func (uc *SimulationUseCase) ToggleAllVehiclesMode(alert bool) int {
	var modified []*entity.Vehicle
	targetMode := valueobject.ModeNormal
	if alert {
		targetMode = valueobject.ModeAlert
	}

	uc.state.WithLock(func(net *entity.RoadNetwork, lights map[string]*entity.TrafficLight, vehicles map[string]*entity.Vehicle, isRunning bool) {
		for _, v := range vehicles {
			if v.Mode != targetMode {
				v.SetMode(targetMode)
				if alert && len(v.Route) > 0 && net != nil {
					targetEdge := v.Route[len(v.Route)-1]
					v.Route = uc.router.FindRoute(net, v.CurrentEdgeID, targetEdge, v.Mode)
					v.RouteIdx = 0
				}
				modified = append(modified, v)
			}
		}
	})

	if uc.telemetryDisp != nil {
		for _, v := range modified {
			uc.telemetryDisp.ExportVehicleNow(v)
		}
	}

	if alert {
		uc.AddLog(fmt.Sprintf("🚨 Modo Alerta ativado para toda a frota (%d viaturas em emergência).", len(modified)))
	} else {
		uc.AddLog(fmt.Sprintf("ℹ️ Modo Alerta desativado para toda a frota (%d viaturas em patrulha normal).", len(modified)))
	}

	return len(modified)
}

// Tick executes one simulation step for all vehicles and traffic lights, plus telemetry streaming.
func (uc *SimulationUseCase) Tick(dt float64) {
	var activeVehicles map[string]*entity.Vehicle

	uc.state.WithLock(func(net *entity.RoadNetwork, lights map[string]*entity.TrafficLight, vehicles map[string]*entity.Vehicle, isRunning bool) {
		if !isRunning || net == nil {
			return
		}

		// 1. Tick traffic lights
		for _, tl := range lights {
			tl.Tick(dt)
		}

		// 2. Tick vehicles kinematics
		for _, v := range vehicles {
			uc.kinematicsEng.UpdateVehicle(v, net, lights, dt)
		}

		activeVehicles = vehicles
	})

	// 3. Dispatch telemetry export outside the physics write lock
	if activeVehicles != nil {
		uc.telemetryDisp.Dispatch(activeVehicles)
	}
}

// GetTargetURL returns the currently active telemetry target endpoint.
func (uc *SimulationUseCase) GetTargetURL() string {
	if uc.telemetryDisp != nil {
		return uc.telemetryDisp.GetTargetURL()
	}
	return ""
}

// SetTargetURL reconfigures the telemetry destination URL.
func (uc *SimulationUseCase) SetTargetURL(url string) {
	if uc.telemetryDisp != nil {
		uc.telemetryDisp.SetTargetURL(url)
		uc.AddLog(fmt.Sprintf("📡 Destino de telemetria atualizado para: %s", url))
	}
}

// GetState snapshots the current simulation state for UI synchronization.
func (uc *SimulationUseCase) GetState() *dto.SimulationStateDTO {
	var stateSnapshot *dto.SimulationStateDTO

	targetURL := ""
	if uc.telemetryDisp != nil {
		targetURL = uc.telemetryDisp.GetTargetURL()
	}

	uc.state.WithRLock(func(net *entity.RoadNetwork, lights map[string]*entity.TrafficLight, vehicles map[string]*entity.Vehicle, isRunning bool) {
		stateSnapshot = uc.stateProjector.Project(
			net,
			lights,
			vehicles,
			isRunning,
			uc.telemetryDisp.IsActive(),
			uc.logger.GetLogs(),
			targetURL,
		)
	})

	return stateSnapshot
}
