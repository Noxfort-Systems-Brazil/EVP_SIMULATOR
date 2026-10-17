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

// File: controller_importer_test.go
// Author: Gabriel Moraes
// Date: October 2026

package csv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestControllerImporter_LoadControllersFromCSV(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Test standard comma-separated with header
	csvComma := filepath.Join(tmpDir, "controllers_comma.csv")
	err := os.WriteFile(csvComma, []byte("Intersection ID, IP Address\n1116667894, 192.168.1.100\n1193472473, 192.168.1.101\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write test csv: %v", err)
	}

	importer := NewControllerImporter()
	mappings, err := importer.LoadControllersFromCSV(csvComma)
	if err != nil {
		t.Fatalf("unexpected error loading comma csv: %v", err)
	}
	if len(mappings) != 2 {
		t.Fatalf("expected 2 mappings, got %d", len(mappings))
	}
	if mappings[0].IntersectionID != "1116667894" || mappings[0].IPAddress != "192.168.1.100" {
		t.Errorf("unexpected mapping[0]: %+v", mappings[0])
	}

	// 2. Test semicolon with UTF-8 BOM
	csvSemicolonBOM := filepath.Join(tmpDir, "controllers_bom.csv")
	bomData := append([]byte("\xef\xbb\xbf"), []byte("Intersection ID;IP Address\r\n222333444;10.0.0.50\r\n")...)
	err = os.WriteFile(csvSemicolonBOM, bomData, 0644)
	if err != nil {
		t.Fatalf("failed to write bom csv: %v", err)
	}

	mappingsBOM, err := importer.LoadControllersFromCSV(csvSemicolonBOM)
	if err != nil {
		t.Fatalf("unexpected error loading bom csv: %v", err)
	}
	if len(mappingsBOM) != 1 {
		t.Fatalf("expected 1 mapping, got %d", len(mappingsBOM))
	}
	if mappingsBOM[0].IntersectionID != "222333444" || mappingsBOM[0].IPAddress != "10.0.0.50" {
		t.Errorf("unexpected mapping: %+v", mappingsBOM[0])
	}
}
