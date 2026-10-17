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

// File: signal_state.go
// Author: Gabriel Moraes
// Date: October 2026

package valueobject

// SignalState represents the traffic light signal color for a specific movement/lane.
type SignalState string

const (
	SignalGreen  SignalState = "GREEN"
	SignalYellow SignalState = "YELLOW"
	SignalRed    SignalState = "RED"
	SignalOff    SignalState = "OFF"
)

// ParseSignalChar parses a single character from a SUMO phase state string (e.g. 'G', 'g', 'y', 'r').
func ParseSignalChar(c byte) SignalState {
	switch c {
	case 'G', 'g':
		return SignalGreen
	case 'y', 'Y', 'u':
		return SignalYellow
	case 'r', 'R':
		return SignalRed
	case 'O', 'o':
		return SignalOff
	default:
		return SignalRed
	}
}

// IsOpen returns true if traffic is allowed to proceed.
func (s SignalState) IsOpen() bool {
	return s == SignalGreen
}
