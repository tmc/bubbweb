//go:build !js
// +build !js

package bubbweb

import (
	tea "github.com/charmbracelet/bubbletea"
)

var NewProgram = tea.NewProgram
