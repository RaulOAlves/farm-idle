package ui

import (
	"fmt"
	"strings"

	"farm-idle/internal/engine"
	"farm-idle/internal/model"
	"github.com/gdamore/tcell/v2"
)

var (
	baseStyle   = tcell.StyleDefault.Background(tcell.NewHexColor(0x0b1d18)).Foreground(tcell.ColorWhite)
	borderStyle = baseStyle.Foreground(tcell.NewHexColor(0x6dc36d))
	titleStyle  = baseStyle.Foreground(tcell.NewHexColor(0xe2d36b)).Bold(true)
	labelStyle  = baseStyle.Foreground(tcell.NewHexColor(0x8ccf7e))
	accentStyle = baseStyle.Foreground(tcell.NewHexColor(0x7ed7d1))
	mutedStyle  = baseStyle.Foreground(tcell.NewHexColor(0x7b8f88))
	soilStyle   = baseStyle.Foreground(tcell.NewHexColor(0xa87449))
	leafStyle   = baseStyle.Foreground(tcell.NewHexColor(0x79d45b))
	readyStyle  = baseStyle.Foreground(tcell.NewHexColor(0xf8cb4d)).Bold(true)
	cursorStyle = baseStyle.Foreground(tcell.NewHexColor(0x0b1d18)).Background(tcell.NewHexColor(0x98e87b)).Bold(true)
)

func Draw(screen tcell.Screen, m model.Model) {
	w, h := screen.Size()
	fill(screen, 0, 0, w, h, ' ', baseStyle)

	if m.OfflineReport != nil {
		drawOffline(screen, w, h, *m.OfflineReport)
		return
	}

	drawBanner(screen, 2, 1)
	drawText(screen, 2, 7, fmt.Sprintf("Dia %03d | Receita/min $%.0f | Auto-venda >=%d | Auto-compra %s | Foco %s",
		m.Day, m.RecentRevenue, m.AutoSellThreshold, autoBuyLabel(m), focusLabel(m)), accentStyle)

	if w >= 130 {
		drawWide(screen, m, w, h)
		return
	}
	drawStacked(screen, m, w, h)
}

func drawWide(screen tcell.Screen, m model.Model, w, h int) {
	top := 9
	leftW := maxInt(28, w/5)
	rightW := maxInt(34, w/4)
	centerW := w - leftW - rightW - 8
	panelH := maxInt(16, h-top-2)

	drawResourceBox(screen, 1, top, leftW, panelH, m)
	drawFieldBox(screen, leftW+2, top, centerW, panelH, m)
	drawSideBox(screen, leftW+centerW+3, top, rightW, panelH, m)
}

func drawStacked(screen tcell.Screen, m model.Model, w, h int) {
	top := 9
	boxW := w - 2
	halfH := maxInt(10, (h-top-4)/2)
	drawResourceBox(screen, 1, top, boxW, halfH, m)
	drawFieldBox(screen, 1, top+halfH+1, boxW, halfH, m)
}

func drawResourceBox(screen tcell.Screen, x, y, w, h int, m model.Model) {
	drawBox(screen, x, y, w, h, " RECURSOS ", borderStyle)
	lines := []string{
		fmt.Sprintf("Dinheiro   $%.0f", m.Money),
		fmt.Sprintf("Sementes   %d", m.Seeds),
		fmt.Sprintf("Estoque    %d", m.Stock),
		fmt.Sprintf("Nivel col  %d", m.HarvestLevel),
		fmt.Sprintf("Lote seed  %d", m.SeedsPerPurchase),
		"",
		"SLOT",
		fmt.Sprintf("Indice     %d/%d", selectedIndex(m), maxInt(1, len(m.Plants))),
		fmt.Sprintf("Tipo       %s", selectedType(m)),
		fmt.Sprintf("Estado     %s", selectedState(m)),
		fmt.Sprintf("Tempo      %s", selectedTime(m)),
	}
	drawLines(screen, x+2, y+1, lines, labelStyle, h-2)
}

func drawFieldBox(screen tcell.Screen, x, y, w, h int, m model.Model) {
	drawBox(screen, x, y, w, h, " CAMPO ", borderStyle)
	drawText(screen, x+2, y+1, "Tab troca foco. HJKL/setas movem cursor.", mutedStyle)

	gridY := y + 3
	gridX := x + 2
	cols := gridCols(m)
	for i, p := range m.Plants {
		cellX := gridX + (i%cols)*4
		cellY := gridY + (i / cols)
		style := slotStyle(p.State)
		text := " " + string(slotRune(p.State)) + " "
		if i == m.FieldCursor {
			style = cursorStyle
			text = "[" + string(slotRune(p.State)) + "]"
		}
		drawText(screen, cellX, cellY, text, style)
	}

	active, planted, growing, ready, empty, nextReady := summarizeField(m.Plants)
	progressY := minInt(y+h-5, gridY+(len(m.Plants)/cols)+2)
	drawProgress(screen, x+2, progressY, minInt(w-4, 28), active, maxInt(1, m.FieldSize))
	drawText(screen, x+2, progressY+1, fmt.Sprintf("Prontas %d | Crescendo %d | Vazias %d", ready, planted+growing, empty), labelStyle)
	drawText(screen, x+2, progressY+2, fmt.Sprintf("Proxima %s | Slots %d", nextReadyLabel(ready, nextReady), m.FieldSize), accentStyle)
}

