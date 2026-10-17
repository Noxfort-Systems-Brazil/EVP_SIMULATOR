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

// File: parser_test.go
// Author: Gabriel Moraes
// Date: October 2026

package sumo

import (
	"bytes"
	"compress/gzip"
	"math"
	"os"
	"path/filepath"
	"testing"
)

const sampleNetXML = `<?xml version="1.0" encoding="UTF-8"?>
<net version="1.20" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <location netOffset="0.00,0.00" convBoundary="0.00,0.00,500.00,500.00" origBoundary="-46.633,-23.550,-46.620,-23.540" />
    <junction id="J0" type="dead_end" x="0.00" y="100.00" incLanes="" />
    <junction id="J1" type="traffic_light" x="200.00" y="100.00" incLanes="edge_0_0" />
    <junction id="J2" type="dead_end" x="400.00" y="100.00" incLanes="edge_1_0" />

    <edge id="edge_0" from="J0" to="J1" priority="7" numLanes="1">
        <lane id="edge_0_0" index="0" speed="16.67" length="200.00" shape="0.00,100.00 200.00,100.00" />
    </edge>
    <edge id="edge_1" from="J1" to="J2" priority="7" numLanes="1">
        <lane id="edge_1_0" index="0" speed="16.67" length="200.00" shape="200.00,100.00 400.00,100.00" />
    </edge>

    <connection from="edge_0" to="edge_1" fromLane="0" toLane="0" tl="J1" linkIndex="0" dir="s" />

    <tlLogic id="J1" type="static" programID="0" offset="0">
        <phase duration="42" state="G" />
        <phase duration="3"  state="y" />
        <phase duration="42" state="r" />
    </tlLogic>
</net>`

func TestSumoParser_LoadNetworkFromSUMO(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Test uncompressed .net.xml
	xmlPath := filepath.Join(tmpDir, "test.net.xml")
	err := os.WriteFile(xmlPath, []byte(sampleNetXML), 0644)
	if err != nil {
		t.Fatalf("failed to write net xml: %v", err)
	}

	parser := NewParser()
	net, tls, err := parser.LoadNetworkFromSUMO(xmlPath)
	if err != nil {
		t.Fatalf("failed to parse net xml: %v", err)
	}

	if len(net.Edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(net.Edges))
	}
	if len(net.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(net.Nodes))
	}
	if len(tls) != 1 {
		t.Errorf("expected 1 traffic light, got %d", len(tls))
	}

	conn := net.FindConnection("edge_0", 0, "edge_1")
	if conn == nil {
		t.Fatalf("expected connection edge_0 -> edge_1")
	}
	if conn.TrafficLightID != "J1" || conn.LinkIndex != 0 {
		t.Errorf("unexpected connection properties: %+v", conn)
	}

	// Verify coordinates conversion
	j0, ok := net.Nodes["J0"]
	if !ok {
		t.Fatalf("expected node J0 to exist")
	}
	if j0.Position.X != 0 || j0.Position.Y != 100 {
		t.Errorf("unexpected Cartesian coords for J0: (%f, %f)", j0.Position.X, j0.Position.Y)
	}
	if j0.Position.Lon != -46.633 || (j0.Position.Lat < -23.548001 || j0.Position.Lat > -23.547999) {
		t.Errorf("unexpected GPS coords for J0: Lat=%f, Lon=%f", j0.Position.Lat, j0.Position.Lon)
	}

	// 2. Test compressed .net.xml.gz
	gzPath := filepath.Join(tmpDir, "test.net.xml.gz")
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write([]byte(sampleNetXML))
	_ = gw.Close()
	_ = os.WriteFile(gzPath, buf.Bytes(), 0644)

	netGz, tlsGz, err := parser.LoadNetworkFromSUMO(gzPath)
	if err != nil {
		t.Fatalf("failed to parse gzipped net: %v", err)
	}
	if len(netGz.Edges) != 2 || len(tlsGz) != 1 {
		t.Errorf("mismatch in gzipped net parse: edges=%d, tls=%d", len(netGz.Edges), len(tlsGz))
	}
}

func TestSumoParser_ArbitraryCityAndSynthetic(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Test with an arbitrary city (e.g. Lisbon: Lon ~ -9.14, Lat ~ 38.72)
	lisbonXML := `<?xml version="1.0" encoding="UTF-8"?>
<net version="1.20">
    <location convBoundary="0.0,0.0,1000.0,1000.0" origBoundary="-9.140,38.720,-9.130,38.730" />
    <junction id="J1" type="dead_end" x="500.0" y="500.0" />
</net>`
	lisbonPath := filepath.Join(tmpDir, "lisbon.net.xml")
	_ = os.WriteFile(lisbonPath, []byte(lisbonXML), 0644)

	parser := NewParser()
	net, _, err := parser.LoadNetworkFromSUMO(lisbonPath)
	if err != nil {
		t.Fatalf("failed to load lisbon map: %v", err)
	}
	j1 := net.Nodes["J1"]
	if math.Abs(j1.Position.Lon-(-9.135)) > 1e-6 || math.Abs(j1.Position.Lat-38.725) > 1e-6 {
		t.Errorf("unexpected Lisbon GPS: Lat=%f, Lon=%f", j1.Position.Lat, j1.Position.Lon)
	}

	// 2. Test synthetic network without location metadata
	synthXML := `<?xml version="1.0" encoding="UTF-8"?>
<net version="1.20">
    <junction id="JS" type="dead_end" x="120.0" y="340.0" />
</net>`
	synthPath := filepath.Join(tmpDir, "synth.net.xml")
	_ = os.WriteFile(synthPath, []byte(synthXML), 0644)

	netSynth, _, err := parser.LoadNetworkFromSUMO(synthPath)
	if err != nil {
		t.Fatalf("failed to load synth map: %v", err)
	}
	js := netSynth.Nodes["JS"]
	if js.Position.X != 120.0 || js.Position.Y != 340.0 || js.Position.Lat != 0 || js.Position.Lon != 0 {
		t.Errorf("unexpected synthetic node coords: %+v", js.Position)
	}
}

func TestSumoParser_ActualSampleNetwork(t *testing.T) {
	samplePath := filepath.Join("..", "..", "..", "examples", "sample_network.net.xml.gz")
	if _, err := os.Stat(samplePath); err != nil {
		samplePath = filepath.Join("..", "..", "..", "sample_network.net.xml.gz")
		if _, err := os.Stat(samplePath); err != nil {
			t.Skip("sample_network.net.xml.gz not found, skipping")
		}
	}


	parser := NewParser()
	net, tls, err := parser.LoadNetworkFromSUMO(samplePath)
	if err != nil {
		t.Fatalf("failed to load actual sample network: %v", err)
	}

	if len(net.Edges) == 0 {
		t.Errorf("expected edges > 0, got %d", len(net.Edges))
	}
	if len(tls) == 0 {
		t.Errorf("expected tls > 0, got %d", len(tls))
	}
}

