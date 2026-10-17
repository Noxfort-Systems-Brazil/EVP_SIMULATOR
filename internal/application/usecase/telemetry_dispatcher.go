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

// File: telemetry_dispatcher.go
// Author: Gabriel Moraes
// Date: October 2026

package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/port"
)

// TelemetryDispatcher controls rate-limiting and asynchronous broadcast of vehicle telemetry to CARINA_CORE.
type TelemetryDispatcher struct {
	mu             sync.Mutex
	exporter       port.TelemetryExporter
	interval       time.Duration
	lastExportTime time.Time
	isExporting    bool
	logger         *SimulationLogger
	exportCount    int
}

// SetLogger sets the simulation logger for periodic telemetry confirmations.
func (td *TelemetryDispatcher) SetLogger(l *SimulationLogger) {
	td.mu.Lock()
	defer td.mu.Unlock()
	td.logger = l
}

// NewTelemetryDispatcher creates a new dispatcher with a specified rate limit interval.
func NewTelemetryDispatcher(exporter port.TelemetryExporter, interval time.Duration) *TelemetryDispatcher {
	if interval <= 0 {
		interval = time.Second
	}
	if exporter != nil {
		exporter.SetActive(true)
	}
	return &TelemetryDispatcher{
		exporter:       exporter,
		interval:       interval,
		lastExportTime: time.Now(),
		isExporting:    true,
	}
}

// SetActive ensures telemetry streaming remains active (mandatory for CARINA_CORE).
func (td *TelemetryDispatcher) SetActive(_ bool) bool {
	td.mu.Lock()
	defer td.mu.Unlock()

	td.isExporting = true
	if td.exporter != nil {
		td.exporter.SetActive(true)
	}
	return true
}

// IsActive returns the current streaming state.
func (td *TelemetryDispatcher) IsActive() bool {
	td.mu.Lock()
	defer td.mu.Unlock()
	return td.isExporting
}

// GetTargetURL returns the currently configured telemetry destination endpoint.
func (td *TelemetryDispatcher) GetTargetURL() string {
	td.mu.Lock()
	defer td.mu.Unlock()
	if td.exporter != nil {
		return td.exporter.GetTargetURL()
	}
	return ""
}

// SetTargetURL dynamically reconfigures the telemetry destination endpoint.
func (td *TelemetryDispatcher) SetTargetURL(url string) {
	td.mu.Lock()
	defer td.mu.Unlock()
	if td.exporter != nil {
		td.exporter.SetTargetURL(url)
	}
}

// Dispatch processes vehicle telemetry export if the rate limit interval has passed.
func (td *TelemetryDispatcher) Dispatch(vehicles map[string]*entity.Vehicle) {
	td.mu.Lock()
	defer td.mu.Unlock()

	if !td.isExporting || td.exporter == nil || len(vehicles) == 0 {
		return
	}

	now := time.Now()
	if now.Sub(td.lastExportTime) < td.interval {
		return
	}
	td.lastExportTime = now

	for _, v := range vehicles {
		_ = td.exporter.ExportVehicle(context.Background(), v)
	}

	td.exportCount++
	if td.logger != nil && td.exportCount%5 == 1 {
		delivered, endpoint := td.exporter.GetDeliveryStatus()
		if delivered {
			td.logger.AddLog(fmt.Sprintf("📡 Telemetria entregue (1 Hz): %d viaturas recebidas por CARINA_CORE via %s", len(vehicles), endpoint))
		} else if td.exportCount > 3 {
			td.logger.AddLog(fmt.Sprintf("⚠️ Telemetria não entregue: CARINA_CORE não respondeu (%s). Inicie o CARINA_CORE na porta 2201.", endpoint))
		} else {
			td.logger.AddLog(fmt.Sprintf("📡 Transmitindo telemetria (1 Hz): %d viaturas para CARINA_CORE...", len(vehicles)))
		}
	}
}

// ExportVehicleNow immediately dispatches telemetry for a specific vehicle without rate limiting.
// Used for critical state transitions such as alert mode toggle.
func (td *TelemetryDispatcher) ExportVehicleNow(veh *entity.Vehicle) {
	td.mu.Lock()
	defer td.mu.Unlock()

	if !td.isExporting || td.exporter == nil || veh == nil {
		return
	}

	_ = td.exporter.ExportVehicle(context.Background(), veh)
}

