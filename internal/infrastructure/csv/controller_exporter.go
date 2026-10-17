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

// File: controller_exporter.go
// Author: Gabriel Moraes
// Date: October 2026

package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"

	"evp_simulator/internal/domain/entity"
)

// ControllerCSVExporter handles exporting the traffic light controller inventory to CSV for CARINA_CORE.
type ControllerCSVExporter struct{}

// NewControllerCSVExporter creates a new ControllerCSVExporter instance.
func NewControllerCSVExporter() *ControllerCSVExporter {
	return &ControllerCSVExporter{}
}

// ExportControllersToCSV writes traffic lights to a CSV file matching CARINA's expected schema (Intersection ID, IP Address).
func (cce *ControllerCSVExporter) ExportControllersToCSV(filePath string, trafficLights []*entity.TrafficLight) error {
	if !strings.HasSuffix(strings.ToLower(filePath), ".csv") {
		filePath += ".csv"
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	// Write exact header expected by CARINA_CORE ConnectionConfigRepository
	header := []string{"Intersection ID", "IP Address"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Make a copy and sort by UTMCPort ascending (and ID as fallback) for deterministic output matching CARINA_MOCK_LIGHTS
	sortedLights := make([]*entity.TrafficLight, len(trafficLights))
	copy(sortedLights, trafficLights)
	sort.Slice(sortedLights, func(i, j int) bool {
		if sortedLights[i].UTMCPort != sortedLights[j].UTMCPort {
			return sortedLights[i].UTMCPort < sortedLights[j].UTMCPort
		}
		return sortedLights[i].ID < sortedLights[j].ID
	})

	for _, tl := range sortedLights {
		interID := tl.IntersectionID
		if interID == "" {
			interID = tl.ID
		}
		ipAddr := tl.IPAddress
		if ipAddr == "" && tl.UTMCPort > 0 {
			ipAddr = fmt.Sprintf("127.0.0.1:%d", tl.UTMCPort)
		} else if ipAddr == "" {
			ipAddr = "127.0.0.1:1610"
		}

		row := []string{interID, ipAddr}
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
