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

	// Hover effect colors
	hoveredLineBackgroundColor = lipgloss.AdaptiveColor{Light: "#e0e8ff", Dark: "#2d2d5f"} // Light blue / Dark blue-purple
	hoveredCharColor           = lipgloss.AdaptiveColor{Light: "#0000ff", Dark: "#8888ff"} // Blue for hovered character

	// Apply the adaptive colors to styles
	cursorStyle = lipgloss.NewStyle().Foreground(cursorColor)

	cursorLineStyle = lipgloss.NewStyle().
			Background(cursorLineBackgroundColor).
			Foreground(cursorLineForegroundColor)

	// Hover styles - just background color for the line (no italic to keep it clean)
	hoveredLineStyle = lipgloss.NewStyle().
				Background(hoveredLineBackgroundColor)

	// Bold style for the hovered character in status line
	hoveredCharStyle = lipgloss.NewStyle().
				Foreground(hoveredCharColor).
				Bold(true)

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
	hoveredLine   int    // Track which line the mouse is hovering over
	hoveredChar   int    // Track which character the mouse is hovering over
	hoveredEditor int    // Track which editor the mouse is hovering over
}

func newModel() model {
	m := model{
		inputs:        make([]textarea.Model, initialInputs),
		help:          help.New(),
		hoveredLine:   -1, // Initialize to -1 to indicate no line is hovered
		hoveredChar:   -1, // Initialize to -1 to indicate no character is hovered
		hoveredEditor: -1, // Initialize to -1 to indicate no editor is hovered
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

		// Track hover state for highlighting
		if len(m.inputs) > 0 {
			editorWidth := m.width / len(m.inputs)
			hoveredEditor := msg.X / editorWidth

			// Reset hover state first
			m.hoveredEditor = -1
			m.hoveredLine = -1
			m.hoveredChar = -1

			// Update hovered editor if valid
			if hoveredEditor >= 0 && hoveredEditor < len(m.inputs) {
				m.hoveredEditor = hoveredEditor

				// Get relative position within the editor
				relativeX := msg.X % editorWidth

				// Get the editor content
				ta := m.inputs[hoveredEditor]
				value := ta.Value()

				if value != "" {
					// Get current view information
					info := ta.LineInfo()
					rows := strings.Split(value, "\n")

					// Account for top bar (1 line) and editor borders (1 line)
					topBarOffset := 1 // The title bar at the top
					borderOffset := 1 // The top border of the textarea

					// Calculate which line is hovered, accounting for offsets
					hoveredLine := msg.Y - (topBarOffset + borderOffset) + info.RowOffset

					// Bounds checking for line
					if hoveredLine >= 0 && hoveredLine < len(rows) {
						m.hoveredLine = hoveredLine

						// For character position we need to account for line numbers if shown
						lineNumberOffset := 0
						if ta.ShowLineNumbers {
							// Approximate offset for line numbers (varies with number size)
							lineNumberOffset = 4 // line number + space + vertical bar + space
						}

						// Calculate character position, accounting for line number column if present
						charPos := relativeX - lineNumberOffset
						if charPos >= 0 && charPos < len(rows[hoveredLine]) {
							m.hoveredChar = charPos
						}
					}
				}
			}
		}

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

				// Handle cursor positioning based on click
				ta := m.inputs[m.focus]
				value := ta.Value()

				if value != "" {
					// Get current view information
					info := ta.LineInfo()
					rows := strings.Split(value, "\n")

					// Account for top bar and editor borders
					topBarOffset := 1 // The title bar at the top
					borderOffset := 1 // The top border of the textarea

					// Consider scrolling and UI offsets
					clickedLine := msg.Y - (topBarOffset + borderOffset) + info.RowOffset

					// Bounds checking
					if clickedLine >= len(rows) {
						clickedLine = len(rows) - 1
					}
					if clickedLine < 0 {
						clickedLine = 0
					}

					// Calculate cursor position at start of clicked line
					cursorPos := 0
					for i := 0; i < clickedLine; i++ {
						cursorPos += len(rows[i]) + 1 // +1 for newline
					}

					// Account for line numbers if shown
					lineNumberOffset := 0
					if ta.ShowLineNumbers {
						lineNumberOffset = 4 // Approximate width of line numbers
					}

					// Add horizontal position within the line
					lineLength := len(rows[clickedLine])
					charPos := relativeX - lineNumberOffset
					if charPos < 0 {
						charPos = 0
					}
					if charPos > lineLength {
						charPos = lineLength
					}

					cursorPos += charPos
					ta.SetCursor(cursorPos)
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

// Custom rendering function to support highlighting hovered lines and characters
func (m model) renderTextArea(ta textarea.Model, isHovered bool, hoveredLine, hoveredChar int) string {
	// If not the hovered editor, just return the normal view
	if !isHovered || hoveredLine < 0 {
		return ta.View()
	}

	// For the hovered editor, we need to do custom rendering to highlight the hovered line/char
	value := ta.Value()
	if value == "" {
		return ta.View()
	}

	// Get the default view first, which we'll modify to add highlighting
	defaultView := ta.View()

	// Split the view into lines
	viewLines := strings.Split(defaultView, "\n")

	// Get view information to account for scrolling
	info := ta.LineInfo()
	viewportStartLine := info.RowOffset

	// The line in the rendered view will be different from the actual line number
	// because of viewport scrolling and borders. Calculate the visual line in the rendered view.

	// Get border information from the view structure
	// By examining the view lines, we can determine where the content starts
	// Usually the first line is the top border of the textarea
	borderOffset := 1

	// Look for the line that indicates content has begun (after top border)
	for i, line := range viewLines {
		if strings.Contains(line, "│") && i > 0 {
			// We found the first content line
			borderOffset = i
			break
		}
	}

	visualHoveredLine := hoveredLine - viewportStartLine + borderOffset

	// Bounds check - make sure the line is visible
	if visualHoveredLine < 0 || visualHoveredLine >= len(viewLines) {
		return defaultView // Line not in viewport
	}

	// Get the line that needs highlighting
	lineContent := viewLines[visualHoveredLine]

	// For character highlighting, we need to determine if the character is visible in the view
	// This is challenging with ANSI codes, but we can try a simple approach
	if hoveredChar >= 0 {
		// We need to get the actual text content of the row (without ANSI codes)
		// Since we can't easily strip ANSI codes, we'll use a more direct approach:
		// Go through the original text and get the row we need
		contentLines := strings.Split(value, "\n")
		if hoveredLine < len(contentLines) {
			line := contentLines[hoveredLine]
			if hoveredChar < len(line) {
				// We know the character is in the content
				// Apply background to the whole line
				highlightedLine := hoveredLineStyle.Render(lineContent)
				viewLines[visualHoveredLine] = highlightedLine
			} else {
				// Character is out of bounds, just highlight the line
				viewLines[visualHoveredLine] = hoveredLineStyle.Render(lineContent)
			}
		} else {
			// Line is out of bounds in the content, just highlight what we have
			viewLines[visualHoveredLine] = hoveredLineStyle.Render(lineContent)
		}
	} else {
		// No specific character, just highlight the entire line
		viewLines[visualHoveredLine] = hoveredLineStyle.Render(lineContent)
	}

	// Rejoin the lines to create the complete view
	return strings.Join(viewLines, "\n")
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
	if m.width > 0 {
		// Add top bar with title using adaptive colors
		titleBarBgColor := lipgloss.AdaptiveColor{Light: "#5f5fd7", Dark: "#5f5f87"} // Purple/Dark Blue
		titleBarFgColor := lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#e4e4e4"} // White

		// Create a title bar that spans the full width with centered text
		titleBar := lipgloss.NewStyle().
			Background(titleBarBgColor).
			Foreground(titleBarFgColor).
			Bold(true).
			Padding(0, 1).
			Width(m.width). // Make the bar span the full terminal width
			Align(lipgloss.Center).
			Render("bubbweb - example editor")

		// Set the top bar with a full-width title
		topBar = titleBar
	}

	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.next,
		m.keymap.prev,
		m.keymap.add,
		m.keymap.remove,
		m.keymap.quit,
	})

	// Render each editor, with special rendering for the hovered one
	var views []string
	for i, ta := range m.inputs {
		isHovered := i == m.hoveredEditor
		hovLine := -1
		hovChar := -1

		if isHovered {
			hovLine = m.hoveredLine
			hovChar = m.hoveredChar
		}

		// Use custom rendering for the hovered editor
		view := m.renderTextArea(ta, isHovered, hovLine, hovChar)
		views = append(views, view)
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

	// Include hover information in status
	var hoverInfo string
	if m.hoveredEditor >= 0 && m.hoveredLine >= 0 {
		hoverInfo = fmt.Sprintf(" | Hover: Editor %d, Line %d", m.hoveredEditor, m.hoveredLine)
		if m.hoveredChar >= 0 {
			// Get the character value for better information
			charValue := ""
			if m.hoveredEditor < len(m.inputs) {
				value := m.inputs[m.hoveredEditor].Value()
				lines := strings.Split(value, "\n")
				if m.hoveredLine < len(lines) {
					line := lines[m.hoveredLine]
					if m.hoveredChar < len(line) {
						charValue = string(line[m.hoveredChar])
						// Escape special characters for display
						if charValue == "\t" {
							charValue = "\\t"
						} else if charValue == "\n" {
							charValue = "\\n"
						} else if charValue == "\r" {
							charValue = "\\r"
						} else if charValue == " " {
							charValue = "␣" // Use symbol for space
						}
					}
				}
			}

			if charValue != "" {
				// Apply the hover char style to make the character stand out
				styledChar := hoveredCharStyle.Render(charValue)
				hoverInfo += fmt.Sprintf(", Char %d (%s)", m.hoveredChar, styledChar)
			} else {
				hoverInfo += fmt.Sprintf(", Char %d", m.hoveredChar)
			}
		}
	}

	mouseInfo = fmt.Sprintf("\n%s %s%s",
		mouseStyle.Render("Mouse:"),
		lipgloss.NewStyle().Foreground(mouseTextColor).Render(
			fmt.Sprintf("%s at %s", m.mouseEvent, m.mousePosition),
		),
		hoverInfo,
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
