package bubbweb

import (
	"bytes"
	"fmt"
	"syscall/js"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// MinReadBuffer is a custom buffer for handling bubbletea's input expectations in WASM
type MinReadBuffer struct {
	buf *bytes.Buffer
}

// Read from the buffer, blocking until data is available
func (b *MinReadBuffer) Read(p []byte) (n int, err error) {
	for b.buf.Len() == 0 {
		time.Sleep(100 * time.Millisecond)
	}
	return b.buf.Read(p)
}

// Write to the buffer
func (b *MinReadBuffer) Write(p []byte) (n int, err error) {
	return b.buf.Write(p)
}

// Program represents a BubbleTea program configured for WASM
type Program struct {
	teaProgram *tea.Program
}

// NewProgram creates a new BubbleTea program configured for WASM
func NewProgram(model tea.Model, options ...tea.ProgramOption) *Program {
	fromJs := &MinReadBuffer{buf: bytes.NewBuffer(nil)}
	fromGo := bytes.NewBuffer(nil)

	// Combine default options with user-provided options
	defaultOptions := []tea.ProgramOption{
		tea.WithInput(fromJs),
		tea.WithOutput(fromGo),
	}
	allOptions := append(defaultOptions, options...)

	prog := tea.NewProgram(model, allOptions...)

	// Register write function in WASM
	js.Global().Set("bubbletea_write", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fromJs.Write([]byte(args[0].String()))
		fmt.Println("Wrote to Go:", args[0].String())
		return nil
	}))

	// Register read function in WASM
	js.Global().Set("bubbletea_read", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		b := make([]byte, fromGo.Len())
		_, _ = fromGo.Read(b)
		fromGo.Reset()
		return string(b)
	}))

	// Register resize function in WASM
	js.Global().Set("bubbletea_resize", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		width := args[0].Int()
		height := args[1].Int()
		prog.Send(tea.WindowSizeMsg{Width: width, Height: height})
		return nil
	}))

	return &Program{teaProgram: prog}
}

// Run starts the BubbleTea program
func (p *Program) Run() (tea.Model, error) {
	fmt.Println("Starting program...")
	return p.teaProgram.Run()
}