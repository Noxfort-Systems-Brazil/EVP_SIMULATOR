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

// File: controller_repo.go
// Author: Gabriel Moraes
// Date: October 2026

package port

import (
	"evp_simulator/internal/domain/entity"
)

// ControllerMapping represents an entry from the CSV inventory.
type ControllerMapping struct {
	IntersectionID string `json:"intersection_id"`
	IPAddress      string `json:"ip_address"`
}

// ControllerRepository defines the port for loading controller inventories.
type ControllerRepository interface {
	LoadControllersFromCSV(filePath string) ([]ControllerMapping, error)
}

// ControllerExporter defines the port for exporting traffic light controller inventories to CSV for CARINA_CORE.
type ControllerExporter interface {
	ExportControllersToCSV(filePath string, trafficLights []*entity.TrafficLight) error
}

// NetworkRepository defines the port for loading and querying road network models.
type NetworkRepository interface {
	LoadNetworkFromSUMO(filePath string) (*entity.RoadNetwork, map[string]*entity.TrafficLight, error)
}

// ControllerServerManager defines the port for managing live UTMC/SNMP UDP agent servers.
type ControllerServerManager interface {
	StartServers(lights map[string]*entity.TrafficLight, logger func(string)) error
	StopServers()
	ActiveServersCount() int
}

