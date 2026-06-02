package renderer

import (
	"fmt"
	"math"
	"strings"

	"farm-idle/internal/engine"
	"farm-idle/internal/model"
	"github.com/charmbracelet/lipgloss"
)

var (
	outerStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	boldStyle   = lipgloss.NewStyle().Bold(true)
	faintStyle  = lipgloss.NewStyle().Faint(true)
	cursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	panelStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
)

func View(m model.Model) string {
	contentWidth := contentWidth(m)

	if m.OfflineReport != nil {
		body := outerStyle.Width(contentWidth).Render(renderOfflineReport(*m.OfflineReport))
		return placeOnScreen(m, body)
	}

	sections := []string{
		renderBanner(m, contentWidth),
		renderStatusLine(m),
		sep(contentWidth),
	}

	if contentWidth >= 130 {
		sections = append(sections, renderWideDashboard(m, contentWidth))
	} else {
		sections = append(sections,
			renderMainContent(m, contentWidth),
			sep(contentWidth),
			renderActions(m),
			sep(contentWidth),
			renderLog(m, visibleLogEntries(m)),
		)
	}

	body := outerStyle.Width(contentWidth).Render(strings.Join(sections, "\n"))
	return placeOnScreen(m, body)
}

func contentWidth(m model.Model) int {
	if m.ViewWidth <= 0 {
		return 68
	}
	width := m.ViewWidth - 4
	if width < 68 {
		return 68
	}
	return width
}

func sep(width int) string {
	if width < 32 {
		width = 32
	}
	return strings.Repeat("─", width-4)
}

func renderMainContent(m model.Model, width int) string {
	panelHeight := stackedPanelHeight(m)
	left := panelStyle.Width(leftColumnWidth(width)).Height(panelHeight).Render(strings.Join([]string{
		renderResources(m),
		"",
		renderSelectedSlot(m),
	}, "\n"))
	right := panelStyle.Width(rightColumnWidth(width)).Height(panelHeight).Render(renderField(m))

	if width < 96 {
		return lipgloss.JoinVertical(lipgloss.Left, left, right)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
}

func renderWideDashboard(m model.Model, width int) string {
	leftWidth := maxInt(28, width/5)
	rightWidth := maxInt(34, width/4)
	centerWidth := width - leftWidth - rightWidth - 10
	panelHeight := widePanelHeight(m)

	left := panelStyle.Width(leftWidth).Height(panelHeight).Render(strings.Join([]string{
		renderResources(m),
		"",
		renderSelectedSlot(m),
	}, "\n"))
	center := panelStyle.Width(centerWidth).Height(panelHeight).Render(renderField(m))
	right := panelStyle.Width(rightWidth).Height(panelHeight).Render(strings.Join([]string{
		renderActions(m),
		"",
		sep(rightWidth),
		renderLog(m, panelHeight-8),
	}, "\n"))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", center, "  ", right)
}

func leftColumnWidth(width int) int {
	if width < 96 {
		return width - 4
	}
	return 30
}

func rightColumnWidth(width int) int {
	if width < 96 {
		return width - 4
	}
	return width - leftColumnWidth(width) - 8
}

func renderStatusLine(m model.Model) string {
	autoBuy := "off"
	if m.AutoBuyEnabled {
		autoBuy = fmt.Sprintf("min %d", m.AutoBuyMinimum)
	}

	focus := "menu"
	if m.FocusMode == model.FocusField {
		focus = "campo"
	}

	return fmt.Sprintf("Dia %03d │ Receita/min $%.0f │ Auto-venda ≥%d │ Auto-compra %s │ Foco %s", m.Day, m.RecentRevenue, m.AutoSellThreshold, autoBuy, focus)
}

func renderResources(m model.Model) string {
	lines := []string{
		panelTitle("RECURSOS"),
		fmt.Sprintf("💰 Dinheiro:  $%.0f", m.Money),
		fmt.Sprintf("🌱 Sementes:  %d", m.Seeds),
		fmt.Sprintf("📦 Estoque:   %d", m.Stock),
		fmt.Sprintf("📈 Nível col: %d", m.HarvestLevel),
		fmt.Sprintf("⚙ Lote seed:  %d", m.SeedsPerPurchase),
	}

	return strings.Join(lines, "\n")
}

func renderField(m model.Model) string {
	active, planted, growing, ready, empty, nextReady := summarizeField(m.Plants)
	pct := 0.0
	if m.FieldSize > 0 {
		pct = float64(active) / float64(m.FieldSize)
	}
	nextReadyText := "sem cultivo"
	if ready > 0 {
		nextReadyText = "agora"
	} else if nextReady > 0 {
		nextReadyText = fmt.Sprintf("%ds", nextReady)
	}
	cols := gridCols(m)

	lines := []string{
		panelTitle("CAMPO"),
		faintStyle.Render("Tab alterna foco. HJKL/setas movem no grid."),
		renderMiniGrid(m, cols),
		progressBar(active, m.FieldSize, 20, pct),
		fmt.Sprintf("Prontas: %d  Crescendo: %d  Vazias: %d", ready, planted+growing, empty),
		fmt.Sprintf("Próxima: %s  Slots: %d", nextReadyText, m.FieldSize),
	}

	return strings.Join(lines, "\n")
}

func renderSelectedSlot(m model.Model) string {
	lines := []string{panelTitle("SLOT")}
	slot, ok := selectedSlot(m)
	if !ok {
		lines = append(lines, faintStyle.Render("Campo vazio"))
		return strings.Join(lines, "\n")
	}

	plantType := slot.PlantType
	if plantType == "" {
		plantType = "—"
	}

	lines = append(lines,
		fmt.Sprintf("Índice: %d/%d", m.FieldCursor+1, len(m.Plants)),
		fmt.Sprintf("Tipo:   %s", plantType),
		fmt.Sprintf("Estado: %s", plantStateLabel(slot.State)),
		fmt.Sprintf("Tempo:  %s", slotRemainingLabel(slot)),
	)
	return strings.Join(lines, "\n")
}

func progressBar(current, max, width int, pct float64) string {
	if max == 0 {
		return "[" + strings.Repeat("░", width) + "] 0/0"
	}

	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}

	bar := fmt.Sprintf("[%s%s] %d/%d",
		strings.Repeat("█", filled),
		strings.Repeat("░", width-filled),
		current, max)

	switch {
	case pct < 0.7:
		return greenStyle.Render(bar)
	case pct < 0.9:
		return yellowStyle.Render(bar)
	default:
		return redStyle.Render(bar)
	}
}

