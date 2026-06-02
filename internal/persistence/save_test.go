// internal/persistence/save_test.go
package persistence_test

import (
	"encoding/json"
	"farm-idle/internal/model"
	"farm-idle/internal/persistence"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoad_Roundtrip(t *testing.T) {
	m := persistence.DefaultModel()
	m.Money = 500
	m.Seeds = 3
	m.Day = 7
	m.HarvestLevel = 2
	m.StockByCrop = map[string]int{"Alface": 2, "Milho": 3}
	m.Stock = 5

	path := filepath.Join(t.TempDir(), "save.json")

	if err := persistence.Save(m, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := persistence.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Money != m.Money {
		t.Errorf("money: want %.0f, got %.0f", m.Money, loaded.Money)
	}
	if loaded.Seeds != m.Seeds {
		t.Errorf("seeds: want %d, got %d", m.Seeds, loaded.Seeds)
	}
	if loaded.Day != m.Day {
		t.Errorf("day: want %d, got %d", m.Day, loaded.Day)
	}
	if loaded.HarvestLevel != m.HarvestLevel {
		t.Errorf("harvest_level: want %d, got %d", m.HarvestLevel, loaded.HarvestLevel)
	}
	if loaded.Stock != 5 {
		t.Errorf("stock: want 5, got %d", loaded.Stock)
	}
	if loaded.StockByCrop["Alface"] != 2 || loaded.StockByCrop["Milho"] != 3 {
		t.Errorf("stock_by_crop: want Alface=2 Milho=3, got %+v", loaded.StockByCrop)
	}
	if loaded.FieldSize != m.FieldSize {
		t.Errorf("field_size: want %d, got %d", m.FieldSize, loaded.FieldSize)
	}
}

func TestLoad_NonExistentReturnsDefault(t *testing.T) {
	m, err := persistence.Load("/nonexistent/path/save.json")
	if err != nil {
		t.Fatalf("Load of non-existent file should not error, got: %v", err)
	}
	def := persistence.DefaultModel()
	if m.Money != def.Money {
		t.Errorf("money: want %.0f, got %.0f", def.Money, m.Money)
	}
	if m.FieldSize != def.FieldSize {
		t.Errorf("field_size: want %d, got %d", def.FieldSize, m.FieldSize)
	}
	if m.HarvestLevel != def.HarvestLevel {
		t.Errorf("harvest_level: want %d, got %d", def.HarvestLevel, m.HarvestLevel)
	}
}

func TestDefaultModel_HasCorrectInitialValues(t *testing.T) {
	m := persistence.DefaultModel()
	if m.Money != 100 {
		t.Errorf("money: want 100, got %.0f", m.Money)
	}
	if m.Seeds != 10 {
		t.Errorf("seeds: want 10, got %d", m.Seeds)
	}
	if m.FieldSize != 10 {
		t.Errorf("field_size: want 10, got %d", m.FieldSize)
	}
	if len(m.Plants) != 10 {
		t.Errorf("len(plants): want 10, got %d", len(m.Plants))
	}
	if m.HarvestLevel != 1 {
		t.Errorf("harvest_level: want 1, got %d", m.HarvestLevel)
	}
	if m.MoneySnapshot != 100 {
		t.Errorf("money_snapshot: want 100, got %.0f", m.MoneySnapshot)
	}
}

func TestLoad_LegacyZeroSnapshotUsesCurrentMoney(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	content := `{
  "money": 250,
  "seeds": 4,
  "stock": 0,
  "field_size": 10,
  "plants": [],
  "harvest_level": 1,
  "auto_sell_threshold": 5,
  "day": 1,
  "tick_count": 0,
  "last_save": "0001-01-01T00:00:00Z",
  "money_snapshot": 0
}`

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	m, err := persistence.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if m.MoneySnapshot != m.Money {
		t.Errorf("money_snapshot: want %.0f, got %.0f", m.Money, m.MoneySnapshot)
	}
}

func TestSaveLoad_RoundtripAutoBuyAndRevenueTracker(t *testing.T) {
	m := persistence.DefaultModel()
	m.SeedsPerPurchase = 12
	m.AutoBuyEnabled = true
	m.AutoBuyMinimum = 8
	m.AutoBuyMaxCashFraction = 0.45
	m.SelectedCrop = "Milho"
	m.RevenueTracker = model.RevenueTracker{
		Buckets: [model.TicksPerDay]float64{0: 10, 3: 7.5, 59: 2.5},
		Cursor:  3,
		Total:   20,
	}
	m.RecentRevenue = 20

	path := filepath.Join(t.TempDir(), "save.json")

	if err := persistence.Save(m, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := persistence.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.SeedsPerPurchase != m.SeedsPerPurchase {
		t.Fatalf("seeds_per_purchase: want %d, got %d", m.SeedsPerPurchase, loaded.SeedsPerPurchase)
	}
	if loaded.AutoBuyEnabled != m.AutoBuyEnabled {
		t.Fatalf("auto_buy_enabled: want %t, got %t", m.AutoBuyEnabled, loaded.AutoBuyEnabled)
	}
	if loaded.AutoBuyMinimum != m.AutoBuyMinimum {
		t.Fatalf("auto_buy_minimum: want %d, got %d", m.AutoBuyMinimum, loaded.AutoBuyMinimum)
	}
	if loaded.AutoBuyMaxCashFraction != m.AutoBuyMaxCashFraction {
		t.Fatalf("auto_buy_max_cash_fraction: want %.2f, got %.2f", m.AutoBuyMaxCashFraction, loaded.AutoBuyMaxCashFraction)
	}
	if loaded.SelectedCrop != m.SelectedCrop {
		t.Fatalf("selected_crop: want %q, got %q", m.SelectedCrop, loaded.SelectedCrop)
	}
	if loaded.RevenueTracker != m.RevenueTracker {
		t.Fatalf("revenue_tracker: want %+v, got %+v", m.RevenueTracker, loaded.RevenueTracker)
	}
	if loaded.RecentRevenue != loaded.RevenueTracker.Total {
		t.Fatalf("recent_revenue: want derived %.2f, got %.2f", loaded.RevenueTracker.Total, loaded.RecentRevenue)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	var saved map[string]any
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if _, exists := saved["recent_revenue"]; exists {
		t.Fatalf("recent_revenue should not be persisted")
	}
}

func TestLoad_LegacySaveDefaultsAutoBuyFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy-save.json")
	content := `{
  "money": 125,
  "seeds": 1,
  "stock": 2,
  "field_size": 10,
  "plants": [],
  "harvest_level": 1,
  "auto_sell_threshold": 5,
  "day": 3,
  "tick_count": 9,
  "last_save": "2026-05-31T12:00:00Z",
  "money_snapshot": 90
}`

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	loaded, err := persistence.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.SeedsPerPurchase != 5 {
		t.Fatalf("seeds_per_purchase: want 5, got %d", loaded.SeedsPerPurchase)
	}
	if loaded.AutoBuyMinimum != 5 {
		t.Fatalf("auto_buy_minimum: want 5, got %d", loaded.AutoBuyMinimum)
	}
	if loaded.AutoBuyMaxCashFraction != model.DefaultAutoBuyMaxCashFraction {
		t.Fatalf(
			"auto_buy_max_cash_fraction: want %.2f, got %.2f",
			model.DefaultAutoBuyMaxCashFraction,
			loaded.AutoBuyMaxCashFraction,
		)
	}
	if loaded.SelectedCrop != model.DefaultPlantType {
		t.Fatalf("selected_crop: want default %q, got %q", model.DefaultPlantType, loaded.SelectedCrop)
	}
	if loaded.StockByCrop[model.DefaultPlantType] != 2 {
		t.Fatalf("legacy stock_by_crop: want %s=2, got %+v", model.DefaultPlantType, loaded.StockByCrop)
	}
}
