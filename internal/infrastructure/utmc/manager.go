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

// File: manager.go
// Author: Gabriel Moraes
// Date: October 2026

package utmc

import (
	"fmt"
	"sort"
	"sync"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/port"
)

// ServerManager manages the lifecycle of multiple UTMC2 SNMP UDP agent servers.
type ServerManager struct {
	mu      sync.Mutex
	servers []*Server
}

var _ port.ControllerServerManager = (*ServerManager)(nil)

// NewServerManager creates a new ServerManager instance.
func NewServerManager() *ServerManager {
	return &ServerManager{
		servers: make([]*Server, 0),
	}
}

// StartServers terminates any active servers and starts an SNMP UDP listener for each traffic light.
func (m *ServerManager) StartServers(lights map[string]*entity.TrafficLight, logger func(string)) error {
	m.StopServers()

	m.mu.Lock()
	defer m.mu.Unlock()

	if len(lights) == 0 {
		return nil
	}

	started := 0
	minPort := 65535
	maxPort := 0

	// Sort lights by UTMCPort for deterministic startup order
	lightList := make([]*entity.TrafficLight, 0, len(lights))
	for _, tl := range lights {
		lightList = append(lightList, tl)
	}
	sort.Slice(lightList, func(i, j int) bool {
		return lightList[i].UTMCPort < lightList[j].UTMCPort
	})

	for _, tl := range lightList {
		if tl.UTMCPort <= 0 {
			continue
		}

		server := NewServer(tl, tl.UTMCPort, logger)
		if err := server.Start(); err != nil {
			if logger != nil {
				logger(fmt.Sprintf("⚠️ Falha ao vincular UDP :%d para semáforo %s: %v", tl.UTMCPort, tl.ID, err))
			}
			continue
		}

		m.servers = append(m.servers, server)
		started++
		if tl.UTMCPort < minPort {
			minPort = tl.UTMCPort
		}
		if tl.UTMCPort > maxPort {
			maxPort = tl.UTMCPort
		}
	}

	if started > 0 && logger != nil {
		logger(fmt.Sprintf("🚦 %d servidores UTMC/SNMP INICIADOS e ONLINE (portas %d a %d) • Para conectar no CARINA_CORE: use 'Exportar Controladores CSV' ou configure 127.0.0.1:<porta> na aba Conexões do CARINA!",
			started, minPort, maxPort))
	}

	return nil
}

// StopServers safely shuts down all active UDP listeners.
func (m *ServerManager) StopServers() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, s := range m.servers {
		s.Stop()
	}
	m.servers = make([]*Server, 0)
}

// ActiveServersCount returns the total number of running UDP listeners.
func (m *ServerManager) ActiveServersCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.servers)
}
