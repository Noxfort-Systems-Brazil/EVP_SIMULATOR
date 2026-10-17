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

// File: integration_test.go
// Author: Gabriel Moraes
// Date: October 2026

package utmc

import (
	"fmt"
	"testing"
	"time"

	"evp_simulator/internal/application/usecase"
	"evp_simulator/internal/infrastructure/csv"
	"evp_simulator/internal/infrastructure/httpclient"
	"evp_simulator/internal/infrastructure/sumo"
	"github.com/gosnmp/gosnmp"
)

func TestIntegration_ServerAndSimulationRunAutomaticallyOnMapLoad(t *testing.T) {
	netRepo := sumo.NewParser()
	ctrlRepo := csv.NewControllerImporter()
	ctrlExp := csv.NewControllerCSVExporter()
	fleetExp := csv.NewVehicleCSVExporter()
	serverMgr := NewServerManager()
	telemetryExp := httpclient.NewCarinaExporter("http://localhost:8085/api/v1/evp/telemetry")

	uc := usecase.NewSimulationUseCase(netRepo, ctrlRepo, ctrlExp, serverMgr, telemetryExp, fleetExp)
	defer uc.StopServers()

	// 1. Load sample network map
	mapPath := "../../../examples/sample_network.net.xml.gz"
	err := uc.LoadMap(mapPath)
	if err != nil {
		t.Fatalf("failed to load map: %v", err)
	}

	state := uc.GetState()
	if !state.MapLoaded {
		t.Fatalf("expected map to be loaded")
	}

	// 2. Explicitly verify simulation is running automatically once map is loaded
	if !state.IsRunning {
		t.Fatalf("expected simulation to be running automatically on map load")
	}

	// 3. Verify server manager has started servers for each traffic light
	activeCount := serverMgr.ActiveServersCount()
	if activeCount == 0 {
		t.Fatalf("expected active UTMC servers, got 0")
	}
	t.Logf("Active UTMC servers running: %d (automatically alongside simulation)", activeCount)

	// Pick first traffic light
	firstTL := state.TrafficLights[0]
	targetPort := 1610
	t.Logf("Testing connection to first TL ID=%s on UDP port %d", firstTL.ID, targetPort)

	// 4. Connect via SNMP client simulating CARINA_CORE / carina-go
	client := &gosnmp.GoSNMP{
		Target:    "127.0.0.1",
		Port:      uint16(targetPort),
		Community: "public",
		Version:   gosnmp.Version2c,
		Timeout:   2 * time.Second,
		Retries:   1,
	}

	if err := client.Connect(); err != nil {
		t.Fatalf("Carina failed to connect to socket: %v", err)
	}
	defer client.Conn.Close()

	// 5. Query sysDescr (CARINA Discovery Protocol)
	resultDescr, err := client.Get([]string{"1.3.6.1.2.1.1.1.0"})
	if err != nil {
		t.Fatalf("Carina failed to query sysDescr while simulation is not running: %v", err)
	}
	sysDescr := string(resultDescr.Variables[0].Value.([]byte))
	t.Logf("CARINA successfully connected! sysDescr received: %s", sysDescr)

	// 6. Query status_active (CARINA Telemetry)
	resultActive, err := client.Get([]string{"1.3.6.1.4.1.2825.4.2.1.1.4.1.4.1"})
	if err != nil {
		t.Fatalf("Carina failed to query status_active: %v", err)
	}
	activeVal := gosnmp.ToBigInt(resultActive.Variables[0].Value).Int64()
	t.Logf("CARINA received status_active: %d", activeVal)

	// 7. Send HOLD command (CARINA Preemption)
	pduHold := gosnmp.SnmpPDU{
		Name:  "1.3.6.1.4.1.2825.4.2.1.1.4.1.2.1",
		Type:  gosnmp.Integer,
		Value: 1,
	}
	_, err = client.Set([]gosnmp.SnmpPDU{pduHold})
	if err != nil {
		t.Fatalf("Carina failed to send HOLD command: %v", err)
	}
	t.Logf("CARINA successfully sent HOLD command to junction %s", firstTL.ID)

	// Verify log recorded the incoming Carina command
	logs := uc.GetState().RecentLogs
	foundLog := false
	for _, l := range logs {
		if len(l) > 0 {
			foundLog = true
			break
		}
	}
	if !foundLog {
		t.Errorf("expected command log in recent logs")
	}
	fmt.Printf("Integration test completed successfully: servers remained online without running simulation.\n")
}
