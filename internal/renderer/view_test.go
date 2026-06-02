package renderer

import (
	"regexp"
	"strings"
	"testing"

	"farm-idle/internal/model"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestView_ShowsReceitaMinAndAutoBuy(t *testing.T) {
	view := plainView(model.Model{
		Day:               3,
		Money:             150,
		MoneySnapshot:     999,
		RecentRevenue:     42,
		AutoSellThreshold: 7,
		AutoBuyEnabled:    true,
		AutoBuyMinimum:    5,
		FieldSize:         4,
		Plants:            []model.PlantSlot{{State: model.PlantEmpty}},
	})

	mustContain(t, view, "Receita/min $42")
	mustContain(t, view, "Auto-venda ≥7")
	mustContain(t, view, "Auto-compra min 5")
	mustContain(t, view, "painel operacional de fazenda idle")
	if strings.Contains(view, "Lucro/min") {
		t.Fatalf("view should not contain old label %q\nview:\n%s", "Lucro/min", view)
	}
}

func TestView_ShowsMiniGridAndAction6State(t *testing.T) {
	view := plainView(model.Model{
		Day:               1,
		RecentRevenue:     12,
		AutoSellThreshold: 5,
		AutoBuyEnabled:    true,
		AutoBuyMinimum:    8,
		FieldSize:         5,
		Plants: []model.PlantSlot{
			{State: model.PlantEmpty},
			{State: model.PlantPlanted},
			{State: model.PlantGrowing},
			{State: model.PlantReady},
			{State: model.PlantEmpty},
		},
	})

	mustContain(t, view, "[□] ·  ▓  █")
	mustContain(t, view, "[6] Auto-compra min")
	mustContain(t, view, "8")
	mustContain(t, view, "Próxima:")

	editView := plainView(model.Model{
		Day:               1,
		RecentRevenue:     12,
		AutoSellThreshold: 5,
		AutoBuyEnabled:    true,
		AutoBuyMinimum:    8,
		FieldSize:         5,
		Plants:            []model.PlantSlot{{State: model.PlantEmpty}},
		Cursor:            5,
		InputMode:         true,
		InputBuffer:       "11",
	})

	mustContain(t, editView, "[6] Auto-compra min: 11_")
}

func TestView_ShowsSelectedSlotDetailsAndFieldFocus(t *testing.T) {
	view := plainView(model.Model{
		Day:         2,
		ViewWidth:   110,
		FocusMode:   model.FocusField,
		FieldCursor: 1,
		FieldSize:   4,
		Plants: []model.PlantSlot{
			{State: model.PlantEmpty},
			{State: model.PlantGrowing, TicksRemaining: 9, PlantType: model.DefaultPlantType},
			{State: model.PlantReady, PlantType: model.DefaultPlantType},
			{State: model.PlantEmpty},
		},
	})

	mustContain(t, view, "Foco campo")
	mustContain(t, view, "SLOT")
	mustContain(t, view, "Tipo:   Trigo")
	mustContain(t, view, "Estado: crescendo")
	mustContain(t, view, "Tempo:  9s")
}

func plainView(m model.Model) string {
	return ansiPattern.ReplaceAllString(View(m), "")
}

func mustContain(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("view missing %q\nview:\n%s", want, got)
	}
}
