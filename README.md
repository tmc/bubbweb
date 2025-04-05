# bubbweb

[![Go Reference](https://pkg.go.dev/badge/github.com/tmc/bubbweb.svg)](https://pkg.go.dev/github.com/tmc/bubbweb)

BubbWeb is a library for running [Bubbletea](https://github.com/charmbracelet/bubbletea) terminal user interfaces in WebAssembly.

## Live Demo

Check out the [live demo](https://tmc.github.io/bubbweb/example/) to see BubbWeb in action.

<p align="center">
  <img src="./.github/screenshot.png" alt="Screenshot of the example" />
</p>

## Features

- Run Bubbletea TUIs directly in the browser
- Uses xterm.js for terminal emulation
- Handles input/output between JavaScript and Go
- Full mouse support (clicks, movement, wheel scrolling)
- Manages terminal resize events
- Includes ETag-based caching for efficient updates

## Usage

```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/tmc/bubbweb"
)

func main() {
    // Create your BubbleTea model as usual
    model := yourModel()

    // Use bubbweb to run the program in WebAssembly
    prog := bubbweb.NewProgram(model, tea.WithAltScreen())
    if _, err := prog.Run(); err != nil {
        // Handle error
    }
}
```

## Building a WebAssembly Application

```shell
# Build everything with go generate
go generate

# For local testing with auto-reload
cd example
go run github.com/tmc/serve@latest # or any other HTTP server
```

Then open http://localhost:8080 in your browser.

## Example

See the `example` directory for a complete example including:

- A multi-pane text editor built with Bubbletea
- HTML/JavaScript integration with xterm.js
- Update notification system
- ETag-based caching for efficient updates

## Deployment

This project can be easily deployed on GitHub Pages:

1. Push the `example` directory to your GitHub repository
2. Go to repository Settings → Pages
3. Set the source to the branch containing your `example` directory
4. Configure the root directory to `/` or `/example` depending on your repository structure
5. Save the settings and GitHub Pages will deploy your application

Your Bubbletea WebAssembly application will be available at `https://[username].github.io/[repository]/example`

## Implementation Notes

BubbWeb handles several challenges of running Bubbletea in WebAssembly:

1. Provides custom I/O implementation for WebAssembly
2. Exposes JavaScript functions for browser communication:
   - `bubbletea_write`: Sends input from JavaScript to the Go program
   - `bubbletea_read`: Reads output from the Go program
   - `bubbletea_resize`: Sends terminal resize events to the Go program
   - `bubbletea_mouse`: Sends mouse events to the Go program
3. Enables full mouse support with standard BubbleTea event handling
4. Uses replacements for packages that don't fully support WebAssembly

### Mouse Support

Mouse events are translated from browser events to BubbleTea's mouse event system:

```go
case tea.MouseMsg:
    switch msg.Type {
    case tea.MousePress:
        // Handle mouse press at (msg.X, msg.Y)
    case tea.MouseRelease:
        // Handle mouse release
    case tea.MouseMotion:
        // Handle mouse movement
    case tea.MouseWheelUp, tea.MouseWheelDown:
        // Handle scrolling
    }
}
```

The coordinates are automatically converted from pixel coordinates to terminal cell coordinates.

## License

MIT