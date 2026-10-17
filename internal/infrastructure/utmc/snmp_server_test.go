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

// File: snmp_server_test.go
// Author: Gabriel Moraes
// Date: October 2026

package utmc

import (
	"testing"
	"time"

	"evp_simulator/internal/domain/entity"
	"github.com/gosnmp/gosnmp"
)

func TestSNMPServer_FullUTMCProtocol(t *testing.T) {
	tl := entity.NewTrafficLight("J1", []entity.Phase{
		{Duration: 30, State: "GGrr"},
		{Duration: 30, State: "rrGG"},
	})
	tl.IntersectionID = "1116667894"
	tl.UTMCPort = 16198

	var loggedMsg string
	logger := func(msg string) {
		loggedMsg = msg
	}

	server := NewServer(tl, 16198, logger)
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	client := &gosnmp.GoSNMP{
		Target:    "127.0.0.1",
		Port:      16198,
		Community: "public",
		Version:   gosnmp.Version2c,
		Timeout:   2 * time.Second,
		Retries:   1,
	}

	if err := client.Connect(); err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer client.Conn.Close()

	// 1. GET sysDescr
	result, err := client.Get([]string{"1.3.6.1.2.1.1.1.0"})
	if err != nil {
		t.Fatalf("failed to GET sysDescr: %v", err)
	}
	sysDescr := string(result.Variables[0].Value.([]byte))
	if sysDescr != "CARINA UTMC2 Mock Traffic Controller (EVP Simulator)" {
		t.Errorf("unexpected sysDescr: %s", sysDescr)
	}

	// 2. GET status_active
	resultActive, err := client.Get([]string{"1.3.6.1.4.1.2825.4.2.1.1.4.1.4.1"})
	if err != nil {
		t.Fatalf("failed to GET status_active: %v", err)
	}
	activeVal := gosnmp.ToBigInt(resultActive.Variables[0].Value).Int64()
	if activeVal != 1 { // Phase 0 is active (1 << 0 = 1)
		t.Errorf("expected active green mask 1, got %d", activeVal)
	}

	// 3. SET HOLD (Stage 1 -> mask 2)
	pduHold := gosnmp.SnmpPDU{
		Name:  "1.3.6.1.4.1.2825.4.2.1.1.4.1.2.1",
		Type:  gosnmp.Integer,
		Value: 2,
	}
	setResult, err := client.Set([]gosnmp.SnmpPDU{pduHold})
	if err != nil {
		t.Fatalf("failed to SET HOLD: %v", err)
	}
	if setResult.Error != gosnmp.NoError {
		t.Errorf("SET error: %v", setResult.Error)
	}
	if tl.OverrideState != "HOLD" {
		t.Errorf("expected OverrideState HOLD, got %s", tl.OverrideState)
	}

	// 4. SET FLASH
	pduFlash := gosnmp.SnmpPDU{
		Name:  "1.3.6.1.4.1.2825.4.2.1.1.2.1.255.1",
		Type:  gosnmp.Integer,
		Value: 1,
	}
	_, err = client.Set([]gosnmp.SnmpPDU{pduFlash})
	if err != nil {
		t.Fatalf("failed to SET FLASH: %v", err)
	}
	if tl.OverrideState != "ALERT" {
		t.Errorf("expected OverrideState ALERT, got %s", tl.OverrideState)
	}

	// 5. SET FORCE_OFF
	pduForceOff := gosnmp.SnmpPDU{
		Name:  "1.3.6.1.4.1.2825.4.2.1.1.4.1.3.1",
		Type:  gosnmp.Integer,
		Value: 1,
	}
	_, err = client.Set([]gosnmp.SnmpPDU{pduForceOff})
	if err != nil {
		t.Fatalf("failed to SET FORCE_OFF: %v", err)
	}
	if tl.OverrideState != "NORMAL" {
		t.Errorf("expected OverrideState NORMAL, got %s", tl.OverrideState)
	}
	_ = loggedMsg
}
