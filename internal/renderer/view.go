package renderer

import (
	"fmt"
	"strings"

	"farm-idle/internal/engine"
	"farm-idle/internal/model"
	"github.com/charmbracelet/lipgloss"
)

const dashWidth = 52

var (
	outerStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	boldStyle   = lipgloss.NewStyle().Bold(true)
	faintStyle  = lipgloss.NewStyle().Faint(true)
	cursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

func View(m model.Model) string {
	if m.OfflineReport != nil {
		return outerStyle.Render(renderOfflineReport(*m.OfflineReport))
	}

	sections := []string{
		boldStyle.Render(fmt.Sprintf("FARM IDLE v0.1  —  Dia %d", m.Day)),
		renderStatusLine(m),
		sep(),
		lipgloss.JoinHorizontal(lipgloss.Top, renderResources(m), "    ", renderField(m)),
		sep(),
		renderActions(m),
		sep(),
		renderLog(m),
	}

	return outerStyle.Render(strings.Join(sections, "\n"))
}

func sep() string {
	return strings.Repeat("─", dashWidth)
}

func renderStatusLine(m model.Model) string {
	profit := m.Money - m.MoneySnapshot
	return fmt.Sprintf("Lucro/min: $%.0f  │  Auto-venda: ≤%d", profit, m.AutoSellThreshold)
}

func renderResources(m model.Model) string {
	lines := []string{
		boldStyle.Render("RECURSOS"),
		fmt.Sprintf("💰 Dinheiro:  $%.0f", m.Money),
		fmt.Sprintf("🌱 Sementes:  %d", m.Seeds),
		fmt.Sprintf("📦 Estoque:   %d", m.Stock),
		fmt.Sprintf("📈 Nível col: %d", m.HarvestLevel),
	}

	return strings.Join(lines, "\n")
}

func renderField(m model.Model) string {
	active := countActive(m)
	pct := 0.0
	if m.FieldSize > 0 {
		pct = float64(active) / float64(m.FieldSize)
	}

	var planted, growing, ready, empty int
	for _, p := range m.Plants {
		switch p.State {
		case model.PlantPlanted:
			planted++
		case model.PlantGrowing:
			growing++
		case model.PlantReady:
			ready++
		default:
			empty++
		}
	}

	lines := []string{
		boldStyle.Render("CAMPO"),
		progressBar(active, m.FieldSize, 20, pct),
		fmt.Sprintf("Plantadas: %d  Crescendo: %d", planted, growing),
		fmt.Sprintf("Prontas:   %d  Vazias:    %d", ready, empty),
	}

	return strings.Join(lines, "\n")
}

func countActive(m model.Model) int {
	n := 0
	for _, p := range m.Plants {
		if p.State != model.PlantEmpty {
			n++
		}
	}
	return n
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

type actionItem struct {
	label   string
	enabled bool
}

func buildActions(m model.Model) []actionItem {
	expandCost := engine.ExpandFieldCost(m.FieldSize)
	label5 := fmt.Sprintf("[5] Config auto-venda  (atual: %d)", m.AutoSellThreshold)
	if m.InputMode {
		label5 = fmt.Sprintf("[5] Novo threshold: %s_", m.InputBuffer)
	}

	return []actionItem{
		{fmt.Sprintf("[1] Comprar semente     $%.0f", model.SeedCost), m.Money >= model.SeedCost},
		{fmt.Sprintf("[2] Expandir campo      $%.0f", expandCost), m.Money >= expandCost},
		{fmt.Sprintf("[3] Upgrade colheita    $%.0f", model.HarvestUpgradeCost), m.Money >= model.HarvestUpgradeCost},
		{"[4] Vender tudo          —", true},
		{label5, true},
	}
}

func renderActions(m model.Model) string {
	items := buildActions(m)

	var b strings.Builder
	b.WriteString(boldStyle.Render("AÇÕES") + "\n")
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
	b.WriteString(faintStyle.Render("  [Q] Salvar e sair"))

	return b.String()
}

func renderLog(m model.Model) string {
	var b strings.Builder
	b.WriteString(boldStyle.Render("LOG") + "\n")

	entries := m.Log
	if len(entries) > 5 {
		entries = entries[len(entries)-5:]
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
		boldStyle.Render("BEM-VINDO DE VOLTA"),
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
