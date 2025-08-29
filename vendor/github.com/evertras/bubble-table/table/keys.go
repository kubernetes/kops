package table

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the keybindings for the table when it's focused.
type KeyMap struct {
	RowDown key.Binding
	RowUp   key.Binding

	RowSelectToggle key.Binding

	PageDown  key.Binding
	PageUp    key.Binding
	PageFirst key.Binding
	PageLast  key.Binding

	// Filter allows the user to start typing and filter the rows.
	Filter key.Binding

	// FilterBlur is the key that stops the user's input from typing into the filter.
	FilterBlur key.Binding

	// FilterClear will clear the filter while it's blurred.
	FilterClear key.Binding

	// ScrollRight will move one column to the right when overflow occurs.
	ScrollRight key.Binding

	// ScrollLeft will move one column to the left when overflow occurs.
	ScrollLeft key.Binding
}

// DefaultKeyMap returns a set of sensible defaults for controlling a focused table with help text.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		RowDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		RowUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		RowSelectToggle: key.NewBinding(
			key.WithKeys(" ", "enter"),
			key.WithHelp("<space>/enter", "select row"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("right", "l", "pgdown"),
			key.WithHelp("→/h/page down", "next page"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("left", "h", "pgup"),
			key.WithHelp("←/h/page up", "previous page"),
		),
		PageFirst: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "first page"),
		),
		PageLast: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "last page"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		FilterBlur: key.NewBinding(
			key.WithKeys("enter", "esc"),
			key.WithHelp("enter/esc", "unfocus"),
		),
		FilterClear: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),
		ScrollRight: key.NewBinding(
			key.WithKeys("shift+right"),
			key.WithHelp("shift+→", "scroll right"),
		),
		ScrollLeft: key.NewBinding(
			key.WithKeys("shift+left"),
			key.WithHelp("shift+←", "scroll left"),
		),
	}
}

// FullHelp returns a multi row view of all the helpkeys that are defined. Needed to fullfil the 'help.Model' interface.
// Also appends all user defined extra keys to the help.
func (m Model) FullHelp() [][]key.Binding {
	keyBinds := [][]key.Binding{
		{m.keyMap.RowDown, m.keyMap.RowUp, m.keyMap.RowSelectToggle},
		{m.keyMap.PageDown, m.keyMap.PageUp, m.keyMap.PageFirst, m.keyMap.PageLast},
		{m.keyMap.Filter, m.keyMap.FilterBlur, m.keyMap.FilterClear, m.keyMap.ScrollRight, m.keyMap.ScrollLeft},
	}
	if m.additionalFullHelpKeys != nil {
		keyBinds = append(keyBinds, m.additionalFullHelpKeys())
	}

	return keyBinds
}

// ShortHelp just returns a single row of help views. Needed to fullfil the 'help.Model' interface.
// Also appends all user defined extra keys to the help.
func (m Model) ShortHelp() []key.Binding {
	keyBinds := []key.Binding{
		m.keyMap.RowDown,
		m.keyMap.RowUp,
		m.keyMap.RowSelectToggle,
		m.keyMap.PageDown,
		m.keyMap.PageUp,
		m.keyMap.Filter,
		m.keyMap.FilterBlur,
		m.keyMap.FilterClear,
	}
	if m.additionalShortHelpKeys != nil {
		keyBinds = append(keyBinds, m.additionalShortHelpKeys()...)
	}

	return keyBinds
}
