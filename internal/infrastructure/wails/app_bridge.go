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

// File: app_bridge.go
// Author: Gabriel Moraes
// Date: October 2026

package wails

import (
	"context"
	"fmt"
	"strings"
	"time"


	"github.com/wailsapp/wails/v2/pkg/runtime"

	"evp_simulator/internal/application/dto"
	"evp_simulator/internal/application/usecase"
)

// App is the Wails bridge connecting the Go backend with the UI frontend.
type App struct {
	ctx      context.Context
	useCase  *usecase.SimulationUseCase
	stopChan chan struct{}
}

// NewApp creates a new App instance.
func NewApp(uc *usecase.SimulationUseCase) *App {
	return &App{
		useCase:  uc,
		stopChan: make(chan struct{}),
	}
}

// Startup is called when the Wails application starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Enable telemetry by default (1 Hz continuous stream to CARINA_CORE)
	a.useCase.SetExportActive(true)
	target := a.useCase.GetTargetURL()
	a.useCase.AddLog(fmt.Sprintf("📡 Telemetria ATIVADA (stream contínuo 1 Hz -> %s). Use 'Conectar CARINA' para alterar o endpoint.", target))
	a.useCase.AddLog("🗺️ Aguardando importação da malha viária: clique em 'Importar Mapa' para carregar a malha SUMO (.net.xml.gz) e iniciar a simulação automaticamente.")

	// Start 10 Hz physics simulation loop
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-a.stopChan:
				return
			case <-ticker.C:
				a.useCase.Tick(0.1) // 100ms dt
			}
		}
	}()
}

// Shutdown is called when the application terminates.
func (a *App) Shutdown(ctx context.Context) {
	close(a.stopChan)
	a.useCase.StopServers()
}

// OpenMapDialog opens native OS file picker for SUMO network files (.net.xml.gz or .net.xml).
func (a *App) OpenMapDialog() (string, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecione a Malha Viária do SUMO (.net.xml.gz ou .net.xml)",
		Filters: []runtime.FileFilter{
			{DisplayName: "SUMO Network (*.net.xml.gz, *.net.xml)", Pattern: "*.net.xml.gz;*.net.xml"},
		},
	})
	if err != nil || file == "" {
		return "", err
	}

	lower := strings.ToLower(file)
	if !strings.HasSuffix(lower, ".net.xml.gz") && !strings.HasSuffix(lower, ".net.xml") {
		err := fmt.Errorf("formato inválido: selecione apenas arquivos de malha compactados (.net.xml.gz) ou XML (.net.xml)")
		a.useCase.AddLog(fmt.Sprintf("❌ Erro ao abrir mapa: %v", err))
		return "", err
	}

	err = a.useCase.LoadMap(file)
	return file, err
}


// ExportControllersCSVDialog opens native OS save file picker to export the traffic light controllers CSV for CARINA_CORE.
func (a *App) ExportControllersCSVDialog() (string, error) {
	state := a.useCase.GetState()
	if !state.MapLoaded || state.TotalLights == 0 {
		err := fmt.Errorf("nenhum semáforo para exportar: carregue a malha viária SUMO (.net.xml.gz) primeiro clicando em 'Mapa'")
		a.useCase.AddLog(fmt.Sprintf("⚠️ %v", err))
		return "", err
	}

	file, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Exportar Inventário de Semáforos (CSV para CARINA)",
		DefaultFilename: "controllers_export.csv",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
		},
	})
	if err != nil || file == "" {
		return "", err
	}

	err = a.useCase.ExportControllersCSV(file)
	return file, err
}

// OpenCSVDialog opens native OS file picker for controller inventory CSV (legacy import).
func (a *App) OpenCSVDialog() (int, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecione o Inventário de Controladores (CSV)",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
		},
	})
	if err != nil || file == "" {
		return 0, err
	}

	return a.useCase.ImportControllers(file)
}

// TogglePlay starts or pauses the simulation.
func (a *App) TogglePlay() bool {
	return a.useCase.TogglePlay()
}

// ToggleExport enables or disables vehicle telemetry broadcast to CARINA_CORE.
func (a *App) ToggleExport(active bool) bool {
	return a.useCase.SetExportActive(active)
}

// ExportVehiclesCSVDialog opens native OS save file picker to export the fleet CSV for CARINA_CORE.
func (a *App) ExportVehiclesCSVDialog() (string, error) {
	state := a.useCase.GetState()
	if !state.MapLoaded || len(state.Vehicles) == 0 {
		err := fmt.Errorf("nenhuma viatura para exportar: carregue a malha viária SUMO (.net.xml.gz) primeiro clicando em 'Mapa'")
		a.useCase.AddLog(fmt.Sprintf("⚠️ %v", err))
		return "", err
	}

	file, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Exportar Inventário de Viaturas (CSV para CARINA)",
		DefaultFilename: "vehicles_export.csv",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
		},
	})
	if err != nil || file == "" {
		return "", err
	}

	err = a.useCase.ExportFleetCSV(file)
	return file, err
}

// ToggleVehicleMode toggles between Normal and Alert mode for vehicleID.
func (a *App) ToggleVehicleMode(vehicleID string) error {
	return a.useCase.ToggleVehicleMode(vehicleID)
}

// ToggleAllVehiclesMode toggles Alert mode for the entire fleet simultaneously.
func (a *App) ToggleAllVehiclesMode(alert bool) int {
	return a.useCase.ToggleAllVehiclesMode(alert)
}


// GetSimulationState returns a snapshot of the live simulation for the UI.
func (a *App) GetSimulationState() *dto.SimulationStateDTO {
	return a.useCase.GetState()
}

// SetTargetURL sets the telemetry destination URL dynamically.
func (a *App) SetTargetURL(url string) string {
	a.useCase.SetTargetURL(url)
	return a.useCase.GetTargetURL()
}

// ToggleTargetURL resets to DefaultLocalURL or returns the active target URL.
func (a *App) ToggleTargetURL() string {
	current := a.useCase.GetTargetURL()
	if current == "" {
		current = "http://localhost:2201/amb/v1"
		a.useCase.SetTargetURL(current)
	}
	return current
}
