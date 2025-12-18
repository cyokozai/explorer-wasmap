//go:build js && wasm
package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

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
			distances[i][j] = math.Inf(1)
			pheromones[i][j] = 0.0
		}
	}

	edges := []Edge{}
	
	// 座標(100x100)における最大ユークリッド距離 (ルート20000)
	const MaxEuclideanDist = 141.421356

	addEdge := func(u, v int) {
		if distances[u][v] != math.Inf(1) { return }
		
		// 実際のユークリッド距離を計算
		rawDist := math.Hypot(nodes[u].X-nodes[v].X, nodes[u].Y-nodes[v].Y)
		
		// ★変更点: 重みを0-1に正規化して設定
		normalizedWeight := rawDist / MaxEuclideanDist
		
		// 重みが0になりすぎると計算(1/dist)でバグるので極小値を保証
		if normalizedWeight < 0.0001 {
			normalizedWeight = 0.0001
		}

		distances[u][v] = normalizedWeight
		distances[v][u] = normalizedWeight
		pheromones[u][v] = InitialPheromone
		pheromones[v][u] = InitialPheromone
		
		// JSには正規化後の重みを送りますが、
		// 距離表示のためにJS側で再計算させるか、ここでrawを送る手もあります。
		// 今回は仕様通り正規化した値をWeightに入れます。
		edges = append(edges, Edge{From: u, To: v, Weight: normalizedWeight})
	}

	// グラフ生成（連結リング）
	for i := 0; i < nodeCount; i++ {
		addEdge(i, (i+1)%nodeCount)
	}
	// ショートカット生成
	extraEdges := nodeCount * 3
	for i := 0; i < extraEdges; i++ {
		u := randSource.Intn(nodeCount)
		v := randSource.Intn(nodeCount)
		if u != v { addEdge(u, v) }
	}

	return &ACO{
		Graph:      GraphData{Nodes: nodes, Edges: edges},
		Distances:  distances,
		Pheromones: pheromones,
		BestDist:   math.MaxFloat64,
		BestPath:   nil,
		Rand:       randSource,
		StartNode:  0,
		GoalNode:   nodeCount - 1,
	}
}

// Step: A地点からB地点への探索
func (aco *ACO) Step() {
	n := len(aco.Graph.Nodes)

	type AntResult struct {
		Path []int
		Dist float64
		Success bool // ゴールできたか？
	}
	antResults := make([]AntResult, AntCount)

	// 1. 全てのアリがスタートからゴールを目指す
	for k := 0; k < AntCount; k++ {
		path, success := aco.constructSolution()
		
		if !success {
			antResults[k] = AntResult{Success: false}
			continue
		}

		dist := aco.calculatePathDistance(path)
		antResults[k] = AntResult{Path: path, Dist: dist, Success: true}

		if dist < aco.BestDist {
			aco.BestDist = dist
			bestPath := make([]int, len(path))
			copy(bestPath, path)
			aco.BestPath = bestPath
			fmt.Printf("New Best Path Found! Distance: %.2f (Nodes: %d)\n", aco.BestDist, len(path))
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

	// 3. フェロモン更新（ゴールできたアリのみ！）
	for _, result := range antResults {
		if !result.Success { continue } // 失敗したアリはフェロモンを残さない
		
		deposit := Q / result.Dist
		for i := 0; i < len(result.Path)-1; i++ {
			u, v := result.Path[i], result.Path[i+1]
			aco.Pheromones[u][v] += deposit
			aco.Pheromones[v][u] += deposit
		}
	}
}

// constructSolution: スタートからゴールへの経路を探索
func (aco *ACO) constructSolution() ([]int, bool) {
	path := []int{aco.StartNode}
	visited := make([]bool, len(aco.Graph.Nodes))
	visited[aco.StartNode] = true
	
	current := aco.StartNode

	// 最大ステップ数制限（無限ループ防止）
	maxSteps := len(aco.Graph.Nodes) * 2

	for step := 0; step < maxSteps; step++ {
		// ゴール到達チェック
		if current == aco.GoalNode {
			return path, true
		}

		next := aco.selectNextCity(current, visited)
		
		if next == -1 {
			// 行き止まり
			return nil, false
		}

		path = append(path, next)
		visited[next] = true // 訪問済みにする（ループ防止）
		current = next
	}

	return nil, false // ステップオーバー
}

func (aco *ACO) selectNextCity(current int, visited []bool) int {
	n := len(aco.Graph.Nodes)
	probabilities := make([]float64, n)
	sumProb := 0.0

	// 隣接ノードのみを候補にする
	for i := 0; i < n; i++ {
		// 未訪問 かつ 接続あり
		if !visited[i] && aco.Distances[current][i] != math.Inf(1) {
			pheromone := math.Pow(aco.Pheromones[current][i], Alpha)
			heuristic := math.Pow(1.0/aco.Distances[current][i], Beta)
			prob := pheromone * heuristic
			probabilities[i] = prob
			sumProb += prob
		}
	}

	if sumProb == 0.0 { return -1 }

	r := aco.Rand.Float64() * sumProb
	cumulative := 0.0
	for i := 0; i < n; i++ {
		if !visited[i] && aco.Distances[current][i] != math.Inf(1) {
			cumulative += probabilities[i]
			if cumulative >= r { return i }
		}
	}
	// 誤差対策のフォールバック
	for i := 0; i < n; i++ {
		if !visited[i] && aco.Distances[current][i] != math.Inf(1) { return i }
	}
	return -1
}

func (aco *ACO) calculatePathDistance(path []int) float64 {
	dist := 0.0
	for i := 0; i < len(path)-1; i++ {
		dist += aco.Distances[path[i]][path[i+1]]
	}
	// TSPではないので、最後にスタートに戻る距離は足さない
	return dist
}
