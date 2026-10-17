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

// File: carina_exporter_test.go
// Author: Gabriel Moraes
// Date: October 2026

package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/valueobject"
)

func TestCarinaExporter_DefaultsAndToggle(t *testing.T) {
	exp := NewCarinaExporter("")
	if exp.GetTargetURL() != DefaultLocalURL {
		t.Errorf("expected default URL %s, got %s", DefaultLocalURL, exp.GetTargetURL())
	}
	if !exp.IsActive() {
		t.Errorf("expected exporter to be active")
	}

	exp.SetTargetURL("http://localhost:2201")
	if exp.GetTargetURL() != DefaultLocalURL {
		t.Errorf("expected target URL %s, got %s", DefaultLocalURL, exp.GetTargetURL())
	}

	exp.SetTargetURL("https://carina.tail08aec6.ts.net/cco")
	expectedTailscale := "https://carina.tail08aec6.ts.net/amb/v1"
	if exp.GetTargetURL() != expectedTailscale {
		t.Errorf("expected target URL %s, got %s", expectedTailscale, exp.GetTargetURL())
	}
}

func TestCarinaExporter_ExportSuccess(t *testing.T) {
	var received atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			received.Add(1)
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	exp := NewCarinaExporter(server.URL)

	veh := &entity.Vehicle{
		ID:          "AMB-01",
		Corporation: "SAMU",
		Position:    valueobject.Coordinate{Lat: -23.55, Lon: -46.63, X: 10, Y: 20},
		Speed:       valueobject.SpeedKmh(15.0),
		HeadingDeg:  90.0,
		Mode:        valueobject.ModeAlert,
	}

	err := exp.ExportVehicle(context.Background(), veh)
	if err != nil {
		t.Fatalf("unexpected error exporting vehicle: %v", err)
	}

	// Wait up to 500ms for goroutine dispatch
	time.Sleep(100 * time.Millisecond)

	if received.Load() != 1 {
		t.Errorf("expected 1 received request, got %d", received.Load())
	}
}

func TestCarinaExporter_Fallback(t *testing.T) {
	var fallbackReceived atomic.Int32

	fallbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fallbackReceived.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer fallbackServer.Close()

	// Dead primary server port
	exp := NewCarinaExporter("http://127.0.0.1:59999/dead")
	exp.fallbackURL = fallbackServer.URL

	veh := &entity.Vehicle{
		ID:          "POL-01",
		Corporation: "POLICIA",
		Position:    valueobject.Coordinate{Lat: -23.55, Lon: -46.63},
		Speed:       valueobject.SpeedKmh(20.0),
		HeadingDeg:  180.0,
		Mode:        valueobject.ModeNormal,
	}

	err := exp.ExportVehicle(context.Background(), veh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Give a little time for primary fail + fallback dispatch
	time.Sleep(300 * time.Millisecond)

	if fallbackReceived.Load() != 1 {
		t.Errorf("expected fallback to receive request, got %d", fallbackReceived.Load())
	}
}

func TestCarinaExporter_TimestampAndStaleProtection(t *testing.T) {
	var receivedTimestamp int64
	var sirenState bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if ts, ok := payload["timestamp"].(float64); ok {
			receivedTimestamp = int64(ts)
		}
		if s, ok := payload["siren_active"].(bool); ok {
			sirenState = s
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	exp := NewCarinaExporter(server.URL)

	veh := &entity.Vehicle{
		ID:          "AMB-99",
		Corporation: "SAMU",
		Position:    valueobject.Coordinate{Lat: -23.55, Lon: -46.63},
		Speed:       valueobject.SpeedKmh(30.0),
		HeadingDeg:  0.0,
		Mode:        valueobject.ModeNormal,
	}

	err := exp.ExportVehicle(context.Background(), veh)
	if err != nil {
		t.Fatalf("unexpected export error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if receivedTimestamp <= 0 {
		t.Errorf("expected positive timestamp in telemetry payload, got %d", receivedTimestamp)
	}
	if sirenState != false {
		t.Errorf("expected siren_active false, got %v", sirenState)
	}

	// Verify staleness logic directly
	oldTs := receivedTimestamp - 1000
	if !exp.isStale("AMB-99", oldTs) {
		t.Errorf("expected timestamp %d to be identified as stale", oldTs)
	}
	futureTs := receivedTimestamp + 1000
	if exp.isStale("AMB-99", futureTs) {
		t.Errorf("expected timestamp %d to NOT be stale", futureTs)
	}
}

