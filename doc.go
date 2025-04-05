// Package bubbweb provides a WebAssembly interface for BubbleTea applications.
//
// BubblTea is a Go framework for building terminal user interfaces. With bubbweb,
// these terminal UIs can be compiled to WebAssembly and run in a browser using a
// terminal emulator like xterm.js.
//
// Basic usage:
//
//	import (
//		tea "github.com/charmbracelet/bubbletea"
//		"github.com/tmc/bubbweb"
//	)
//
//	func main() {
//		// Create your BubbleTea model as usual
//		model := yourModel()
//
//		// Use bubbweb to run the program in WebAssembly
//		prog := bubbweb.NewProgram(model, tea.WithAltScreen())
//		if _, err := prog.Run(); err != nil {
//			// Handle error
//		}
//	}
//
// When compiled to WebAssembly, the program provides JavaScript bindings
// that connect to an xterm.js terminal in the browser. The included example
// demonstrates the complete setup, including HTML and JavaScript.
//
// The bubbweb package handles input and output between the BubbleTea application
// and the browser. It exposes four JavaScript functions:
//
//   - bubbletea_write: Sends input from JavaScript to the Go program
//   - bubbletea_read: Reads output from the Go program
//   - bubbletea_resize: Sends terminal resize events to the Go program
//   - bubbletea_mouse: Sends mouse events to the Go program
//
// Mouse support is enabled by default and works with standard BubbleTea mouse handling.
// Your application will receive mouse events through the tea.MouseMsg type:
//
//	case tea.MouseMsg:
//	    switch msg.Type {
//	    case tea.MousePress:
//	        // Handle mouse press
//	    case tea.MouseRelease:
//	        // Handle mouse release
//	    case tea.MouseMotion:
//	        // Handle mouse movement
//	    case tea.MouseWheelUp, tea.MouseWheelDown:
//	        // Handle scrolling
//	    }
//
// These JavaScript functions are called by the JavaScript code in the HTML page.
//
// To build a WebAssembly application using bubbweb:
//
//  1. Create a Go program that uses bubbweb
//  2. Compile it with GOOS=js GOARCH=wasm
//  3. Copy wasm_exec.js from Go distribution
//  4. Create HTML with xterm.js that loads and communicates with the WebAssembly module
//  5. Configure xterm.js to forward mouse events to the WebAssembly module
//
// See the example directory for a complete implementation.
package bubbweb