# Explorer WASMap

1. Copy `main_wasm.js`

    ```bash
    curl -o wasm_exec.js https://raw.githubusercontent.com/golang/go/ad91f5d241f3b8e85dc866d0087c3f13af96ef33/lib/wasm/wasm_exec.js
    ```

1. WASM Compile

    ```bash
    GOOS=js GOARCH=wasm go build -o main.wasm .
    ```
