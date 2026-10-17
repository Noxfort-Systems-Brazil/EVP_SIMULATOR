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

// File: main.go
// Author: Gabriel Moraes
// Date: October 2026

package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"evp_simulator/internal/application/usecase"
	"evp_simulator/internal/infrastructure/csv"
	"evp_simulator/internal/infrastructure/httpclient"
	"evp_simulator/internal/infrastructure/sumo"
	"evp_simulator/internal/infrastructure/utmc"
	wailsbridge "evp_simulator/internal/infrastructure/wails"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	// Initialize Infrastructure Adapters
	netRepo := sumo.NewParser()
	ctrlRepo := csv.NewControllerImporter()
	ctrlExp := csv.NewControllerCSVExporter()
	fleetExporter := csv.NewVehicleCSVExporter()
	serverMgr := utmc.NewServerManager()
	telemetryExporter := httpclient.NewCarinaExporter(httpclient.DefaultLocalURL)

	// Initialize Application UseCase
	simUseCase := usecase.NewSimulationUseCase(netRepo, ctrlRepo, ctrlExp, serverMgr, telemetryExporter, fleetExporter)

	// Initialize Wails App Bridge
	app := wailsbridge.NewApp(simUseCase)

	// Create and Run Wails Desktop Application
	err := wails.Run(&options.App{
		Title:  "EVP Fleet Simulator - CARINA_CORE",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 255},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error starting Wails application:", err.Error())
	}
}
