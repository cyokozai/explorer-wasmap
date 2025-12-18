//go:build js && wasm
package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
)

var globalACO *ACO

func main() {
	js.Global().Set("initACO", js.FuncOf(initACOWrapper))
	js.Global().Set("getGraph", js.FuncOf(getGraphWrapper))
	js.Global().Set("stepACO", js.FuncOf(stepWrapper))

	fmt.Println("WASM Initialized")
	select {}
}

// initACO(numCities)
func initACOWrapper(this js.Value, args []js.Value) interface{} {
	numCities := 20
	if len(args) > 0 {
		numCities = args[0].Int()
	}
	if numCities < 2 {
		numCities = 2
	}

	globalACO = NewACO(numCities)
	fmt.Printf("Initialized ACO with %d nodes\n", numCities)

	return nil
}

// getGraph() -> JSON string
func getGraphWrapper(this js.Value, args []js.Value) interface{} {
	if globalACO == nil {
		fmt.Println("Error: globalACO is null")

		return "{}"
	}
	jsonData, err := json.Marshal(globalACO.Graph)
	if err != nil {
		fmt.Println("Error marshalling graph:", err)

		return "{}"
	}

	return string(jsonData)
}

// stepACO() -> JSON string {bestDist, bestPath}
func stepWrapper(this js.Value, args []js.Value) interface{} {
	if globalACO == nil {
		return "{}"
	}
	
	globalACO.Step()

	result := struct {
		BestDist float64 `json:"bestDist"`
		BestPath []int   `json:"bestPath"`
	}{
		BestDist: globalACO.BestDist,
		BestPath: globalACO.BestPath,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return "{}"
	}

	return string(jsonData)
}
