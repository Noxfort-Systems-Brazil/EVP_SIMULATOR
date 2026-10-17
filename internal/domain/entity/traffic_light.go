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

// File: traffic_light.go
// Author: Gabriel Moraes
// Date: October 2026

package entity

import (
	"evp_simulator/internal/domain/valueobject"
)

// Phase represents a specific signal state combination and duration from SUMO.
type Phase struct {
	Duration float64 `json:"duration"`
	State    string  `json:"state"` // e.g. "rrGGyy"
}

// TrafficLight represents a smart signal controller managing an intersection.
type TrafficLight struct {
	ID             string  `json:"id"`              // SUMO ID
	IntersectionID string  `json:"intersection_id"` // CARINA ID (from CSV)
	IPAddress      string  `json:"ip_address"`      // Remote IP
	UTMCPort       int     `json:"utmc_port"`       // UDP port for SNMP (e.g. 1610+)
	Phases         []Phase `json:"phases"`

	CurrentPhaseIdx int     `json:"current_phase_idx"`
	TargetPhaseIdx  int     `json:"target_phase_idx"`
	PhaseTimer      float64 `json:"phase_timer"`

	TransitionState string  `json:"transition_state"` // "GREEN", "YELLOW", "ALL_RED"
	TransitionTimer float64 `json:"transition_timer"`
	TransitionLock  bool    `json:"transition_lock"`

	OverrideState string `json:"override_state"` // "NORMAL", "HOLD"
}

// NewTrafficLight creates a new TrafficLight instance.
func NewTrafficLight(id string, phases []Phase) *TrafficLight {
	tl := &TrafficLight{
		ID:              id,
		Phases:          phases,
		CurrentPhaseIdx: 0,
		TargetPhaseIdx:  0,
		TransitionState: "GREEN",
		OverrideState:   "NORMAL",
	}
	if len(phases) > 0 {
		tl.PhaseTimer = phases[0].Duration
	} else {
		tl.PhaseTimer = 30.0
	}
	return tl
}

// SetTargetPhase initiates a safety transition towards targetPhaseIdx.
func (tl *TrafficLight) SetTargetPhase(targetIdx int) {
	if len(tl.Phases) == 0 {
		return
	}
	if targetIdx < 0 || targetIdx >= len(tl.Phases) {
		return
	}
	if tl.CurrentPhaseIdx == targetIdx && tl.TransitionState == "GREEN" {
		tl.TargetPhaseIdx = targetIdx
		return
	}

	tl.TargetPhaseIdx = targetIdx

	if tl.TransitionState == "GREEN" && !tl.TransitionLock {
		// Identify if yellow transition is required
		currState := tl.Phases[tl.CurrentPhaseIdx].State
		targetState := tl.Phases[tl.TargetPhaseIdx].State

		needsYellow := false
		for i := 0; i < len(currState) && i < len(targetState); i++ {
			c := currState[i]
			t := targetState[i]
			if (c == 'G' || c == 'g') && (t != 'G' && t != 'g') {
				needsYellow = true
				break
			}
		}

		if needsYellow {
			tl.TransitionState = "YELLOW"
			tl.TransitionTimer = 3.0 // 3 seconds yellow safety clearance
			tl.TransitionLock = true
		} else {
			tl.TransitionState = "ALL_RED"
			tl.TransitionTimer = 1.0 // 1 second all-red clearance
			tl.TransitionLock = true
		}
	}
}

// Tick advances the signal timer by dt seconds.
func (tl *TrafficLight) Tick(dt float64) {
	if len(tl.Phases) == 0 {
		return
	}

	// 1. Handle remote safety transition state machine
	if tl.TransitionLock {
		if tl.TransitionState == "YELLOW" {
			tl.TransitionTimer -= dt
			if tl.TransitionTimer <= 0 {
				tl.TransitionState = "ALL_RED"
				tl.TransitionTimer = 1.0
			}
		} else if tl.TransitionState == "ALL_RED" {
			tl.TransitionTimer -= dt
			if tl.TransitionTimer <= 0 {
				tl.TransitionState = "GREEN"
				tl.TransitionLock = false
				tl.CurrentPhaseIdx = tl.TargetPhaseIdx
				tl.PhaseTimer = tl.Phases[tl.CurrentPhaseIdx].Duration
			}
		}
		return
	}

	// 2. If in HOLD (commanded by CARINA), maintain the current target phase
	if tl.OverrideState == "HOLD" {
		return
	}

	// 3. Autonomous fixed plan cycling
	tl.PhaseTimer -= dt
	if tl.PhaseTimer <= 0 {
		nextPhase := (tl.CurrentPhaseIdx + 1) % len(tl.Phases)
		tl.SetTargetPhase(nextPhase)
	}
}

// GetCurrentState returns the effective state string of all links.
func (tl *TrafficLight) GetCurrentState() string {
	if len(tl.Phases) == 0 {
		return ""
	}

	currState := tl.Phases[tl.CurrentPhaseIdx].State
	numLinks := len(currState)

	if tl.TransitionState == "ALL_RED" {
		res := make([]byte, numLinks)
		for i := range res {
			res[i] = 'r'
		}
		return string(res)
	}

	if tl.TransitionState == "YELLOW" {
		res := make([]byte, numLinks)
		targetState := ""
		if tl.TargetPhaseIdx >= 0 && tl.TargetPhaseIdx < len(tl.Phases) {
			targetState = tl.Phases[tl.TargetPhaseIdx].State
		}

		for i := 0; i < numLinks; i++ {
			c := currState[i]
			t := byte('r')
			if i < len(targetState) {
				t = targetState[i]
			}

			if (c == 'G' || c == 'g') && (t != 'G' && t != 'g') {
				res[i] = 'y'
			} else {
				res[i] = c
			}
		}
		return string(res)
	}

	return currState
}

// GetSignalForLink evaluates the exact signal state for linkIndex.
func (tl *TrafficLight) GetSignalForLink(linkIndex int) valueobject.SignalState {
	stateStr := tl.GetCurrentState()
	if linkIndex < 0 || linkIndex >= len(stateStr) {
		return valueobject.SignalGreen // default open if unmonitored
	}
	return valueobject.ParseSignalChar(stateStr[linkIndex])
}
