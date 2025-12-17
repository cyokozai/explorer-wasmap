package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// パラメータ設定
const (
	AntCount      = 20    // アリの数
	Alpha         = 1.0   // フェロモンの重要度 (大きいと過去の経験重視)
	Beta          = 5.0   // ヒューリスティック(距離の近さ)の重要度 (大きいと近くの都市重視)
	Evaporation   = 0.5   // フェロモンの蒸発率 (0.0~1.0)
	Q             = 100.0 // フェロモン更新定数
	Iterations    = 100   // ループ回数
	InitialPheromone = 1.0 // フェロモンの初期値
)

// City は都市の座標を表します
type City struct {
	X, Y float64
}

// ACO はアルゴリズム全体の状態を管理します
type ACO struct {
	Cities     []City
	Distances  [][]float64 // 距離行列
	Pheromones [][]float64 // フェロモン行列
	BestDist   float64     // これまでの最短距離
	BestPath   []int       // これまでの最短経路（都市インデックス順）
	Rand       *rand.Rand  // 乱数生成器
}

// NewACO はACOのインスタンスを初期化します
func NewACO(cities []City) *ACO {
	n := len(cities)
	distances := make([][]float64, n)
	pheromones := make([][]float64, n)

	// 距離行列とフェロモン行列の初期化
	for i := 0; i < n; i++ {
		distances[i] = make([]float64, n)
		pheromones[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if i != j {
				// ユークリッド距離の計算
				dist := math.Hypot(cities[i].X-cities[j].X, cities[i].Y-cities[j].Y)
				distances[i][j] = dist
				pheromones[i][j] = InitialPheromone
			}
		}
	}

	return &ACO{
		Cities:     cities,
		Distances:  distances,
		Pheromones: pheromones,
		BestDist:   math.MaxFloat64,
		BestPath:   nil,
		Rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Run は指定回数のイテレーションを実行します
func (aco *ACO) Run(iterations int) {
	for iter := 0; iter < iterations; iter++ {
		aco.Step(iter)
	}
}

// Step は1イテレーション分（全アリの移動とフェロモン更新）を実行します
// ※WASM化する際は、JS側からこの関数をループで呼ぶ形になります
func (aco *ACO) Step(iterCount int) {
	n := len(aco.Cities)
	
	// 各アリが構築したパスとその距離を保持するリスト
	type AntResult struct {
		Path []int
		Dist float64
	}
	antResults := make([]AntResult, AntCount)

	// 1. 全てのアリがソリューションを構築
	for k := 0; k < AntCount; k++ {
		path := aco.constructSolution(n)
		dist := aco.calculatePathDistance(path)
		antResults[k] = AntResult{Path: path, Dist: dist}

		// ベスト解の更新
		if dist < aco.BestDist {
			aco.BestDist = dist
			// pathのスライスをコピーして保存
			bestPath := make([]int, len(path))
			copy(bestPath, path)
			aco.BestPath = bestPath
			fmt.Printf("Iter %d: New Best Distance found: %.2f\n", iterCount, aco.BestDist)
		}
	}

	// 2. フェロモンの蒸発
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			aco.Pheromones[i][j] *= (1.0 - Evaporation)
		}
	}

	// 3. フェロモンの追加 (距離が短いほど多くのフェロモンを残す)
	for _, result := range antResults {
		deposit := Q / result.Dist
		for i := 0; i < n-1; i++ {
			u, v := result.Path[i], result.Path[i+1]
			aco.Pheromones[u][v] += deposit
			aco.Pheromones[v][u] += deposit // 無向グラフとして扱う
		}
		// 始点に戻るエッジも更新
		u, v := result.Path[n-1], result.Path[0]
		aco.Pheromones[u][v] += deposit
		aco.Pheromones[v][u] += deposit
	}
}

// constructSolution は1匹のアリが都市を巡回するパスを作成します
func (aco *ACO) constructSolution(numCities int) []int {
	path := make([]int, 0, numCities)
	visited := make([]bool, numCities)

	// ランダムな都市からスタート
	current := aco.Rand.Intn(numCities)
	path = append(path, current)
	visited[current] = true

	for len(path) < numCities {
		next := aco.selectNextCity(current, visited)
		path = append(path, next)
		visited[next] = true
		current = next
	}
	return path
}

// selectNextCity は確率的に次の都市を選びます
func (aco *ACO) selectNextCity(current int, visited []bool) int {
	n := len(aco.Cities)
	probabilities := make([]float64, n)
	sumProb := 0.0

	// 確率の計算: (フェロモン^Alpha) * ((1/距離)^Beta)
	for i := 0; i < n; i++ {
		if !visited[i] {
			pheromone := math.Pow(aco.Pheromones[current][i], Alpha)
			heuristic := math.Pow(1.0/aco.Distances[current][i], Beta)
			prob := pheromone * heuristic
			probabilities[i] = prob
			sumProb += prob
		}
	}

	// ルーレット選択
	r := aco.Rand.Float64() * sumProb
	cumulative := 0.0
	for i := 0; i < n; i++ {
		if !visited[i] {
			cumulative += probabilities[i]
			if cumulative >= r {
				return i
			}
		}
	}
	
	// 万が一計算誤差で選ばれなかった場合、未訪問の最初の都市を返す
	for i := 0; i < n; i++ {
		if !visited[i] {
			return i
		}
	}
	return -1 // ここには来ないはず
}

// calculatePathDistance はパスの総距離を計算します
func (aco *ACO) calculatePathDistance(path []int) float64 {
	dist := 0.0
	for i := 0; i < len(path)-1; i++ {
		dist += aco.Distances[path[i]][path[i+1]]
	}
	// スタート地点に戻る距離
	dist += aco.Distances[path[len(path)-1]][path[0]]
	return dist
}

func main() {
	// ダミーの地図データ（ランダムな20都市）を作成
	numCities := 20
	cities := make([]City, numCities)
	rand.Seed(time.Now().UnixNano()) // メインの乱数シード
	fmt.Println("Generating cities...")
	for i := 0; i < numCities; i++ {
		cities[i] = City{
			X: rand.Float64() * 100,
			Y: rand.Float64() * 100,
		}
		fmt.Printf("City %d: (%.2f, %.2f)\n", i, cities[i].X, cities[i].Y)
	}

	// ACOの初期化と実行
	fmt.Println("\nStarting ACO...")
	aco := NewACO(cities)
	aco.Run(Iterations)

	fmt.Printf("\nFinal Best Distance: %.2f\n", aco.BestDist)
	fmt.Printf("Best Path: %v\n", aco.BestPath)
}