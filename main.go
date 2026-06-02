package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"farm-idle/internal/engine"
	"farm-idle/internal/model"
	"farm-idle/internal/persistence"
	"farm-idle/internal/ui"
	"github.com/gdamore/tcell/v2"
)

const savePath = "save.json"

type AppModel struct {
	state model.Model
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

	return AppModel{state: normalizeUIState(m)}
}

func (a *AppModel) handleNormalKey(name string) bool {
	a.state = normalizeUIState(a.state)

	if a.state.OfflineReport != nil {
		if name == "enter" {
			a.state.OfflineReport = nil
		}
		return false
	}

	switch name {
	case "tab":
		if a.state.FocusMode == model.FocusField {
			a.state.FocusMode = model.FocusMenu
		} else {
			a.state.FocusMode = model.FocusField
		}
	case "h":
		if a.state.FocusMode == model.FocusField {
			a.state.FieldCursor = moveFieldCursor(a.state, -1, 0)
		}
	case "l":
		if a.state.FocusMode == model.FocusField {
			a.state.FieldCursor = moveFieldCursor(a.state, 1, 0)
		} else {
			a.executeAction(a.state.Cursor)
		}
	case "up", "k":
		if a.state.FocusMode == model.FocusField {
			a.state.FieldCursor = moveFieldCursor(a.state, 0, -1)
		} else if a.state.Cursor > 0 {
			a.state.Cursor--
		}
	case "down", "j":
		if a.state.FocusMode == model.FocusField {
			a.state.FieldCursor = moveFieldCursor(a.state, 0, 1)
		} else if a.state.Cursor < 6 {
			a.state.Cursor++
		}
	case "left":
		if a.state.FocusMode == model.FocusField {
			a.state.FieldCursor = moveFieldCursor(a.state, -1, 0)
		}
	case "right":
		if a.state.FocusMode == model.FocusField {
			a.state.FieldCursor = moveFieldCursor(a.state, 1, 0)
		}
	case "enter":
		if a.state.FocusMode == model.FocusMenu {
			a.executeAction(a.state.Cursor)
		}
	case "1":
		a.executeAction(0)
	case "2":
		a.executeAction(1)
	case "3":
		a.executeAction(2)
	case "4":
		a.executeAction(3)
	case "5":
		a.executeAction(4)
	case "6":
		a.executeAction(5)
	case "7":
		a.executeAction(6)
	case "q", "ctrl+c":
		a.state.LastSave = time.Now()
		_ = persistence.Save(a.state, savePath)
		return true
	}

	return false
}

func (a *AppModel) executeAction(index int) {
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
	case 5:
		a.state.InputMode = true
		a.state.Cursor = 5
		a.state.InputBuffer = strconv.Itoa(a.state.AutoBuyMinimum)
	case 6:
		a.state = engine.CycleSelectedCrop(a.state)
	}

	if err != nil {
		_ = err
	}
}

func (a *AppModel) handleInputKey(name string) {
	switch name {
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
		if len(name) == 1 && name >= "0" && name <= "9" && len(a.state.InputBuffer) < 4 {
			a.state.InputBuffer += name
		}
	}
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
	case m.ViewWidth >= 150:
		return 8
	case m.ViewWidth >= 120:
		return 6
	case m.ViewWidth >= 90:
		return 5
	default:
		return 4
	}
}

func keyName(ev *tcell.EventKey) string {
	switch ev.Key() {
	case tcell.KeyCtrlC:
		return "ctrl+c"
	case tcell.KeyUp:
		return "up"
	case tcell.KeyDown:
		return "down"
	case tcell.KeyLeft:
		return "left"
	case tcell.KeyRight:
		return "right"
	case tcell.KeyEnter:
		return "enter"
	case tcell.KeyEsc:
		return "esc"
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return "backspace"
	case tcell.KeyTAB:
		return "tab"
	case tcell.KeyRune:
		return string(ev.Rune())
	default:
		return ""
	}
}

func updateViewport(a *AppModel, screen tcell.Screen) {
	w, h := screen.Size()
	a.state.ViewWidth = w
	a.state.ViewHeight = h
	a.state = normalizeUIState(a.state)
}

func runApp() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := screen.Init(); err != nil {
		return err
	}
	defer screen.Fini()

	screen.Clear()

	app := newAppModel()
	updateViewport(&app, screen)

	events := make(chan tcell.Event, 32)
	go func() {
		for {
			events <- screen.PollEvent()
		}
	}()

	tickTicker := time.NewTicker(time.Second)
	saveTicker := time.NewTicker(model.AutoSaveInterval)
	defer tickTicker.Stop()
	defer saveTicker.Stop()

	for {
		ui.Draw(screen, app.state)
		screen.Show()

		select {
		case <-tickTicker.C:
			app.state = engine.Tick(app.state)
		case <-saveTicker.C:
			app.state.LastSave = time.Now()
			_ = persistence.Save(app.state, savePath)
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				screen.Sync()
				updateViewport(&app, screen)
			case *tcell.EventKey:
				name := keyName(ev)
				if name == "" {
					continue
				}
				if app.state.InputMode {
					app.handleInputKey(name)
				} else if app.handleNormalKey(name) {
					return nil
				}
			}
		}
	}
}

func main() {
	if err := runApp(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
