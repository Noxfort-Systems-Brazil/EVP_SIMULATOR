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

// File: vehicle_exporter.go
// Author: Gabriel Moraes
// Date: October 2026

package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"evp_simulator/internal/domain/entity"
)

// VehicleCSVExporter handles exporting the simulated fleet configuration to CSV for CARINA_CORE.
type VehicleCSVExporter struct{}

func NewVehicleCSVExporter() *VehicleCSVExporter {
	return &VehicleCSVExporter{}
}

// ExportFleetToCSV writes vehicles to a CSV file matching CARINA's expected schema.
func (vce *VehicleCSVExporter) ExportFleetToCSV(filePath string, vehicles []*entity.Vehicle) error {
	if !strings.HasSuffix(strings.ToLower(filePath), ".csv") {
		filePath += ".csv"
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	// Write exact header specified
	header := []string{
		"vehicle_id",
		"corporation",
		"plate",
		"device_id",
		"ip_address",
		"port",
		"protocol",
		"auth_token",
		"is_active",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write rows
	for _, v := range vehicles {
		portStr := fmt.Sprintf("%d", v.Port)
		if v.Port == 0 {
			portStr = "8085"
		}
		proto := v.Protocol
		if proto == "" {
			proto = "HTTP"
		}
		activeStr := "true"
		if !v.IsActive {
			activeStr = "false"
		}

		row := []string{
			v.ID,
			v.Corporation,
			v.Plate,
			v.DeviceID,
			v.IPAddress,
			portStr,
			proto,
			v.AuthToken,
			activeStr,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	return nil
}
