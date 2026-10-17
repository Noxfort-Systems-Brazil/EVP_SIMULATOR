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

// File: telemetry_exporter.go
// Author: Gabriel Moraes
// Date: October 2026

package port

import (
	"context"
	"evp_simulator/internal/domain/entity"
)

// TelemetryExporter defines the output port for broadcasting vehicle telemetry to CARINA_CORE.
type TelemetryExporter interface {
	ExportVehicle(ctx context.Context, veh *entity.Vehicle) error
	IsActive() bool
	SetActive(active bool)
	SetTargetURL(url string)
	GetTargetURL() string
	GetDeliveryStatus() (bool, string)
}
