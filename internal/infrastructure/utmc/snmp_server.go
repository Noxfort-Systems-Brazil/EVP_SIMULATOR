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

// File: snmp_server.go
// Author: Gabriel Moraes
// Date: October 2026

package utmc

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"evp_simulator/internal/domain/entity"
	"github.com/gosnmp/gosnmp"
)

// Server simulates the UTMC2 SNMP UDP agent for a traffic light controller.
type Server struct {
	mu           sync.Mutex
	tl           *entity.TrafficLight
	port         int
	conn         *net.UDPConn
	running      bool
	onCommandLog func(string)
}

// NewServer creates a UTMC2 UDP listener for a traffic light.
func NewServer(tl *entity.TrafficLight, port int, logger func(string)) *Server {
	return &Server{
		tl:           tl,
		port:         port,
		onCommandLog: logger,
	}
}

// Start begins listening for UTMC2 UDP datagrams.
func (s *Server) Start() error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", s.port))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	s.conn = conn
	s.running = true

	go s.listenLoop()
	return nil
}

// Stop closes the UDP listener.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *Server) listenLoop() {
	buf := make([]byte, 4096)
	for s.running {
		n, remoteAddr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			if !s.running {
				return
			}
			continue
		}

		packetData := make([]byte, n)
		copy(packetData, buf[:n])
		s.handlePacket(packetData, remoteAddr)
	}
}

func (s *Server) handlePacket(data []byte, remoteAddr *net.UDPAddr) {
	// Try parsing via gosnmp
	g := &gosnmp.GoSNMP{Version: gosnmp.Version2c}
	packet, err := g.UnmarshalTrap(data, false)
	if err == nil && packet != nil {
		s.handleSNMPPacket(packet, remoteAddr)
		return
	}

	// Fallback for raw text/command strings
	s.handleFallbackPacket(data, remoteAddr)
}