func renderMiniGrid(m model.Model, cols int) string {
	if cols <= 0 {
		cols = 4
	}

	var lines []string
	var row strings.Builder

	for i, p := range m.Plants {
		cell := string(slotRune(p.State))
		if i == m.FieldCursor {
			if m.FocusMode == model.FocusField {
				cell = cursorStyle.Render("[" + cell + "]")
			} else {
				cell = "[" + cell + "]"
			}
		} else {
			cell = " " + cell + " "
		}
		row.WriteString(cell)
		if (i+1)%cols == 0 {
			lines = append(lines, row.String())
			row.Reset()
		}
	}

	if row.Len() > 0 {
		lines = append(lines, row.String())
	}

	return strings.Join(lines, "\n")
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

type actionItem struct {
	label   string
	enabled bool
}

func buildActions(m model.Model) []actionItem {
	expandCost := engine.ExpandFieldCost(m.FieldSize)
	harvestCost := engine.HarvestUpgradeCost(m.HarvestLevel)
	label5 := fmt.Sprintf("[5] Config auto-venda  (atual: %d)", m.AutoSellThreshold)
	if m.InputMode && m.Cursor != 5 {
		label5 = fmt.Sprintf("[5] Novo threshold: %s_", m.InputBuffer)
	}
	label6 := fmt.Sprintf("[6] Auto-compra min      %d", m.AutoBuyMinimum)
	if m.InputMode && m.Cursor == 5 {
		label6 = fmt.Sprintf("[6] Auto-compra min: %s_", m.InputBuffer)
	}

	return []actionItem{
		{fmt.Sprintf("[1] Comprar semente     $%.0f", model.SeedCost), m.Money >= model.SeedCost},
		{fmt.Sprintf("[2] Expandir campo      $%.0f", expandCost), m.Money >= expandCost},
		{fmt.Sprintf("[3] Upgrade colheita    $%.0f", harvestCost), m.Money >= harvestCost},
		{"[4] Vender tudo          —", true},
		{label5, true},
		{label6, true},
	}
}

func renderActions(m model.Model) string {
	items := buildActions(m)

	var b strings.Builder
	b.WriteString(panelTitle("AÇÕES") + "\n")
	for i, item := range items {
		cursor := "  "
		if i == m.Cursor {
			cursor = cursorStyle.Render("▶ ")
		}

		s := lipgloss.NewStyle()
		if !item.enabled {
			s = faintStyle
		}

		b.WriteString(cursor + s.Render(item.label) + "\n")
	}
	b.WriteString(faintStyle.Render("  [Q] Salvar e sair\n"))
	b.WriteString(faintStyle.Render("  [Tab] Alternar foco menu/campo"))

	return b.String()
}

func renderLog(m model.Model, maxEntries int) string {
	var b strings.Builder
	b.WriteString(panelTitle("LOG") + "\n")

	entries := m.Log
	if maxEntries <= 0 {
		maxEntries = 5
	}
	if len(entries) > maxEntries {
		entries = entries[len(entries)-maxEntries:]
	}

	if len(entries) == 0 {
		b.WriteString(faintStyle.Render("  (sem eventos)"))
		return b.String()
	}

	for _, e := range entries {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			faintStyle.Render(e.Timestamp.Format("15:04:05")),
			e.Message))
	}

	return strings.TrimRight(b.String(), "\n")
}