func drawSideBox(screen tcell.Screen, x, y, w, h int, m model.Model) {
	drawBox(screen, x, y, w, h, " ACOES / LOG ", borderStyle)
	lines := buildActionLines(m)
	logLines := buildLogLines(m, maxInt(3, h-len(lines)-5))
	drawLines(screen, x+2, y+1, append(lines, "", "LOG"), labelStyle, h-2)
	drawLines(screen, x+2, y+len(lines)+3, logLines, mutedStyle, h-len(lines)-4)
}

func buildActionLines(m model.Model) []string {
	expandCost := engine.ExpandFieldCost(m.FieldSize)
	harvestCost := engine.HarvestUpgradeCost(m.HarvestLevel)
	lines := []string{
		fmt.Sprintf("%s [1] Comprar semente  $%.0f", cursorMark(m, 0), model.SeedCost),
		fmt.Sprintf("%s [2] Expandir campo   $%.0f", cursorMark(m, 1), expandCost),
		fmt.Sprintf("%s [3] Upgrade col.    $%.0f", cursorMark(m, 2), harvestCost),
		fmt.Sprintf("%s [4] Vender tudo", cursorMark(m, 3)),
		action5Line(m),
		action6Line(m),
		"",
		"[Q] Salvar e sair",
	}
	return lines
}

func buildLogLines(m model.Model, maxEntries int) []string {
	if len(m.Log) == 0 {
		return []string{"(sem eventos)"}
	}
	entries := m.Log
	if len(entries) > maxEntries {
		entries = entries[len(entries)-maxEntries:]
	}
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%s %s", e.Timestamp.Format("15:04:05"), e.Message))
	}
	return lines
}

func drawOffline(screen tcell.Screen, w, h int, r model.OfflineResult) {
	boxW := minInt(60, w-4)
	boxH := 9
	x := (w - boxW) / 2
	y := maxInt(2, (h-boxH)/3)
	drawBox(screen, x, y, boxW, boxH, " BEM-VINDO DE VOLTA ", borderStyle)
	lines := []string{
		fmt.Sprintf("Ausente    %02dh %02dm %02ds", int(r.Duration.Hours()), int(r.Duration.Minutes())%60, int(r.Duration.Seconds())%60),
		fmt.Sprintf("Colheitas  +%d", r.Harvests),
		fmt.Sprintf("Receita    +$%.0f", r.Revenue),
		fmt.Sprintf("Eficiencia %.0f%%", r.Efficiency*100),
		"",
		"ENTER -> Coletar",
	}
	drawLines(screen, x+2, y+1, lines, accentStyle, boxH-2)
}

func drawBanner(screen tcell.Screen, x, y int) {
	lines := []string{
		"  ____   ___   ____  __  ___   ____  _____",
		" / __/  / _ | / __ \\/  |/  /  /  _/ / ___/",
		"/ _/   / __ |/ /_/ / /|_/ /  _/ /  / /__  ",
		"/_/   /_/ |_|\\____/_/  /_/  /___/  \\___/  ",
	}
	for i, line := range lines {
		drawText(screen, x, y+i, line, leafStyle)
	}
	drawText(screen, x+44, y+3, "v0.2-dev", readyStyle)
	drawText(screen, x, y+5, "terminal farm control room", mutedStyle)
}

func drawBox(screen tcell.Screen, x, y, w, h int, title string, style tcell.Style) {
	if w < 4 || h < 3 {
		return
	}
	for dx := 0; dx < w; dx++ {
		put(screen, x+dx, y, '─', style)
		put(screen, x+dx, y+h-1, '─', style)
	}
	for dy := 0; dy < h; dy++ {
		put(screen, x, y+dy, '│', style)
		put(screen, x+w-1, y+dy, '│', style)
	}
	put(screen, x, y, '┌', style)
	put(screen, x+w-1, y, '┐', style)
	put(screen, x, y+h-1, '└', style)
	put(screen, x+w-1, y+h-1, '┘', style)
	drawText(screen, x+2, y, title, titleStyle)
}

func drawProgress(screen tcell.Screen, x, y, width, current, max int) {
	if max <= 0 || width < 4 {
		return
	}
	filled := current * (width - 2) / max
	drawText(screen, x, y, "[", mutedStyle)
	for i := 0; i < width-2; i++ {
		style := soilStyle
		ch := '░'
		if i < filled {
			style = leafStyle
			ch = '█'
		}
		put(screen, x+1+i, y, ch, style)
	}
	drawText(screen, x+width-1, y, "]", mutedStyle)
}

