//go:build js && wasm
package main

import (
	"math/rand"
)

const (
	AntCount         = 20
	Alpha            = 1.0
	Beta             = 5.0
	Evaporation      = 0.5
	Q                = 100.0
	InitialPheromone = 1.0
)

type Node struct {
	ID int     `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
}

type Edge struct {
	From   int     `json:"from"`
	To     int     `json:"to"`
	Weight float64 `json:"weight"`
}

type GraphData struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type ACO struct {
	Graph      GraphData
	Distances  [][]float64
	Pheromones [][]float64
	BestDist   float64
	BestPath   []int
	Rand       *rand.Rand
	StartNode  int
	GoalNode   int
}
