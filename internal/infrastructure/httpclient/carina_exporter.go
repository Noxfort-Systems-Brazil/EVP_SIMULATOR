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

// File: carina_exporter.go
// Author: Gabriel Moraes
// Date: October 2026

package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"evp_simulator/internal/application/dto"
	"evp_simulator/internal/domain/entity"
)

const (
	// DefaultLocalURL is the direct high-performance Go Web Gateway in CARINA_CORE.
	DefaultLocalURL = "http://localhost:2201/amb/v1"
)

// NormalizeTargetURL ensures the user-provided URL has a valid scheme and points to the telemetry endpoint.
func NormalizeTargetURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultLocalURL
	}
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		if strings.HasPrefix(raw, "localhost") || strings.HasPrefix(raw, "127.0.0.1") {
			raw = "http://" + raw
		} else {
			raw = "https://" + raw
		}
	}
	raw = strings.TrimRight(raw, "/")
	if strings.HasSuffix(raw, "/cco") {
		raw = strings.TrimSuffix(raw, "/cco")
	}
	if !strings.HasSuffix(raw, "/amb") && !strings.HasSuffix(raw, "/amb/v1") {
		raw = raw + "/amb/v1"
	}
	return raw
}

// CarinaExporter sends periodic telemetry HTTP POST requests to CARINA_CORE.
type CarinaExporter struct {
	mu               sync.RWMutex
	targetURL        string
	fallbackURL      string
	client           *http.Client
	active           bool
	lastSuccess      bool
	lastURL          string
	lastError        string
	primaryFailCount int
	latestTs         map[string]int64
}

// NewCarinaExporter creates a new CarinaExporter instance with dynamic endpoint support.
// Defaults to DefaultLocalURL (direct local gateway in CARINA_CORE).
func NewCarinaExporter(defaultURL string) *CarinaExporter {
	if defaultURL == "" {
		defaultURL = DefaultLocalURL
	} else {
		defaultURL = NormalizeTargetURL(defaultURL)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &CarinaExporter{
		targetURL:   defaultURL,
		fallbackURL: "",
		active:      true,
		client: &http.Client{
			Transport: transport,
			Timeout:   1500 * time.Millisecond,
		},
		latestTs: make(map[string]int64),
	}
}

func (ce *CarinaExporter) isStale(vehicleID string, ts int64) bool {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	if ce.latestTs == nil {
		return false
	}
	latest, ok := ce.latestTs[vehicleID]
	return ok && latest > ts
}

func (ce *CarinaExporter) IsActive() bool {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.active
}

func (ce *CarinaExporter) SetActive(active bool) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.active = active
}

func (ce *CarinaExporter) SetTargetURL(url string) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.targetURL = NormalizeTargetURL(url)
	ce.fallbackURL = ""
	ce.primaryFailCount = 0
}

func (ce *CarinaExporter) GetTargetURL() string {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.targetURL
}

// GetDeliveryStatus returns whether the last dispatch succeeded and the endpoint/error details.
func (ce *CarinaExporter) GetDeliveryStatus() (bool, string) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	if ce.lastSuccess {
		return true, ce.lastURL
	}
	if ce.lastError != "" {
		return false, ce.lastError
	}
	return false, "aguardando primeira tentativa de envio"
}

// ExportVehicle marshals vehicle state into the standardized JSON and sends via HTTP POST.
func (ce *CarinaExporter) ExportVehicle(ctx context.Context, veh *entity.Vehicle) error {
	ce.mu.Lock()
	if !ce.active {
		ce.mu.Unlock()
		return nil
	}
	url := ce.targetURL
	fallback := ce.fallbackURL
	client := ce.client

	ts := time.Now().UnixMilli()
	if ce.latestTs == nil {
		ce.latestTs = make(map[string]int64)
	}
	ce.latestTs[veh.ID] = ts
	ce.mu.Unlock()

	payload := dto.VehicleTelemetryDTO{
		VehicleID:   veh.ID,
		Corporation: veh.Corporation,
		Latitude:    veh.Position.Lat,
		Longitude:   veh.Position.Lon,
		SpeedKmh:    veh.Speed.Float64(),
		HeadingDeg:  veh.HeadingDeg,
		SirenActive: veh.Mode.IsAlert(),
		Timestamp:   ts,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Perform non-blocking resilient dispatch with dedicated request context and fallback
	go func() {
		if ce.isStale(veh.ID, ts) {
			return
		}

		reqCtx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode < 400 {
					ce.mu.Lock()
					ce.lastSuccess = true
					ce.lastURL = url
					ce.lastError = ""
					ce.primaryFailCount = 0
					ce.mu.Unlock()
					return
				}
			}
		}

		// Avoid attempting fallback if a newer telemetry packet was already created for this vehicle
		if ce.isStale(veh.ID, ts) {
			return
		}

		// Primary endpoint unreachable or returned error: attempt fallback candidate
		if fallback != "" && fallback != url {
			reqFbCtx, cancelFb := context.WithTimeout(context.Background(), 1500*time.Millisecond)
			defer cancelFb()

			reqFb, errFb := http.NewRequestWithContext(reqFbCtx, http.MethodPost, fallback, bytes.NewReader(bodyBytes))
			if errFb == nil {
				reqFb.Header.Set("Content-Type", "application/json")
				respFb, errFb := client.Do(reqFb)
				if errFb == nil {
					defer respFb.Body.Close()
					if respFb.StatusCode < 400 {
						ce.mu.Lock()
						ce.lastSuccess = true
						ce.lastURL = fallback
						ce.lastError = ""
						ce.primaryFailCount++
						if ce.primaryFailCount >= 2 {
							ce.targetURL = fallback
							ce.fallbackURL = url
							ce.primaryFailCount = 0
						}
						ce.mu.Unlock()
						return
					}
				}
			}
		}

		// Both endpoints failed
		ce.mu.Lock()
		ce.lastSuccess = false
		ce.lastError = "ambos endpoints inacessíveis (" + url + " e " + fallback + ")"
		ce.mu.Unlock()
	}()

	return nil
}
