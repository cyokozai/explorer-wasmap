//go:build js && wasm
package main

import (
	"math/rand"
)


const (
	AntCount         = 20    // アリの数
	Alpha            = 1.0   // フェロモンの重要度
	Beta             = 5.0   // ヒューリスティックの重要度
	Evaporation      = 0.5   // フェロモンの蒸発率
	Q                = 100.0 // フェロモン更新定数
	InitialPheromone = 1.0   // フェロモンの初期値
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
	Distances  [][]float64 // 距離行列 (接続なしは Inf)
	Pheromones [][]float64 // フェロモン行列
	BestDist   float64     // これまでの最短距離
	BestPath   []int       // これまでの最短経路
	Rand       *rand.Rand  // 乱数生成器
}
