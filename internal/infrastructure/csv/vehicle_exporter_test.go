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

// File: vehicle_exporter_test.go
// Author: Gabriel Moraes
// Date: October 2026

package csv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/valueobject"
)

func TestVehicleCSVExporter_ExportFleetToCSV(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "vehicles_export.csv")

	exporter := NewVehicleCSVExporter()

	v1 := entity.NewVehicle("SAMU-01", "SAMU", valueobject.Coordinate{})
	v1.Plate = "BRA2E19"
	v1.DeviceID = "IMEI-864209041234567"
	v1.IPAddress = "10.8.0.15"
	v1.Port = 8085
	v1.Protocol = "HTTP"
	v1.AuthToken = "sec_token_samu_01_xyz"
	v1.IsActive = true

	v2 := entity.NewVehicle("BOMBEIRO-01", "BOMBEIRO", valueobject.Coordinate{})
	v2.Plate = "SPB1930"
	v2.DeviceID = "IMEI-864209041234569"
	v2.IPAddress = "10.8.0.20"
	v2.Port = 8085
	v2.Protocol = "HTTP"
	v2.AuthToken = "sec_token_bomb_01_xyz"
	v2.IsActive = true

	err := exporter.ExportFleetToCSV(outPath, []*entity.Vehicle{v1, v2})
	if err != nil {
		t.Fatalf("unexpected error exporting fleet CSV: %v", err)
	}

	contentBytes, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed reading exported CSV: %v", err)
	}
	content := string(contentBytes)

	expectedHeader := "vehicle_id,corporation,plate,device_id,ip_address,port,protocol,auth_token,is_active"
	if !strings.Contains(content, expectedHeader) {
		t.Errorf("header missing in exported CSV:\n%s", content)
	}

	expectedRow1 := "SAMU-01,SAMU,BRA2E19,IMEI-864209041234567,10.8.0.15,8085,HTTP,sec_token_samu_01_xyz,true"
	if !strings.Contains(content, expectedRow1) {
		t.Errorf("expected row 1 missing in exported CSV:\n%s", content)
	}

	expectedRow2 := "BOMBEIRO-01,BOMBEIRO,SPB1930,IMEI-864209041234569,10.8.0.20,8085,HTTP,sec_token_bomb_01_xyz,true"
	if !strings.Contains(content, expectedRow2) {
		t.Errorf("expected row 2 missing in exported CSV:\n%s", content)
	}
}
