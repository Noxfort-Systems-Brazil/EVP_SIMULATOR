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

// File: controller_importer.go
// Author: Gabriel Moraes
// Date: October 2026

package csv

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"

	"evp_simulator/internal/domain/port"
)

// ControllerImporter parses controller inventory CSV spreadsheets.
type ControllerImporter struct{}

func NewControllerImporter() *ControllerImporter {
	return &ControllerImporter{}
}

// LoadControllersFromCSV loads mappings supporting UTF-8 BOM and multiple delimiters (, ; \t).
func (ci *ControllerImporter) LoadControllersFromCSV(filePath string) ([]port.ControllerMapping, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Strip UTF-8 BOM if present
	content = bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))

	reader := bufio.NewReader(bytes.NewReader(content))
	var mappings []port.ControllerMapping

	isHeader := true
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if line != "" {
			// Detect delimiter: check comma, semicolon, tab
			var parts []string
			if strings.Contains(line, ";") {
				parts = strings.Split(line, ";")
			} else if strings.Contains(line, "\t") {
				parts = strings.Split(line, "\t")
			} else {
				parts = strings.Split(line, ",")
			}

			if len(parts) >= 2 {
				col0 := strings.TrimSpace(parts[0])
				col1 := strings.TrimSpace(parts[1])

				if isHeader {
					// Check if this line is header (e.g. "Intersection ID", "IP Address")
					if strings.Contains(strings.ToLower(col0), "intersection") || strings.Contains(strings.ToLower(col1), "ip") {
						isHeader = false
						continue
					}
					isHeader = false
				}

				if col0 != "" && col1 != "" {
					mappings = append(mappings, port.ControllerMapping{
						IntersectionID: col0,
						IPAddress:      col1,
					})
				}
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}

	return mappings, nil
}
