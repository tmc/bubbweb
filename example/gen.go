//go:generate go mod tidy
//go:generate go mod edit -dropreplace=github.com/charmbracelet/bubbletea
//go:generate go mod edit -dropreplace=github.com/atotto/clipboard

//go:generate go mod tidy
//go:generate go mod edit -replace=github.com/charmbracelet/bubbletea=github.com/tmc/bubbletea@wasm

//go:generate go mod tidy
//go:generate go mod edit -replace=github.com/atotto/clipboard=github.com/tmc/clipboard@wasm
//go:generate go mod tidy

//go:generate cp "$GOROOT/lib/wasm/wasm_exec.js" ./wasm_exec.js
//go:generate env GOOS=js GOARCH=wasm go build -o bubbletea.wasm .

//go:generate echo To launch:
//go:generate echo go run github.com/tmc/serve@latest
//go:generate echo open http://localhost:8080
package main
