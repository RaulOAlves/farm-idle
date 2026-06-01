// internal/engine/tick_test.go
package engine_test

import (
	"testing"
	"farm-idle/internal/engine"
	"farm-idle/internal/model"
)

func TestTick_PlantsSeeds(t *testing.T) {
	m := model.Model{
		Seeds:             1,
		Plants:            []model.PlantSlot{{State: model.PlantEmpty}},
		HarvestLevel:      1,
		AutoSellThreshold: 999,
	}
	got := engine.Tick(m)
	if got.Seeds != 0 {
		t.Errorf("seeds: want 0, got %d", got.Seeds)
	}
	if got.Plants[0].State != model.PlantPlanted {
		t.Errorf("state: want %q, got %q", model.PlantPlanted, got.Plants[0].State)
	}
	if got.Plants[0].TicksRemaining != model.GrowTicks {
		t.Errorf("ticks: want %d, got %d", model.GrowTicks, got.Plants[0].TicksRemaining)
	}
}

func TestTick_PlantGrowsFromPlanted(t *testing.T) {
	m := model.Model{
		Plants:            []model.PlantSlot{{State: model.PlantPlanted, TicksRemaining: 5}},
		HarvestLevel:      1,
		AutoSellThreshold: 999,
	}
	got := engine.Tick(m)
	if got.Plants[0].State != model.PlantGrowing {
		t.Errorf("state: want %q, got %q", model.PlantGrowing, got.Plants[0].State)
	}
	if got.Plants[0].TicksRemaining != 4 {
		t.Errorf("ticks: want 4, got %d", got.Plants[0].TicksRemaining)
	}
}

func TestTick_GrowingBecomesReady(t *testing.T) {
	m := model.Model{
		Plants:            []model.PlantSlot{{State: model.PlantGrowing, TicksRemaining: 1}},
		HarvestLevel:      1,
		AutoSellThreshold: 999,
	}
	got := engine.Tick(m)
	if got.Plants[0].State != model.PlantReady {
		t.Errorf("state: want %q, got %q", model.PlantReady, got.Plants[0].State)
	}
}

func TestTick_HarvestsReadyPlant(t *testing.T) {
	m := model.Model{
		Plants:            []model.PlantSlot{{State: model.PlantReady}},
		HarvestLevel:      2,
		AutoSellThreshold: 999,
		Stock:             0,
	}
	got := engine.Tick(m)
	if got.Stock != 2 {
		t.Errorf("stock: want 2, got %d", got.Stock)
	}
	if got.Plants[0].State != model.PlantEmpty {
		t.Errorf("state: want %q, got %q", model.PlantEmpty, got.Plants[0].State)
	}
}

func TestTick_AutoSellExcess(t *testing.T) {
	m := model.Model{
		Plants:            []model.PlantSlot{},
		Stock:             8,
		AutoSellThreshold: 5,
		Money:             0,
		HarvestLevel:      1,
	}
	got := engine.Tick(m)
	if got.Stock != 5 {
		t.Errorf("stock: want 5, got %d", got.Stock)
	}
	if got.Money != 15 { // 3 * 5
		t.Errorf("money: want 15, got %.0f", got.Money)
	}
}

func TestTick_DayIncrements(t *testing.T) {
	m := model.Model{
		TickCount:         model.TicksPerDay - 1,
		Day:               1,
		HarvestLevel:      1,
		AutoSellThreshold: 999,
		Plants:            []model.PlantSlot{},
	}
	got := engine.Tick(m)
	if got.Day != 2 {
		t.Errorf("day: want 2, got %d", got.Day)
	}
}

func TestTick_OriginalNotMutated(t *testing.T) {
	m := model.Model{
		Seeds:             1,
		Plants:            []model.PlantSlot{{State: model.PlantEmpty}},
		HarvestLevel:      1,
		AutoSellThreshold: 999,
	}
	_ = engine.Tick(m)
	if m.Plants[0].State != model.PlantEmpty {
		t.Error("original model.Plants was mutated by Tick()")
	}
}
