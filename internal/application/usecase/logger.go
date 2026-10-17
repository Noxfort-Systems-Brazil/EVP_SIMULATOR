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

// File: logger.go
// Author: Gabriel Moraes
// Date: October 2026

package usecase

import (
	"fmt"
	"sync"
	"time"
)

// SimulationLogger maintains an in-memory sliding buffer of simulation event logs.
type SimulationLogger struct {
	mu         sync.RWMutex
	logs       []string
	maxEntries int
}

// NewSimulationLogger creates a logger with a defined capacity limit.
func NewSimulationLogger(maxEntries int) *SimulationLogger {
	if maxEntries <= 0 {
		maxEntries = 50
	}
	return &SimulationLogger{
		logs:       make([]string, 0, maxEntries),
		maxEntries: maxEntries,
	}
}

// AddLog appends a timestamped log message, keeping only the most recent entries.
func (l *SimulationLogger) AddLog(msg string) {
	entry := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)

	l.mu.Lock()
	defer l.mu.Unlock()

	l.logs = append(l.logs, entry)
	if len(l.logs) > l.maxEntries {
		l.logs = l.logs[len(l.logs)-l.maxEntries:]
	}
}

// GetLogs returns a defensive copy of recent log messages.
func (l *SimulationLogger) GetLogs() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]string, len(l.logs))
	copy(result, l.logs)
	return result
}
