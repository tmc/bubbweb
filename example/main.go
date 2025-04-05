package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tmc/bubbweb"
)

const (
	initialInputs = 2
	maxInputs     = 6
	minInputs     = 1
	helpHeight    = 5
)

var (
	// Define adaptive colors for light/dark mode support
	cursorColor               = lipgloss.AdaptiveColor{Light: "#d787ff", Dark: "#d787ff"} // Purple
	cursorLineBackgroundColor = lipgloss.AdaptiveColor{Light: "#5f87ff", Dark: "#5f5f87"} // Blue/Dark Blue
	cursorLineForegroundColor = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#e4e4e4"} // Black/White
	placeholderColor          = lipgloss.AdaptiveColor{Light: "#808080", Dark: "#a8a8a8"} // Gray - increased contrast for dark mode
	endOfBufferColor          = lipgloss.AdaptiveColor{Light: "#bcbcbc", Dark: "#3a3a3a"} // Light Gray/Dark Gray
	focusedPlaceholderColor   = lipgloss.AdaptiveColor{Light: "#8787ff", Dark: "#afbfff"} // Blue - brighter in dark mode
	borderColor               = lipgloss.AdaptiveColor{Light: "#a8a8a8", Dark: "#787878"} // Medium Gray - increased contrast

	// Apply the adaptive colors to styles
	cursorStyle = lipgloss.NewStyle().Foreground(cursorColor)

	cursorLineStyle = lipgloss.NewStyle().
			Background(cursorLineBackgroundColor).
			Foreground(cursorLineForegroundColor)

	placeholderStyle = lipgloss.NewStyle().
				Foreground(placeholderColor)

	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(endOfBufferColor)

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(focusedPlaceholderColor)

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor)

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder())
)

type keymap = struct {
	next, prev, add, remove, quit key.Binding
}

func newTextarea() textarea.Model {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = "Type something!"
	t.ShowLineNumbers = true
	t.Cursor.Style = cursorStyle
	t.FocusedStyle.Placeholder = focusedPlaceholderStyle
	t.BlurredStyle.Placeholder = placeholderStyle
	t.FocusedStyle.CursorLine = cursorLineStyle
	t.FocusedStyle.Base = focusedBorderStyle
	t.BlurredStyle.Base = blurredBorderStyle
	t.FocusedStyle.EndOfBuffer = endOfBufferStyle
	t.BlurredStyle.EndOfBuffer = endOfBufferStyle
	t.KeyMap.DeleteWordBackward.SetEnabled(false)
	t.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	t.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	t.Blur()
	return t
}

type model struct {
	width         int
	height        int
	keymap        keymap
	help          help.Model
	inputs        []textarea.Model
	focus         int
	mousePosition string // Store last mouse position for display
	mouseEvent    string // Store last mouse event type
}

func newModel() model {
	m := model{
		inputs: make([]textarea.Model, initialInputs),
		help:   help.New(),
		keymap: keymap{
			next: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next"),
			),
			prev: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "prev"),
			),
			add: key.NewBinding(
				key.WithKeys("ctrl+n"),
				key.WithHelp("ctrl+n", "add an editor"),
			),
			remove: key.NewBinding(
				key.WithKeys("ctrl+w"),
				key.WithHelp("ctrl+w", "remove an editor"),
			),
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
	}
	for i := 0; i < initialInputs; i++ {
		m.inputs[i] = newTextarea()
	}
	m.inputs[m.focus].Focus()
	m.updateKeybindings()
	return m
}