func renderOfflineReport(r model.OfflineResult) string {
	h := int(r.Duration.Hours())
	min := int(r.Duration.Minutes()) % 60
	s := int(r.Duration.Seconds()) % 60

	lines := []string{
		panelTitle("BEM-VINDO DE VOLTA"),
		"",
		fmt.Sprintf("Ausente:    %02dh %02dm %02ds", h, min, s),
		fmt.Sprintf("Colheitas:  +%d", r.Harvests),
		fmt.Sprintf("Receita:    +$%.0f", r.Revenue),
		fmt.Sprintf("Eficiência: %.0f%%", r.Efficiency*100),
		"",
		cursorStyle.Render("ENTER → Coletar"),
	}

	return strings.Join(lines, "\n")
}

func summarizeField(plants []model.PlantSlot) (active, planted, growing, ready, empty, nextReady int) {
	nextReady = math.MaxInt
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
	if nextReady == math.MaxInt {
		nextReady = 0
	}
	return active, planted, growing, ready, empty, nextReady
}

func selectedSlot(m model.Model) (model.PlantSlot, bool) {
	if len(m.Plants) == 0 {
		return model.PlantSlot{}, false
	}
	idx := m.FieldCursor
	if idx < 0 {
		idx = 0
	}
	if idx >= len(m.Plants) {
		idx = len(m.Plants) - 1
	}
	return m.Plants[idx], true
}

func slotRemainingLabel(slot model.PlantSlot) string {
	switch slot.State {
	case model.PlantEmpty:
		return "—"
	case model.PlantReady:
		return "pronta"
	default:
		return fmt.Sprintf("%ds", slot.TicksRemaining)
	}
}

func plantStateLabel(state model.PlantState) string {
	switch state {
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func panelTitle(label string) string {
	return boldStyle.Render("╢ " + label + " ╟")
}

func renderBanner(m model.Model, _ int) string {
	lines := []string{
		"  ______   ___   ____  __  ___      ____  _____",
		" / ____/  /   | / __ \\/  |/  /     /  _/ / ___/",
		"/ /_     / /| |/ /_/ / /|_/ /_____ / /   \\__ \\ ",
		"\\__/    /_/ |_|\\____/_/  /_/_____//___/ /____/ ",
	}
	banner := greenStyle.Render(strings.Join(lines, "\n"))
	tagline := faintStyle.Render("painel operacional de fazenda idle")
	version := yellowStyle.Render("v0.2-dev")
	head := lipgloss.JoinHorizontal(lipgloss.Bottom, banner, "   ", version)
	return lipgloss.JoinVertical(lipgloss.Left, head, tagline)
}

func stackedPanelHeight(m model.Model) int {
	if m.ViewHeight >= 40 {
		return 12
	}
	return 10
}

func widePanelHeight(m model.Model) int {
	if m.ViewHeight <= 0 {
		return 18
	}
	height := m.ViewHeight - 12
	if height < 16 {
		return 16
	}
	if height > 24 {
		return 24
	}
	return height
}

func placeOnScreen(m model.Model, body string) string {
	if m.ViewWidth <= 0 || m.ViewHeight <= 0 {
		return body
	}
	return lipgloss.Place(m.ViewWidth, m.ViewHeight, lipgloss.Center, lipgloss.Top, body)
}

func visibleLogEntries(m model.Model) int {
	if m.ViewHeight >= 42 {
		return 8
	}
	if m.ViewHeight >= 32 {
		return 6
	}
	return 5
}
