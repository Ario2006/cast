// Package ui holds small terminal interactions that are independent from cast's
// domain services.
package ui

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Select presents choices in a keyboard-driven Bubble Tea menu. It returns
// selected=false when the user cancels with q, Escape, or Ctrl+C.
func Select(input io.Reader, output io.Writer, title string, choices []string) (choice string, selected bool, err error) {
	index, selected, err := SelectIndex(input, output, title, choices)
	if err != nil || !selected {
		return "", selected, err
	}
	return choices[index], true, nil
}

// SelectIndex presents choices and returns the selected zero-based index.
func SelectIndex(input io.Reader, output io.Writer, title string, choices []string) (index int, selected bool, err error) {
	if len(choices) == 0 {
		return 0, false, fmt.Errorf("cannot select from an empty list")
	}
	program := tea.NewProgram(selector{title: title, choices: choices}, tea.WithInput(input), tea.WithOutput(output))
	final, err := program.Run()
	if err != nil {
		return 0, false, err
	}
	model, ok := final.(selector)
	if !ok {
		return 0, false, fmt.Errorf("unexpected selector model")
	}
	if model.cancelled || !model.selected {
		return 0, false, nil
	}
	return model.cursor, true, nil
}

type selector struct {
	title     string
	choices   []string
	cursor    int
	selected  bool
	cancelled bool
}

func (m selector) Init() tea.Cmd { return nil }

func (m selector) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := message.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "q", "esc", "ctrl+c":
		m.cancelled = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.choices)-1 {
			m.cursor++
		}
	case "enter":
		m.selected = true
		return m, tea.Quit
	default:
		if index, err := strconv.Atoi(key.String()); err == nil && index >= 1 && index <= len(m.choices) {
			m.cursor = index - 1
			m.selected = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selector) View() tea.View {
	var view strings.Builder
	view.WriteString(m.title)
	view.WriteString("\n\n")
	for index, choice := range m.choices {
		prefix := " "
		if index == m.cursor {
			prefix = ">"
		}
		fmt.Fprintf(&view, "%s %d. %s\n", prefix, index+1, choice)
	}
	view.WriteString("\n↑ ↓ move   Enter select   number select   q quit\n")
	return tea.NewView(view.String())
}