func (m model) Init() tea.Cmd {
	// Request an initial window size on startup
	// This ensures we have a proper width for the top bar
	return tea.Batch(
		textarea.Blink,
		tea.EnterAltScreen,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			for i := range m.inputs {
				m.inputs[i].Blur()
			}
			return m, tea.Quit
		case key.Matches(msg, m.keymap.next):
			m.inputs[m.focus].Blur()
			m.focus++
			if m.focus > len(m.inputs)-1 {
				m.focus = 0
			}
			cmd := m.inputs[m.focus].Focus()
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.keymap.prev):
			m.inputs[m.focus].Blur()
			m.focus--
			if m.focus < 0 {
				m.focus = len(m.inputs) - 1
			}
			cmd := m.inputs[m.focus].Focus()
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.keymap.add):
			m.inputs = append(m.inputs, newTextarea())
		case key.Matches(msg, m.keymap.remove):
			m.inputs = m.inputs[:len(m.inputs)-1]
			if m.focus > len(m.inputs)-1 {
				m.focus = len(m.inputs) - 1
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case tea.MouseMsg:
		m.mousePosition = fmt.Sprintf("(%d,%d)", msg.X, msg.Y)
		m.mouseEvent = fmt.Sprint(msg)

		// Handle mouse clicks to change focus and set cursor position
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Determine which editor was clicked based on X position
			if len(m.inputs) > 0 {
				editorWidth := m.width / len(m.inputs)
				clickedEditor := msg.X / editorWidth

				// Only change focus if clicking on a different editor
				if clickedEditor < len(m.inputs) && clickedEditor != m.focus {
					// Blur the currently focused editor
					m.inputs[m.focus].Blur()
					m.focus = clickedEditor
					cmd := m.inputs[m.focus].Focus()
					cmds = append(cmds, cmd)
				}

				// Calculate the relative position within the editor
				relativeX := msg.X % editorWidth

				// Set cursor position based on clicked position
				if m.inputs[m.focus].Value() != "" {
					rows := strings.Split(m.inputs[m.focus].Value(), "\n")

					// First, find which line was clicked
					clickedLine := msg.Y
					if clickedLine >= len(rows) {
						clickedLine = len(rows) - 1
					}
					if clickedLine < 0 {
						clickedLine = 0
					}

					// Calculate cursor position
					// Get the position at the start of the clicked line
					cursorPos := 0
					for i := 0; i < clickedLine; i++ {
						cursorPos += len(rows[i]) + 1 // +1 for newline
					}

					// Add the horizontal position (constrained by line length)
					lineLength := len(rows[clickedLine])
					charPos := relativeX / 1 // Assuming each character is ~1 cell wide
					if charPos > lineLength {
						charPos = lineLength
					}

					cursorPos += charPos
					m.inputs[m.focus].SetCursor(cursorPos)
				}
			}
		}
	}

	m.updateKeybindings()
	m.sizeInputs()

	// Update all textareas
	for i := range m.inputs {
		newModel, cmd := m.inputs[i].Update(msg)
		m.inputs[i] = newModel
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) sizeInputs() {
	if m.width <= 0 || len(m.inputs) == 0 {
		return
	}

	// Calculate the width of each editor
	editorWidth := m.width / len(m.inputs)

	// Ensure we're not wasting any horizontal space
	remainingWidth := m.width - (editorWidth * len(m.inputs))

	for i := range m.inputs {
		// Give extra width to first editor if there's remaining space
		width := editorWidth
		if i == 0 && remainingWidth > 0 {
			width += remainingWidth
		}

		m.inputs[i].SetWidth(width)

		// Adjust editor height to account for title bar (always shown)
		heightAdjustment := helpHeight + 1 // +1 for title bar

		m.inputs[i].SetHeight(m.height - heightAdjustment)
	}
}

func (m *model) updateKeybindings() {
	m.keymap.add.SetEnabled(len(m.inputs) < maxInputs)
	m.keymap.remove.SetEnabled(len(m.inputs) > minInputs)
}

func (m model) View() string {
	var topBar string

	// Create a top bar if we have a valid width

	titleBar := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render("bubbweb - example editor")

	// Set the top bar with a full-width title
	topBar = titleBar

	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.next,
		m.keymap.prev,
		m.keymap.add,
		m.keymap.remove,
		m.keymap.quit,
	})

	var views []string
	for i := range m.inputs {
		views = append(views, m.inputs[i].View())
	}

	// Create a mouse information status line
	mouseInfo := ""
	// Define adaptive colors for the mouse status display
	mouseHeaderColor := lipgloss.AdaptiveColor{Light: "#5fafaf", Dark: "#5f8787"}     // Teal
	mouseTextColor := lipgloss.AdaptiveColor{Light: "#5fafaf", Dark: "#5f8787"}       // Teal
	mouseBackgroundColor := lipgloss.AdaptiveColor{Light: "#f0f0f0", Dark: "#303030"} // Light/Dark gray

	mouseStyle := lipgloss.NewStyle().
		Foreground(mouseHeaderColor).
		Background(mouseBackgroundColor).
		Padding(0, 1)

	mouseInfo = fmt.Sprintf("\n%s %s",
		mouseStyle.Render("Mouse:"),
		lipgloss.NewStyle().Foreground(mouseTextColor).Render(
			fmt.Sprintf("%s at %s", m.mouseEvent, m.mousePosition),
		),
	)

	// Only add the top bar and a newline if we have a valid top bar
	viewContent := ""
	if topBar != "" {
		viewContent = topBar + "\n"
	}

	viewContent += lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n" + help + mouseInfo
	return viewContent
}

func main() {
	// Enable both mouse cell motion and all motion for better mouse interactions
	prog := bubbweb.NewProgram(newModel(),
		tea.WithMouseAllMotion(),  // Track all mouse motion
		tea.WithMouseCellMotion()) // Track cell-based mouse motion

	if _, err := prog.Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
