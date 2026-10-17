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

// File: controller_exporter_test.go
// Author: Gabriel Moraes
// Date: October 2026

package csv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"evp_simulator/internal/domain/entity"
)

func TestControllerCSVExporter_ExportControllersToCSV(t *testing.T) {
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "controllers_export.csv")

	tl1 := entity.NewTrafficLight("J2", nil)
	tl1.IntersectionID = "J2"
	tl1.UTMCPort = 1611
	tl1.IPAddress = "127.0.0.1:1611"

	tl2 := entity.NewTrafficLight("J1", nil)
	tl2.IntersectionID = "J1"
	tl2.UTMCPort = 1610
	tl2.IPAddress = "127.0.0.1:1610"

	lights := []*entity.TrafficLight{tl1, tl2}

	exporter := NewControllerCSVExporter()
	err := exporter.ExportControllersToCSV(exportPath, lights)
	if err != nil {
		t.Fatalf("unexpected error exporting controllers: %v", err)
	}

	contentBytes, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	content := string(contentBytes)
	lines := strings.Split(strings.TrimSpace(content), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (1 header + 2 items), got %d: \n%s", len(lines), content)
	}

	expectedHeader := "Intersection ID,IP Address"
	if strings.ReplaceAll(lines[0], " ", "") != strings.ReplaceAll(expectedHeader, " ", "") {
		t.Errorf("expected header '%s', got '%s'", expectedHeader, lines[0])
	}

	// Should be sorted by UTMCPort ascending: J1 (1610) then J2 (1611)
	if !strings.Contains(lines[1], "J1") || !strings.Contains(lines[1], "127.0.0.1:1610") {
		t.Errorf("expected first row to be J1 (port 1610), got '%s'", lines[1])
	}
	if !strings.Contains(lines[2], "J2") || !strings.Contains(lines[2], "127.0.0.1:1611") {
		t.Errorf("expected second row to be J2 (port 1611), got '%s'", lines[2])
	}
}

func TestControllerCSVExporter_ExportToDesktop(t *testing.T) {
	desktopDir := "/home/gabriel-moraes/Área de trabalho"
	if _, err := os.Stat(desktopDir); os.IsNotExist(err) {
		t.Skip("Desktop directory does not exist on this test environment")
	}

	testTarget := filepath.Join(desktopDir, "test_controllers_export.csv")
	defer os.Remove(testTarget)

	tl := entity.NewTrafficLight("1116667894", nil)
	tl.IntersectionID = "1116667894"
	tl.UTMCPort = 1610
	tl.IPAddress = "127.0.0.1:1610"

	exporter := NewControllerCSVExporter()
	// Test path without .csv to verify automatic suffix appending
	targetWithoutExt := filepath.Join(desktopDir, "test_controllers_export")
	err := exporter.ExportControllersToCSV(targetWithoutExt, []*entity.TrafficLight{tl})
	if err != nil {
		t.Fatalf("failed to export to desktop: %v", err)
	}

	if _, err := os.Stat(testTarget); err != nil {
		t.Fatalf("expected file '%s' to be created: %v", testTarget, err)
	}

	contentBytes, err := os.ReadFile(testTarget)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if !strings.Contains(string(contentBytes), "1116667894,127.0.0.1:1610") && !strings.Contains(string(contentBytes), "1116667894, 127.0.0.1:1610") {
		t.Errorf("unexpected content in exported file: %s", string(contentBytes))
	}
}

