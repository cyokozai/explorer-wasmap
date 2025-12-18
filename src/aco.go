//go:build js && wasm
package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// NewACO: ランダムなネットワークグラフを生成して初期化
func NewACO(nodeCount int) *ACO {
	randSource := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 1. ノード生成
	nodes := make([]Node, nodeCount)
	for i := 0; i < nodeCount; i++ {
		nodes[i] = Node{
			ID: i,
			X:  randSource.Float64() * 100,
			Y:  randSource.Float64() * 100,
		}
	}

	// 2. 行列初期化
	distances := make([][]float64, nodeCount)
	pheromones := make([][]float64, nodeCount)

	for i := 0; i < nodeCount; i++ {
		distances[i] = make([]float64, nodeCount)
		pheromones[i] = make([]float64, nodeCount)

		for j := 0; j < nodeCount; j++ {
			distances[i][j] = math.Inf(1) // 初期状態は接続なし
			pheromones[i][j] = 0.0
		}
	}

	edges := []Edge{}

	// エッジ追加ヘルパー関数
	addEdge := func(u, v int) {
		// 既に接続済みなら何もしない
		if distances[u][v] != math.Inf(1) {
			return
		}

		dist := math.Hypot(nodes[u].X-nodes[v].X, nodes[u].Y-nodes[v].Y)
		trafficFactor := 1.0 + randSource.Float64()*2.0
		weight := dist * trafficFactor

		distances[u][v] = weight
		distances[v][u] = weight
		pheromones[u][v] = InitialPheromone
		pheromones[v][u] = InitialPheromone

		edges = append(edges, Edge{From: u, To: v, Weight: weight})
	}

	// リング状に接続（孤立防止）
	for i := 0; i < nodeCount; i++ {
		addEdge(i, (i+1)%nodeCount)
	}

	// ランダムなショートカットを追加
	extraEdges := nodeCount * 2
	for i := 0; i < extraEdges; i++ {
		u := randSource.Intn(nodeCount)
		v := randSource.Intn(nodeCount)
		if u != v {
			addEdge(u, v)
		}
	}

	return &ACO{
		Graph:      GraphData{Nodes: nodes, Edges: edges},
		Distances:  distances,
		Pheromones: pheromones,
		BestDist:   math.MaxFloat64,
		BestPath:   nil,
		Rand:       randSource,
	}
}

// Step: 1世代分のシミュレーションを実行
func (aco *ACO) Step() {
	n := len(aco.Graph.Nodes)

	type AntResult struct {
		Path []int
		Dist float64
	}
	antResults := make([]AntResult, AntCount)

	// 1. 全てのアリが解を構築
	for k := 0; k < AntCount; k++ {
		path := aco.constructSolution(n)
		
		// 経路が見つからなかった場合（袋小路など）のガード
		if len(path) != n {
			antResults[k] = AntResult{Path: nil, Dist: math.Inf(1)}
			continue
		}

		dist := aco.calculatePathDistance(path)
		antResults[k] = AntResult{Path: path, Dist: dist}

		if dist < aco.BestDist {
			aco.BestDist = dist
			bestPath := make([]int, len(path))
			copy(bestPath, path)
			aco.BestPath = bestPath
			fmt.Printf("New Best Distance: %.2f\n", aco.BestDist)
		}
	}

	// 2. フェロモン蒸発
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if aco.Distances[i][j] != math.Inf(1) {
				aco.Pheromones[i][j] *= (1.0 - Evaporation)
			}
		}
	}

	// 3. フェロモン更新
	for _, result := range antResults {
		if result.Path == nil { continue }
		
		deposit := Q / result.Dist
		for i := 0; i < n-1; i++ {
			u, v := result.Path[i], result.Path[i+1]
			aco.Pheromones[u][v] += deposit
			aco.Pheromones[v][u] += deposit
		}
		// 始点に戻るエッジ
		u, v := result.Path[n-1], result.Path[0]
		aco.Pheromones[u][v] += deposit
		aco.Pheromones[v][u] += deposit
	}
}

func (aco *ACO) constructSolution(numNodes int) []int {
	path := make([]int, 0, numNodes)
	visited := make([]bool, numNodes)

	current := aco.Rand.Intn(numNodes)
	path = append(path, current)
	visited[current] = true

	for len(path) < numNodes {
		next := aco.selectNextCity(current, visited)
		if next == -1 {
			// 行き止まり（通常リング構造なので起きないはずだが安全のため）
			break
		}
		path = append(path, next)
		visited[next] = true
		current = next
	}
	return path
}

func (aco *ACO) selectNextCity(current int, visited []bool) int {
	n := len(aco.Graph.Nodes)
	probabilities := make([]float64, n)
	sumProb := 0.0

	for i := 0; i < n; i++ {
		// 未訪問 かつ エッジが存在する場合のみ
		if !visited[i] && aco.Distances[current][i] != math.Inf(1) {
			pheromone := math.Pow(aco.Pheromones[current][i], Alpha)
			heuristic := math.Pow(1.0/aco.Distances[current][i], Beta)
			prob := pheromone * heuristic
			probabilities[i] = prob
			sumProb += prob
		}
	}

	// 移動可能な先がない場合
	if sumProb == 0.0 {
		return -1
	}

	r := aco.Rand.Float64() * sumProb
	cumulative := 0.0
	for i := 0; i < n; i++ {
		if !visited[i] && aco.Distances[current][i] != math.Inf(1) {
			cumulative += probabilities[i]
			if cumulative >= r {
				return i
			}
		}
	}

	// 浮動小数点誤差対策
	for i := 0; i < n; i++ {
		if !visited[i] && aco.Distances[current][i] != math.Inf(1) {
			return i
		}
	}
	return -1
}

func (aco *ACO) calculatePathDistance(path []int) float64 {
	dist := 0.0
	for i := 0; i < len(path)-1; i++ {
		dist += aco.Distances[path[i]][path[i+1]]
	}
	dist += aco.Distances[path[len(path)-1]][path[0]]
	return dist
}