func drawLines(screen tcell.Screen, x, y int, lines []string, style tcell.Style, maxLines int) {
	for i, line := range lines {
		if i >= maxLines {
			return
		}
		drawText(screen, x, y+i, line, style)
	}
}

func drawText(screen tcell.Screen, x, y int, text string, style tcell.Style) {
	for i, r := range text {
		put(screen, x+i, y, r, style)
	}
}

func fill(screen tcell.Screen, x, y, w, h int, ch rune, style tcell.Style) {
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			put(screen, x+col, y+row, ch, style)
		}
	}
}

func put(screen tcell.Screen, x, y int, ch rune, style tcell.Style) {
	screen.SetContent(x, y, ch, nil, style)
}

func autoBuyLabel(m model.Model) string {
	if m.AutoBuyEnabled {
		return fmt.Sprintf("min %d", m.AutoBuyMinimum)
	}
	return "off"
}

func focusLabel(m model.Model) string {
	if m.FocusMode == model.FocusField {
		return "campo"
	}
	return "menu"
}

func cursorMark(m model.Model, idx int) string {
	if m.FocusMode == model.FocusMenu && m.Cursor == idx {
		return ">"
	}
	return " "
}

func action5Line(m model.Model) string {
	if m.InputMode && m.Cursor != 5 {
		return fmt.Sprintf("> [5] Auto-venda       %s_", m.InputBuffer)
	}
	return fmt.Sprintf("%s [5] Auto-venda       %d", cursorMark(m, 4), m.AutoSellThreshold)
}

func action6Line(m model.Model) string {
	if m.InputMode && m.Cursor == 5 {
		return fmt.Sprintf("> [6] Auto-compra      %s_", m.InputBuffer)
	}
	return fmt.Sprintf("%s [6] Auto-compra      %d", cursorMark(m, 5), m.AutoBuyMinimum)
}

func selectedIndex(m model.Model) int {
	if len(m.Plants) == 0 {
		return 0
	}
	return m.FieldCursor + 1
}

func selectedType(m model.Model) string {
	if len(m.Plants) == 0 {
		return "-"
	}
	if m.Plants[m.FieldCursor].PlantType == "" {
		return "-"
	}
	return m.Plants[m.FieldCursor].PlantType
}

func selectedState(m model.Model) string {
	if len(m.Plants) == 0 {
		return "vazia"
	}
	switch m.Plants[m.FieldCursor].State {
	case model.PlantPlanted:
		return "plantada"
	case model.PlantGrowing:
		return "crescendo"
	case model.PlantReady:
		return "pronta"
	default:
		return "vazia"
	}
}

func selectedTime(m model.Model) string {
	if len(m.Plants) == 0 {
		return "-"
	}
	slot := m.Plants[m.FieldCursor]
	switch slot.State {
	case model.PlantReady:
		return "pronta"
	case model.PlantEmpty:
		return "-"
	default:
		return fmt.Sprintf("%ds", slot.TicksRemaining)
	}
}

func summarizeField(plants []model.PlantSlot) (active, planted, growing, ready, empty, nextReady int) {
	nextReady = 9999
	for _, p := range plants {
		switch p.State {
		case model.PlantPlanted:
			active++
			planted++
			if p.TicksRemaining > 0 && p.TicksRemaining < nextReady {
				nextReady = p.TicksRemaining
			}
		case model.PlantGrowing:
			active++
			growing++
			if p.TicksRemaining > 0 && p.TicksRemaining < nextReady {
				nextReady = p.TicksRemaining
			}
		case model.PlantReady:
			active++
			ready++
			nextReady = 0
		default:
			empty++
		}
	}
	if nextReady == 9999 {
		nextReady = 0
	}
	return active, planted, growing, ready, empty, nextReady
}

func nextReadyLabel(ready, nextReady int) string {
	if ready > 0 {
		return "agora"
	}
	if nextReady > 0 {
		return fmt.Sprintf("%ds", nextReady)
	}
	return "sem cultivo"
}

func slotStyle(state model.PlantState) tcell.Style {
	switch state {
	case model.PlantPlanted:
		return soilStyle
	case model.PlantGrowing:
		return leafStyle
	case model.PlantReady:
		return readyStyle
	default:
		return mutedStyle
	}
}

func slotRune(state model.PlantState) rune {
	switch state {
	case model.PlantPlanted:
		return '·'
	case model.PlantGrowing:
		return '▓'
	case model.PlantReady:
		return '█'
	default:
		return '□'
	}
}

func gridCols(m model.Model) int {
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func _unused(_ ...string) {
	_ = strings.Builder{}
}
