package main

import (
	"testing"

	"farm-idle/internal/model"
)

func TestHandleNormalKey_Action6EntersInputMode(t *testing.T) {
	a := AppModel{
		state: model.Model{
			AutoBuyMinimum: 7,
		},
	}

	quit := a.handleNormalKey("6")
	if quit {
		t.Fatal("quit: want false, got true")
	}
	if !a.state.InputMode {
		t.Fatal("input mode: want true, got false")
	}
	if a.state.InputBuffer != "7" {
		t.Fatalf("input buffer: want existing minimum '7', got %q", a.state.InputBuffer)
	}
	if a.state.Cursor != 5 {
		t.Fatalf("cursor: want 5, got %d", a.state.Cursor)
	}
}

func TestHandleNormalKey_DownReachesAction6(t *testing.T) {
	a := AppModel{state: model.Model{}}
	for i := 0; i < 7; i++ {
		a.handleNormalKey("j")
	}
	if a.state.Cursor != 6 {
		t.Fatalf("cursor: want 6, got %d", a.state.Cursor)
	}
}

func TestHandleInputKey_EnterOnAction6SavesAutoBuyMinimum(t *testing.T) {
	a := AppModel{
		state: model.Model{
			Cursor:         5,
			InputMode:      true,
			InputBuffer:    "12",
			AutoBuyMinimum: 5,
		},
	}

	a.handleInputKey("enter")
	if a.state.AutoBuyMinimum != 12 {
		t.Fatalf("auto buy minimum: want 12, got %d", a.state.AutoBuyMinimum)
	}
	if !a.state.AutoBuyEnabled {
		t.Fatal("auto buy enabled: want true, got false")
	}
	if a.state.InputMode {
		t.Fatal("input mode: want false, got true")
	}
	if a.state.InputBuffer != "" {
		t.Fatalf("input buffer: want empty, got %q", a.state.InputBuffer)
	}
}

func TestHandleInputKey_EnterOnAction6ZeroDisablesAutoBuy(t *testing.T) {
	a := AppModel{
		state: model.Model{
			Cursor:         5,
			InputMode:      true,
			InputBuffer:    "0",
			AutoBuyEnabled: true,
			AutoBuyMinimum: 5,
		},
	}

	a.handleInputKey("enter")
	if a.state.AutoBuyMinimum != 0 {
		t.Fatalf("auto buy minimum: want 0, got %d", a.state.AutoBuyMinimum)
	}
	if a.state.AutoBuyEnabled {
		t.Fatal("auto buy enabled: want false, got true")
	}
}

func TestHandleInputKey_EscCancelsAction6Edit(t *testing.T) {
	a := AppModel{
		state: model.Model{
			Cursor:         5,
			InputMode:      true,
			InputBuffer:    "12",
			AutoBuyEnabled: true,
			AutoBuyMinimum: 5,
		},
	}

	a.handleInputKey("esc")
	if a.state.AutoBuyMinimum != 5 {
		t.Fatalf("auto buy minimum: want 5, got %d", a.state.AutoBuyMinimum)
	}
	if !a.state.AutoBuyEnabled {
		t.Fatal("auto buy enabled: want true, got false")
	}
	if a.state.InputMode {
		t.Fatal("input mode: want false, got true")
	}
	if a.state.InputBuffer != "" {
		t.Fatalf("input buffer: want empty, got %q", a.state.InputBuffer)
	}
}

func TestHandleNormalKey_TabTogglesFieldFocus(t *testing.T) {
	a := AppModel{
		state: model.Model{
			FocusMode: model.FocusMenu,
		},
	}

	a.handleNormalKey("tab")
	if a.state.FocusMode != model.FocusField {
		t.Fatalf("focus mode: want %q, got %q", model.FocusField, a.state.FocusMode)
	}
}

func TestHandleNormalKey_FieldFocusMovesGridCursor(t *testing.T) {
	a := AppModel{
		state: model.Model{
			FocusMode:   model.FocusField,
			FieldCursor: 0,
			ViewWidth:   80,
			Plants:      make([]model.PlantSlot, 20),
		},
	}

	a.handleNormalKey("l")
	if a.state.FieldCursor != 1 {
		t.Fatalf("field cursor after right: want 1, got %d", a.state.FieldCursor)
	}

	a.handleNormalKey("j")
	if a.state.FieldCursor != 11 {
		t.Fatalf("field cursor after down: want 11, got %d", a.state.FieldCursor)
	}
}

func TestHandleNormalKey_Action7CyclesCrop(t *testing.T) {
	a := AppModel{
		state: model.Model{
			SelectedCrop: model.DefaultPlantType,
		},
	}

	a.handleNormalKey("7")
	if a.state.SelectedCrop == model.DefaultPlantType {
		t.Fatalf("selected crop should change from %q", model.DefaultPlantType)
	}
}
