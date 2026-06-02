package ui

import (
	"fmt"

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
	skyStyle    = tcell.StyleDefault.Background(tcell.NewHexColor(0x19324a)).Foreground(tcell.NewHexColor(0xa9e7ff))
	grassStyle  = tcell.StyleDefault.Background(tcell.NewHexColor(0x254f22)).Foreground(tcell.NewHexColor(0x87d96c))
	furrowA     = tcell.StyleDefault.Background(tcell.NewHexColor(0x4b2e1f)).Foreground(tcell.NewHexColor(0x8b5c3a))
	furrowB     = tcell.StyleDefault.Background(tcell.NewHexColor(0x5a3825)).Foreground(tcell.NewHexColor(0xa87449))
	sproutStyle = tcell.StyleDefault.Background(tcell.NewHexColor(0x4b2e1f)).Foreground(tcell.NewHexColor(0x9ee06d)).Bold(true)
	growStyle   = tcell.StyleDefault.Background(tcell.NewHexColor(0x5a3825)).Foreground(tcell.NewHexColor(0x63d471)).Bold(true)
	wheatStyle  = tcell.StyleDefault.Background(tcell.NewHexColor(0x5a3825)).Foreground(tcell.NewHexColor(0xffd966)).Bold(true)
	nightSky    = tcell.StyleDefault.Background(tcell.NewHexColor(0x10192f)).Foreground(tcell.NewHexColor(0xd7e8ff))
	dawnSky     = tcell.StyleDefault.Background(tcell.NewHexColor(0x5b3659)).Foreground(tcell.NewHexColor(0xffd7b0))
	daySky      = tcell.StyleDefault.Background(tcell.NewHexColor(0x19324a)).Foreground(tcell.NewHexColor(0xa9e7ff))
	duskSky     = tcell.StyleDefault.Background(tcell.NewHexColor(0x4a2941)).Foreground(tcell.NewHexColor(0xffca8a))
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
	drawText(screen, 2, 8, fmt.Sprintf("Ambiente %s | Vento %s | Cultura %s $%.0f", phaseLabel(m), breezeLabel(m), model.CropByName(m.SelectedCrop).Name, model.CropByName(m.SelectedCrop).SellPrice), mutedStyle)

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
		fmt.Sprintf("Valor est. $%.0f", engine.StockMarketValue(m)),
		fmt.Sprintf("Nivel col  %d", m.HarvestLevel),
		fmt.Sprintf("Lote seed  %d", m.SeedsPerPurchase),
		fmt.Sprintf("Plantio    %s", m.SelectedCrop),
		fmt.Sprintf("Preco/un   $%.0f", model.CropByName(m.SelectedCrop).SellPrice),
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
	drawText(screen, x+2, y+1, fmt.Sprintf("Tab troca foco. HJKL/setas movem cursor. Clima: %s", phaseLabel(m)), mutedStyle)

	fieldTop := y + 3
	fieldHeight := maxInt(6, h-7)
	fieldBottom := minInt(y+h-4, fieldTop+fieldHeight-1)
	gridX := x + 2
	cols := gridCols(m)
	drawFieldBackdrop(screen, gridX, fieldTop, w-4, fieldBottom-fieldTop+1, m)

	tileW := 5
	tileH := 2
	gridY := fieldTop + 2
	blocksPerRow := fieldBlocksPerRow(w)
	for i, p := range m.Plants {
		cellX, cellY := fieldSlotPosition(gridX, gridY, i, tileW, tileH, blocksPerRow)
		if cellY+1 > fieldBottom-1 || cellX+tileW-1 > x+w-3 {
			continue
		}
		drawCropTile(screen, cellX, cellY, tileW, p, i == m.FieldCursor, m.FocusMode == model.FocusField, (i/cols)%2 == 0, m.TickCount)
	}

	active, planted, growing, ready, empty, nextReady := summarizeField(m.Plants)
	rows := visibleFieldRows(len(m.Plants), blocksPerRow)
	progressY := minInt(y+h-5, gridY+rows*tileH+1)
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
		action7Line(m),
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

func action7Line(m model.Model) string {
	crop := model.CropByName(m.SelectedCrop)
	return fmt.Sprintf("%s [7] Cultura ativa    %s (%ds $%.0f)", cursorMark(m, 6), crop.Name, crop.GrowTicks, crop.SellPrice)
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
		return sproutStyle
	case model.PlantGrowing:
		return growStyle
	case model.PlantReady:
		return wheatStyle
	default:
		return furrowA
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

func drawFieldBackdrop(screen tcell.Screen, x, y, w, h int, m model.Model) {
	sky := skyForPhase(m)
	skyRows := minInt(2, h)
	for row := 0; row < skyRows; row++ {
		for col := 0; col < w; col++ {
			ch := ' '
			if isNightPhase(m) && row == 0 && (col+m.TickCount)%11 == 0 {
				ch = '✦'
			}
			if row == 1 && col%9 == 0 && !isNightPhase(m) {
				ch = '·'
			}
			put(screen, x+col, y+row, ch, sky)
		}
	}
	drawSkyAccent(screen, x, y, w, skyRows, m)
	if h <= skyRows {
		return
	}
	for col := 0; col < w; col++ {
		ch := '▄'
		style := grassStyle
		if col%7 == 0 {
			ch = '▆'
		}
		put(screen, x+col, y+skyRows, ch, style)
	}
	for row := skyRows + 1; row < h; row++ {
		for col := 0; col < w; col++ {
			style := furrowA
			ch := ' '
			if ((row+col)/2)%2 == 0 {
				style = furrowB
			}
			if col%5 == 0 {
				ch = '·'
			}
			put(screen, x+col, y+row, ch, style)
		}
	}
}

func drawCropTile(screen tcell.Screen, x, y, w int, slot model.PlantSlot, selected, focused, alt bool, tickCount int) {
	base := furrowA
	if alt {
		base = furrowB
	}
	style := slotStyle(slot.State)
	if slot.State == model.PlantEmpty {
		style = base
	}

	for dx := 0; dx < w; dx++ {
		put(screen, x+dx, y, ' ', base)
		put(screen, x+dx, y+1, ' ', base)
	}

	drawText(screen, x, y+1, "~~~~~", base)

	if selected {
		frame := cursorStyle
		if !focused {
			frame = accentStyle.Background(tcell.NewHexColor(0x173038)).Bold(true)
		}
		put(screen, x, y, '[', frame)
		put(screen, x+w-1, y, ']', frame)
		put(screen, x, y+1, '[', frame)
		put(screen, x+w-1, y+1, ']', frame)
	}

	glyph := cropNameToGlyph(slot.PlantType, slot.State, tickCount)
	textX := x + 2
	for i, r := range glyph {
		put(screen, textX+i, y, r, style)
	}
	if slot.State == model.PlantReady {
		readyTail := []rune(readyTailGlyph(tickCount))
		put(screen, x+2, y+1, readyTail[0], wheatStyle)
		put(screen, x+3, y+1, readyTail[1], wheatStyle)
	}
}

func cropGlyph(state model.PlantState, tickCount int) string {
	sway := tickCount % 4
	switch state {
	case model.PlantPlanted:
		if sway < 2 {
			return " i"
		}
		return " i"
	case model.PlantGrowing:
		if sway < 2 {
			return "Yv"
		}
		return "vY"
	case model.PlantReady:
		if sway < 2 {
			return "WW"
		}
		return "MM"
	default:
		return "··"
	}
}

func cropNameToGlyph(name string, state model.PlantState, tickCount int) string {
	sway := tickCount % 4
	switch name {
	case "Alface":
		switch state {
		case model.PlantPlanted:
			return " i"
		case model.PlantGrowing:
			if sway < 2 {
				return "()"
			}
			return "(("
		case model.PlantReady:
			return "@@"
		}
	case "Milho":
		switch state {
		case model.PlantPlanted:
			return " !"
		case model.PlantGrowing:
			if sway < 2 {
				return "||"
			}
			return "!!"
		case model.PlantReady:
			return "H#"
		}
	}
	return cropGlyph(state, tickCount)
}

func readyTailGlyph(tickCount int) string {
	if tickCount%4 < 2 {
		return "mm"
	}
	return "nn"
}

func gridCols(m model.Model) int {
	return 10
}

func fieldBlocksPerRow(panelWidth int) int {
	tileW := 5
	blockWidth := 10 * tileW
	withGap := blockWidth + 2
	blocks := maxInt(1, (panelWidth-4)/withGap)
	return blocks
}

func fieldSlotPosition(gridX, gridY, slotIndex, tileW, tileH, blocksPerRow int) (int, int) {
	const blockSize = 10
	blockCapacity := blockSize * blockSize
	block := slotIndex / blockCapacity
	inBlock := slotIndex % blockCapacity
	blockX := block % blocksPerRow
	blockY := block / blocksPerRow
	col := inBlock % blockSize
	row := inBlock / blockSize
	x := gridX + blockX*(blockSize*tileW+2) + col*tileW
	y := gridY + blockY*(blockSize*tileH+1) + row*tileH
	return x, y
}

func visibleFieldRows(slotCount, blocksPerRow int) int {
	const blockSize = 10
	const blockCapacity = blockSize * blockSize
	if slotCount <= 0 {
		return 0
	}
	blocks := (slotCount + blockCapacity - 1) / blockCapacity
	blockRows := (blocks + blocksPerRow - 1) / blocksPerRow
	lastBlockSlots := slotCount % blockCapacity
	if lastBlockSlots == 0 {
		lastBlockSlots = blockCapacity
	}
	lastBlockRows := (lastBlockSlots + blockSize - 1) / blockSize
	if blockRows <= 1 {
		return lastBlockRows
	}
	return (blockRows-1)*blockSize + lastBlockRows
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

func phaseIndex(m model.Model) int {
	if model.TicksPerDay <= 0 {
		return 1
	}
	segment := model.TicksPerDay / 4
	if segment <= 0 {
		segment = 1
	}
	return (m.TickCount / segment) % 4
}

func phaseLabel(m model.Model) string {
	switch phaseIndex(m) {
	case 0:
		return "amanhecer"
	case 1:
		return "dia aberto"
	case 2:
		return "entardecer"
	default:
		return "noite"
	}
}

func breezeLabel(m model.Model) string {
	switch m.TickCount % 6 {
	case 0, 1:
		return "calmo"
	case 2, 3:
		return "leve"
	default:
		return "soprando"
	}
}

func isNightPhase(m model.Model) bool {
	return phaseIndex(m) == 3
}

func skyForPhase(m model.Model) tcell.Style {
	switch phaseIndex(m) {
	case 0:
		return dawnSky
	case 1:
		return daySky
	case 2:
		return duskSky
	default:
		return nightSky
	}
}

func drawSkyAccent(screen tcell.Screen, x, y, w, skyRows int, m model.Model) {
	if skyRows == 0 || w < 6 {
		return
	}
	iconX := x + minInt(w-4, 2+(m.TickCount%(maxInt(1, w-8))))
	switch phaseIndex(m) {
	case 0:
		drawText(screen, iconX, y, "◔", readyStyle)
	case 1:
		drawText(screen, iconX, y, "☼", readyStyle)
	case 2:
		drawText(screen, iconX, y, "◕", readyStyle)
	default:
		drawText(screen, iconX, y, "☾", accentStyle)
	}
}
