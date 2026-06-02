package main

import (
	"testing"

	"farm-idle/internal/input"
	"farm-idle/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleNormalMode_Action6EntersInputMode(t *testing.T) {
	a := AppModel{
		state: model.Model{
			AutoBuyMinimum: 7,
		},
		keys: input.DefaultKeyMap,
	}

	got, _ := a.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}})
	next := got.(AppModel)

	if !next.state.InputMode {
		t.Fatal("input mode: want true, got false")
	}
	if next.state.InputBuffer != "7" {
		t.Fatalf("input buffer: want existing minimum '7', got %q", next.state.InputBuffer)
	}
	if next.state.Cursor != 5 {
		t.Fatalf("cursor: want 5, got %d", next.state.Cursor)
	}
}

func TestHandleNormalMode_DownReachesAction6(t *testing.T) {
	a := AppModel{
		state: model.Model{},
		keys:  input.DefaultKeyMap,
	}

	var got tea.Model
	for i := 0; i < 6; i++ {
		got, _ = a.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		a = got.(AppModel)
	}

	if a.state.Cursor != 5 {
		t.Fatalf("cursor: want 5, got %d", a.state.Cursor)
	}
}

func TestHandleInputMode_EnterOnAction6SavesAutoBuyMinimum(t *testing.T) {
	a := AppModel{
		state: model.Model{
			Cursor:         5,
			InputMode:      true,
			InputBuffer:    "12",
			AutoBuyMinimum: 5,
		},
	}

	got, _ := a.handleInputMode(tea.KeyMsg{Type: tea.KeyEnter})
	next := got.(AppModel)

	if next.state.AutoBuyMinimum != 12 {
		t.Fatalf("auto buy minimum: want 12, got %d", next.state.AutoBuyMinimum)
	}
	if !next.state.AutoBuyEnabled {
		t.Fatal("auto buy enabled: want true, got false")
	}
	if next.state.InputMode {
		t.Fatal("input mode: want false, got true")
	}
	if next.state.InputBuffer != "" {
		t.Fatalf("input buffer: want empty, got %q", next.state.InputBuffer)
	}
}

func TestHandleInputMode_EnterOnAction6ZeroDisablesAutoBuy(t *testing.T) {
	a := AppModel{
		state: model.Model{
			Cursor:         5,
			InputMode:      true,
			InputBuffer:    "0",
			AutoBuyEnabled: true,
			AutoBuyMinimum: 5,
		},
	}

	got, _ := a.handleInputMode(tea.KeyMsg{Type: tea.KeyEnter})
	next := got.(AppModel)

	if next.state.AutoBuyMinimum != 0 {
		t.Fatalf("auto buy minimum: want 0, got %d", next.state.AutoBuyMinimum)
	}
	if next.state.AutoBuyEnabled {
		t.Fatal("auto buy enabled: want false, got true")
	}
}

func TestHandleInputMode_EscCancelsAction6Edit(t *testing.T) {
	a := AppModel{
		state: model.Model{
			Cursor:         5,
			InputMode:      true,
			InputBuffer:    "12",
			AutoBuyEnabled: true,
			AutoBuyMinimum: 5,
		},
	}

	got, _ := a.handleInputMode(tea.KeyMsg{Type: tea.KeyEsc})
	next := got.(AppModel)

	if next.state.AutoBuyMinimum != 5 {
		t.Fatalf("auto buy minimum: want 5, got %d", next.state.AutoBuyMinimum)
	}
	if !next.state.AutoBuyEnabled {
		t.Fatal("auto buy enabled: want true, got false")
	}
	if next.state.InputMode {
		t.Fatal("input mode: want false, got true")
	}
	if next.state.InputBuffer != "" {
		t.Fatalf("input buffer: want empty, got %q", next.state.InputBuffer)
	}
}
