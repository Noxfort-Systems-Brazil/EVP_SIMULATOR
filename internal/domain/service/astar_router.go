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

// File: astar_router.go
// Author: Gabriel Moraes
// Date: October 2026

package service

import (
	"container/heap"

	"evp_simulator/internal/domain/entity"
	"evp_simulator/internal/domain/valueobject"
)

// AStarNode represents an item in the priority queue.
type AStarNode struct {
	EdgeID   string
	CostG    float64 // Actual cost from start
	CostF    float64 // CostG + Heuristic H
	Index    int     // Priority queue index
}

type PriorityQueue []*AStarNode

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].CostF < pq[j].CostF }
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*AStarNode)
	item.Index = n
	*pq = append(*pq, item)
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

// AStarRouter computes fastest emergency and standard routes on a RoadNetwork.
type AStarRouter struct{}

// NewAStarRouter creates a new AStarRouter instance.
func NewAStarRouter() *AStarRouter {
	return &AStarRouter{}
}

// FindRoute calculates the optimal edge sequence from startEdgeID to targetEdgeID.
func (r *AStarRouter) FindRoute(
	network *entity.RoadNetwork,
	startEdgeID string,
	targetEdgeID string,
	mode valueobject.VehicleMode,
) []string {
	if startEdgeID == targetEdgeID {
		return []string{startEdgeID}
	}

	startEdge, existsStart := network.Edges[startEdgeID]
	targetEdge, existsTarget := network.Edges[targetEdgeID]
	if !existsStart || !existsTarget {
		return nil
	}

	targetNode, existsTargetNode := network.Nodes[targetEdge.ToNode]
	var targetPos valueobject.Coordinate
	if existsTargetNode {
		targetPos = targetNode.Position
	}

	pq := &PriorityQueue{}
	heap.Init(pq)

	cameFrom := make(map[string]string)
	costG := make(map[string]float64)

	costG[startEdgeID] = 0
	startNode, _ := network.Nodes[startEdge.ToNode]
	h0 := 0.0
	if startNode != nil && existsTargetNode {
		h0 = startNode.Position.DistanceToCartesian(targetPos) / 30.0 // heuristic time in seconds
	}

	heap.Push(pq, &AStarNode{
		EdgeID: startEdgeID,
		CostG:  0,
		CostF:  h0,
	})

	visited := make(map[string]bool)

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*AStarNode)
		currID := current.EdgeID

		if currID == targetEdgeID {
			// Reconstruct path
			path := []string{currID}
			for p, ok := cameFrom[currID]; ok; p, ok = cameFrom[p] {
				path = append([]string{p}, path...)
			}
			return path
		}

		if visited[currID] {
			continue
		}
		visited[currID] = true

		currEdge := network.Edges[currID]
		if currEdge == nil {
			continue
		}

		// Explore outgoing connections
		conns := network.Connections[currID]
		for _, conn := range conns {
			nextID := conn.ToEdge
			nextEdge := network.Edges[nextID]
			if nextEdge == nil || visited[nextID] {
				continue
			}

			// Calculate edge traversal cost
			edgeSpeed := nextEdge.Speed.ToMps()
			if edgeSpeed <= 0 {
				edgeSpeed = 10.0
			}

			var effectiveSpeed float64
			var penalty float64

			if mode.IsAlert() {
				// In Alert mode: higher speed, but penalize narrow 1-lane roads (hard for cars to yield)
				effectiveSpeed = edgeSpeed * 1.4
				if nextEdge.NumLanes == 1 {
					penalty += 15.0 // +15 seconds penalty for narrow streets
				}
				// If junction has traffic light, CARINA grants preemption -> 0s signal wait!
				if conn.TrafficLightID != "" {
					penalty += 0.0
				}
			} else {
				// In Normal mode: regular speed + expected wait time on traffic lights
				effectiveSpeed = edgeSpeed
				if conn.TrafficLightID != "" {
					penalty += 20.0 // average wait time on red phase
				}
			}

			travelTime := (nextEdge.Length / effectiveSpeed) + penalty
			tentativeG := costG[currID] + travelTime

			if existingG, found := costG[nextID]; !found || tentativeG < existingG {
				cameFrom[nextID] = currID
				costG[nextID] = tentativeG

				var h float64
				if nextToNode, ok := network.Nodes[nextEdge.ToNode]; ok && existsTargetNode {
					h = nextToNode.Position.DistanceToCartesian(targetPos) / 30.0
				}

				heap.Push(pq, &AStarNode{
					EdgeID: nextID,
					CostG:  tentativeG,
					CostF:  tentativeG + h,
				})
			}
		}
	}

	// Fallback if no full path reached: return at least starting edge
	return []string{startEdgeID}
}