func (s *Server) handleSNMPPacket(packet *gosnmp.SnmpPacket, remoteAddr *net.UDPAddr) {
	respVars := make([]gosnmp.SnmpPDU, 0, len(packet.Variables))

	switch packet.PDUType {
	case gosnmp.GetRequest, gosnmp.GetNextRequest:
		for _, pdu := range packet.Variables {
			oid := strings.TrimPrefix(pdu.Name, ".")
			respPdu := gosnmp.SnmpPDU{
				Name: pdu.Name,
			}

			switch oid {
			case "1.3.6.1.2.1.1.1.0": // sysDescr
				respPdu.Type = gosnmp.OctetString
				respPdu.Value = []byte("CARINA UTMC2 Mock Traffic Controller (EVP Simulator)")
			case "1.3.6.1.4.1.2825.4.2.1.1.4.1.4.1": // status_active (greens)
				respPdu.Type = gosnmp.Integer
				stageMask := 0
				if s.tl.TransitionState == "GREEN" {
					stageMask = 1 << s.tl.CurrentPhaseIdx
				}
				respPdu.Value = stageMask
			case "1.3.6.1.4.1.2825.4.2.1.1.4.1.8.1": // status_leaving (yellows)
				respPdu.Type = gosnmp.Integer
				stageMask := 0
				if s.tl.TransitionState == "YELLOW" {
					stageMask = 1 << s.tl.CurrentPhaseIdx
				}
				respPdu.Value = stageMask
			case "1.3.6.1.4.1.2825.4.2.1.1.4.1.5.1": // status_demand
				respPdu.Type = gosnmp.Integer
				respPdu.Value = 255
			case "1.3.6.1.4.1.2825.4.2.1.1.4.1.9.1": // status_ped_demand
				respPdu.Type = gosnmp.Integer
				respPdu.Value = 0
			case "1.3.6.1.4.1.2825.4.2.1.1.2.1.21.1": // watchdog
				respPdu.Type = gosnmp.Integer
				respPdu.Value = 1
			default:
				respPdu.Type = gosnmp.Integer
				respPdu.Value = 0
			}
			respVars = append(respVars, respPdu)
		}

	case gosnmp.SetRequest:
		for _, pdu := range packet.Variables {
			oid := strings.TrimPrefix(pdu.Name, ".")
			val := int(gosnmp.ToBigInt(pdu.Value).Int64())

			switch oid {
			case "1.3.6.1.4.1.2825.4.2.1.1.4.1.2.1": // HOLD
				if val == 0 {
					s.tl.OverrideState = "NORMAL"
					if s.onCommandLog != nil {
						s.onCommandLog(fmt.Sprintf("[Semáforo %s] UTMC2 HOLD liberado -> ciclo normal.", s.tl.ID))
					}
				} else {
					stageIdx := 0
					for j := 0; j < 16; j++ {
						if (val & (1 << j)) != 0 {
							stageIdx = j
							break
						}
					}
					s.tl.SetTargetPhase(stageIdx)
					s.tl.OverrideState = "HOLD"
					if s.onCommandLog != nil {
						s.onCommandLog(fmt.Sprintf("[Semáforo %s] UTMC2 HOLD recebido (mask=%d) -> Estágio %d forçado!", s.tl.ID, val, stageIdx))
					}
				}

			case "1.3.6.1.4.1.2825.4.2.1.1.4.1.3.1": // FORCE_OFF
				s.tl.OverrideState = "NORMAL"
				if s.onCommandLog != nil {
					s.onCommandLog(fmt.Sprintf("[Semáforo %s] UTMC2 FORCE_OFF recebido -> liberado para ciclo padrão.", s.tl.ID))
				}

			case "1.3.6.1.4.1.2825.4.2.1.1.2.1.255.1": // FLASH
				if val == 1 {
					s.tl.OverrideState = "ALERT"
				} else {
					s.tl.OverrideState = "NORMAL"
				}
				if s.onCommandLog != nil {
					s.onCommandLog(fmt.Sprintf("[Semáforo %s] UTMC2 FLASH=%d recebido.", s.tl.ID, val))
				}

			case "1.3.6.1.4.1.2825.4.2.1.1.2.1.254.1": // DARK
				if val == 1 {
					s.tl.OverrideState = "OFF"
				} else {
					s.tl.OverrideState = "NORMAL"
				}
				if s.onCommandLog != nil {
					s.onCommandLog(fmt.Sprintf("[Semáforo %s] UTMC2 DARK=%d recebido.", s.tl.ID, val))
				}
			}

			// In SNMP SetRequest responses, echo back the set PDU
			respVars = append(respVars, pdu)
		}

	default:
		return
	}

	respPacket := &gosnmp.SnmpPacket{
		Version:   packet.Version,
		Community: packet.Community,
		PDUType:   gosnmp.GetResponse,
		RequestID: packet.RequestID,
		Error:     gosnmp.NoError,
		Variables: respVars,
	}

	respBytes, err := respPacket.MarshalMsg()
	if err == nil && s.conn != nil {
		_, _ = s.conn.WriteToUDP(respBytes, remoteAddr)
	}
}

func (s *Server) handleFallbackPacket(data []byte, remoteAddr *net.UDPAddr) {
	content := string(data)

	if strings.Contains(content, "2825.4.2.1.1.4.1.2.1") || strings.Contains(content, "HOLD") {
		stageMask := 1
		if len(data) > 0 {
			lastByte := int(data[len(data)-1])
			if lastByte > 0 && lastByte < 16 {
				stageMask = lastByte
			}
		}

		stageIdx := 0
		for j := 0; j < 8; j++ {
			if (stageMask & (1 << j)) != 0 {
				stageIdx = j
				break
			}
		}

		s.tl.SetTargetPhase(stageIdx)
		s.tl.OverrideState = "HOLD"

		msg := fmt.Sprintf("[Semáforo %s] Comando UTMC2 HOLD recebido de %s -> Estágio %d forçado!",
			s.tl.ID, remoteAddr.String(), stageIdx)
		if s.onCommandLog != nil {
			s.onCommandLog(msg)
		}
	} else if strings.Contains(content, "FORCE_OFF") {
		s.tl.OverrideState = "NORMAL"
		if s.onCommandLog != nil {
			s.onCommandLog(fmt.Sprintf("[Semáforo %s] Comando UTMC2 FORCE_OFF recebido -> liberado para ciclo padrão.", s.tl.ID))
		}
	}
}

