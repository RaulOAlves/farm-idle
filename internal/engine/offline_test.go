// internal/engine/offline_test.go
package engine_test

import (
	"testing"
	"time"
	"farm-idle/internal/engine"
	"farm-idle/internal/model"
)

func TestCalcOfflineProgress_ZeroTime(t *testing.T) {
	now := time.Now()
	m := model.Model{
		LastSave:          now,
		HarvestLevel:      1,
		AutoSellThreshold: 999,
		Plants:            []model.PlantSlot{},
	}
	_, report := engine.CalcOfflineProgress(m, now)
	if report.Harvests != 0 {
		t.Errorf("harvests: want 0, got %d", report.Harvests)
	}
	if report.Revenue != 0 {
		t.Errorf("revenue: want 0, got %.0f", report.Revenue)
	}
	if report.Duration != 0 {
		t.Errorf("duration: want 0, got %v", report.Duration)
	}
}

func TestCalcOfflineProgress_HarvestsAfterGrowTime(t *testing.T) {
	now := time.Now()
	// ausente tempo suficiente para 1 ciclo de crescimento completo
	past := now.Add(-time.Duration(model.GrowTicks+5) * time.Second)
	m := model.Model{
		LastSave:          past,
		HarvestLevel:      1,
		AutoSellThreshold: 999,
		Plants:            []model.PlantSlot{{State: model.PlantPlanted, TicksRemaining: model.GrowTicks}},
		Seeds:             0,
	}
	_, report := engine.CalcOfflineProgress(m, now)
	if report.Harvests < 1 {
		t.Errorf("harvests: want >= 1, got %d", report.Harvests)
	}
	if report.Revenue <= 0 {
		t.Errorf("revenue: want > 0, got %.0f", report.Revenue)
	}
}

func TestCalcOfflineProgress_CapAt24h(t *testing.T) {
	now := time.Now()
	past := now.Add(-48 * time.Hour)
	m := model.Model{
		LastSave:          past,
		HarvestLevel:      1,
		AutoSellThreshold: 999,
		Plants:            []model.PlantSlot{},
	}
	_, report := engine.CalcOfflineProgress(m, now)
	if report.Duration > 24*time.Hour {
		t.Errorf("duration: want <= 24h, got %v", report.Duration)
	}
}
