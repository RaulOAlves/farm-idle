package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"farm-idle/internal/engine"
	"farm-idle/internal/input"
	"farm-idle/internal/model"
	"farm-idle/internal/persistence"
	"farm-idle/internal/renderer"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const savePath = "save.json"

type tickMsg time.Time
type autoSaveMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func autoSaveCmd() tea.Cmd {
	return tea.Tick(model.AutoSaveInterval, func(t time.Time) tea.Msg {
		return autoSaveMsg(t)
	})
}

type AppModel struct {
	state model.Model
	keys  input.KeyMap
}

func newAppModel() AppModel {
	m, _ := persistence.Load(savePath)
	if !m.LastSave.IsZero() {
		newState, report := engine.CalcOfflineProgress(m, time.Now())
		if report.Duration >= time.Minute {
			newState.OfflineReport = &report
		}
		m = newState
	}

	return AppModel{state: m, keys: input.DefaultKeyMap}
}

func (a AppModel) Init() tea.Cmd {
	return tea.Batch(tickCmd(), autoSaveCmd())
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		a.state = engine.Tick(a.state)
		return a, tickCmd()

	case autoSaveMsg:
		a.state.LastSave = time.Now()
		_ = persistence.Save(a.state, savePath)
		return a, autoSaveCmd()

	case tea.WindowSizeMsg:
		a.state.ViewWidth = msg.Width
		a.state.ViewHeight = msg.Height
		a.state = normalizeUIState(a.state)
		return a, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
		if a.state.InputMode {
			return a.handleInputMode(msg)
		}
		return a.handleNormalMode(msg)
	}

	return a, nil
}

func (a AppModel) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	a.state = normalizeUIState(a.state)

	if a.state.OfflineReport != nil {
		if key.Matches(msg, a.keys.Select) {
			a.state.OfflineReport = nil
		}
		return a, nil
	}

	switch {
	case msg.String() == "tab":
		if a.state.FocusMode == model.FocusField {
			a.state.FocusMode = model.FocusMenu
		} else {
			a.state.FocusMode = model.FocusField
		}
	case a.state.FocusMode == model.FocusField && msg.String() == "h":
		a.state.FieldCursor = moveFieldCursor(a.state, -1, 0)
	case a.state.FocusMode == model.FocusField && msg.String() == "l":
		a.state.FieldCursor = moveFieldCursor(a.state, 1, 0)
	case key.Matches(msg, a.keys.Up):
		if a.state.FocusMode == model.FocusField {
			a.state.FieldCursor = moveFieldCursor(a.state, 0, -1)
		} else if a.state.Cursor > 0 {
			a.state.Cursor--
		}
	case key.Matches(msg, a.keys.Down):
		if a.state.FocusMode == model.FocusField {
			a.state.FieldCursor = moveFieldCursor(a.state, 0, 1)
		} else if a.state.Cursor < 5 {
			a.state.Cursor++
		}
	case msg.String() == "left":
		if a.state.FocusMode == model.FocusField {
			a.state.FieldCursor = moveFieldCursor(a.state, -1, 0)
		}
	case msg.String() == "right":
		if a.state.FocusMode == model.FocusField {
			a.state.FieldCursor = moveFieldCursor(a.state, 1, 0)
		}
	case key.Matches(msg, a.keys.Select):
		if a.state.FocusMode == model.FocusMenu {
			a = a.executeAction(a.state.Cursor)
		}
	case key.Matches(msg, a.keys.Action1):
		a = a.executeAction(0)
	case key.Matches(msg, a.keys.Action2):
		a = a.executeAction(1)
	case key.Matches(msg, a.keys.Action3):
		a = a.executeAction(2)
	case key.Matches(msg, a.keys.Action4):
		a = a.executeAction(3)
	case key.Matches(msg, a.keys.Action5):
		a = a.executeAction(4)
	case key.Matches(msg, a.keys.Action6):
		a.state.InputMode = true
		a.state.Cursor = 5
		a.state.InputBuffer = strconv.Itoa(a.state.AutoBuyMinimum)
	case key.Matches(msg, a.keys.Quit):
		a.state.LastSave = time.Now()
		_ = persistence.Save(a.state, savePath)
		return a, tea.Quit
	}

	return a, nil
}

func (a AppModel) executeAction(index int) AppModel {
	var err error

	switch index {
	case 0:
		a.state, err = engine.BuySeeds(a.state)
	case 1:
		a.state, err = engine.ExpandField(a.state)
	case 2:
		a.state, err = engine.UpgradeHarvest(a.state)
	case 3:
		a.state, _ = engine.SellAll(a.state)
	case 4:
		a.state.InputMode = true
		a.state.InputBuffer = ""
	}

	if err != nil {
		_ = err
	}

	return a
}

func normalizeUIState(m model.Model) model.Model {
	if m.FocusMode == "" {
		m.FocusMode = model.FocusMenu
	}
	if len(m.Plants) == 0 {
		m.FieldCursor = 0
		return m
	}
	if m.FieldCursor < 0 {
		m.FieldCursor = 0
	}
	if m.FieldCursor >= len(m.Plants) {
		m.FieldCursor = len(m.Plants) - 1
	}
	return m
}

func moveFieldCursor(m model.Model, dx, dy int) int {
	if len(m.Plants) == 0 {
		return 0
	}

	cols := fieldGridCols(m)
	if cols <= 0 {
		cols = 4
	}

	idx := m.FieldCursor
	x := idx % cols
	y := idx / cols
	nextX := x + dx
	nextY := y + dy
	if nextX < 0 || nextX >= cols || nextY < 0 {
		return idx
	}

	next := nextY*cols + nextX
	if next < 0 || next >= len(m.Plants) {
		return idx
	}
	return next
}

func fieldGridCols(m model.Model) int {
	switch {
	case m.ViewWidth >= 120:
		return 6
	case m.ViewWidth >= 90:
		return 5
	default:
		return 4
	}
}

func (a AppModel) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		var v int
		fmt.Sscanf(a.state.InputBuffer, "%d", &v)
		if a.state.Cursor == 5 {
			if v < 0 {
				v = 0
			}
			a.state.AutoBuyMinimum = v
			a.state.AutoBuyEnabled = v > 0
		} else {
			a.state = engine.SetAutoSellThreshold(a.state, v)
		}
		a.state.InputMode = false
		a.state.InputBuffer = ""
	case "esc":
		a.state.InputMode = false
		a.state.InputBuffer = ""
	case "backspace":
		if len(a.state.InputBuffer) > 0 {
			a.state.InputBuffer = a.state.InputBuffer[:len(a.state.InputBuffer)-1]
		}
	default:
		ch := msg.String()
		if len(ch) == 1 && ch >= "0" && ch <= "9" && len(a.state.InputBuffer) < 4 {
			a.state.InputBuffer += ch
		}
	}

	return a, nil
}

func (a AppModel) View() string {
	return renderer.View(normalizeUIState(a.state))
}

func main() {
	p := tea.NewProgram(newAppModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
