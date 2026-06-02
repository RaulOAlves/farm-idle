package input

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Select  key.Binding
	Back    key.Binding
	Action1 key.Binding
	Action2 key.Binding
	Action3 key.Binding
	Action4 key.Binding
	Action5 key.Binding
	Action6 key.Binding
	Quit    key.Binding
}

var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "cima"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "baixo"),
	),
	Select: key.NewBinding(
		key.WithKeys("l", "enter"),
		key.WithHelp("l/enter", "confirmar"),
	),
	Back: key.NewBinding(
		key.WithKeys("h", "esc"),
		key.WithHelp("h/esc", "cancelar"),
	),
	Action1: key.NewBinding(key.WithKeys("1")),
	Action2: key.NewBinding(key.WithKeys("2")),
	Action3: key.NewBinding(key.WithKeys("3")),
	Action4: key.NewBinding(key.WithKeys("4")),
	Action5: key.NewBinding(key.WithKeys("5")),
	Action6: key.NewBinding(key.WithKeys("6")),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "salvar e sair"),
	),
}
