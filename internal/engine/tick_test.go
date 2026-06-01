// internal/engine/tick_test.go
package engine_test

import (
	"farm-idle/internal/engine"
	"farm-idle/internal/model"
	"testing"
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

func TestTick_AutoSellSellsAllAtOrAboveThreshold(t *testing.T) {
	m := model.Model{
		Plants:            []model.PlantSlot{},
		Stock:             8,
		AutoSellThreshold: 5,
		Money:             0,
		HarvestLevel:      1,
	}
	got := engine.Tick(m)
	if got.Stock != 0 {
		t.Errorf("stock: want 0, got %d", got.Stock)
	}
	if got.Money != 40 { // 8 * 5
		t.Errorf("money: want 40, got %.0f", got.Money)
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
		Money:             25,
		Seeds:             1,
		Plants:            []model.PlantSlot{{State: model.PlantEmpty, TicksRemaining: 7}},
		HarvestLevel:      1,
		AutoSellThreshold: 999,
	}
	_ = engine.Tick(m)
	if m.Seeds != 1 {
		t.Errorf("original seeds mutated: want 1, got %d", m.Seeds)
	}
	if m.Money != 25 {
		t.Errorf("original money mutated: want 25, got %.0f", m.Money)
	}
	if m.Plants[0].State != model.PlantEmpty {
		t.Error("original model.Plants was mutated by Tick()")
	}
	if m.Plants[0].TicksRemaining != 7 {
		t.Errorf("original ticks mutated: want 7, got %d", m.Plants[0].TicksRemaining)
	}
}

func TestBuySeeds_DoesNotMutateCallerLogBackingArray(t *testing.T) {
	backing := make([]model.LogEntry, 1, 4)
	backing[0] = model.LogEntry{Message: "original"}
	exposedBacking := backing[:cap(backing)]
	exposedBacking[1] = model.LogEntry{Message: "sentinel"}

	m := model.Model{
		Money: 20,
		Log:   backing,
	}

	got, err := engine.BuySeeds(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Log) != 1 {
		t.Fatalf("original log len mutated: want 1, got %d", len(m.Log))
	}
	if m.Log[0].Message != "original" {
		t.Fatalf("original log entry mutated: want %q, got %q", "original", m.Log[0].Message)
	}
	if exposedBacking[1].Message != "sentinel" {
		t.Fatalf("caller backing array mutated: want %q, got %q", "sentinel", exposedBacking[1].Message)
	}
	if len(got.Log) != 2 {
		t.Fatalf("new log len: want 2, got %d", len(got.Log))
	}
	if got.Log[1].Message != "🌱 Comprou 5 sementes" {
		t.Fatalf("new log message: want %q, got %q", "🌱 Comprou 5 sementes", got.Log[1].Message)
	}
}

func TestBuySeeds_SuccessDefaultsToBatchOfFive(t *testing.T) {
	m := model.Model{Money: 20}
	got, err := engine.BuySeeds(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Seeds != 5 {
		t.Errorf("seeds: want 5, got %d", got.Seeds)
	}
	if got.Money != 10 {
		t.Errorf("money: want 10, got %.0f", got.Money)
	}
}

func TestBuySeeds_SuccessBuysBatch(t *testing.T) {
	m := model.Model{Money: 100, Seeds: 0, SeedsPerPurchase: 5}

	got, err := engine.BuySeeds(m)
	if err != nil {
		t.Fatalf("BuySeeds returned error: %v", err)
	}
	if got.Seeds != 5 {
		t.Fatalf("seeds: want 5, got %d", got.Seeds)
	}
	if got.Money != 90 {
		t.Fatalf("money: want 90, got %.0f", got.Money)
	}
}

func TestBuySeeds_InsufficientFunds(t *testing.T) {
	m := model.Model{Money: 5}
	_, err := engine.BuySeeds(m)
	if err != engine.ErrInsufficientFunds {
		t.Errorf("want ErrInsufficientFunds, got %v", err)
	}
}

func TestExpandField_Success(t *testing.T) {
	m := model.Model{
		Money:     200,
		FieldSize: 10,
		Plants:    make([]model.PlantSlot, 10),
	}
	got, err := engine.ExpandField(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.FieldSize != 15 {
		t.Errorf("field_size: want 15, got %d", got.FieldSize)
	}
	if len(got.Plants) != 15 {
		t.Errorf("len(plants): want 15, got %d", len(got.Plants))
	}
	if got.Money != 100 {
		t.Errorf("money: want 100, got %.0f", got.Money)
	}
}

func TestExpandField_ScalingCost(t *testing.T) {
	m := model.Model{
		Money:     10000,
		FieldSize: 20,
		Plants:    make([]model.PlantSlot, 20),
	}
	got, err := engine.ExpandField(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// custo = 100 * (20/10) = 200
	if got.Money != 9800 {
		t.Errorf("money: want 9800, got %.0f", got.Money)
	}
}

func TestExpandField_InsufficientFunds(t *testing.T) {
	m := model.Model{Money: 50, FieldSize: 10, Plants: make([]model.PlantSlot, 10)}
	_, err := engine.ExpandField(m)
	if err != engine.ErrInsufficientFunds {
		t.Errorf("want ErrInsufficientFunds, got %v", err)
	}
}

func TestUpgradeHarvest_Success(t *testing.T) {
	m := model.Model{Money: 300, HarvestLevel: 1}
	got, err := engine.UpgradeHarvest(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.HarvestLevel != 2 {
		t.Errorf("harvest_level: want 2, got %d", got.HarvestLevel)
	}
	if got.Money != 50 {
		t.Errorf("money: want 50, got %.0f", got.Money)
	}
}

func TestSellAll_SellsStock(t *testing.T) {
	m := model.Model{Stock: 4, Money: 0}
	got, earned := engine.SellAll(m)
	if got.Stock != 0 {
		t.Errorf("stock: want 0, got %d", got.Stock)
	}
	if earned != 20 {
		t.Errorf("earned: want 20, got %.0f", earned)
	}
	if got.Money != 20 {
		t.Errorf("money: want 20, got %.0f", got.Money)
	}
}

func TestSellAll_EmptyStock(t *testing.T) {
	m := model.Model{Stock: 0, Money: 50}
	got, earned := engine.SellAll(m)
	if earned != 0 {
		t.Errorf("earned: want 0, got %.0f", earned)
	}
	if got.Money != 50 {
		t.Errorf("money: want 50, got %.0f", got.Money)
	}
}

func TestSetAutoSellThreshold(t *testing.T) {
	m := model.Model{AutoSellThreshold: 5}
	got := engine.SetAutoSellThreshold(m, 3)
	if got.AutoSellThreshold != 3 {
		t.Errorf("threshold: want 3, got %d", got.AutoSellThreshold)
	}
}

func TestExpandFieldCost(t *testing.T) {
	tests := []struct {
		fieldSize int
		wantCost  float64
	}{
		{10, 100},
		{20, 200},
		{5, 100}, // fieldSize < 10 → multiplier = 1
	}
	for _, tc := range tests {
		got := engine.ExpandFieldCost(tc.fieldSize)
		if got != tc.wantCost {
			t.Errorf("ExpandFieldCost(%d): want %.0f, got %.0f", tc.fieldSize, tc.wantCost, got)
		}
	}
}

func TestTick_UpdatesMoneySnapshot(t *testing.T) {
	m := model.Model{
		TickCount:         model.TicksPerDay - 1,
		Day:               1,
		Money:             150,
		HarvestLevel:      1,
		AutoSellThreshold: 999,
		Plants:            []model.PlantSlot{},
	}
	got := engine.Tick(m)
	if got.MoneySnapshot != 150 {
		t.Errorf("MoneySnapshot: want 150, got %.0f", got.MoneySnapshot)
	}
}

func TestTick_RevenueTrackerExpiresOldBucketAfter60Ticks(t *testing.T) {
	m := model.Model{
		Money:             0,
		Stock:             5,
		AutoSellThreshold: 5,
		Plants:            []model.PlantSlot{},
	}

	got := engine.Tick(m)
	for i := 0; i < model.TicksPerDay-1; i++ {
		got = engine.Tick(got)
	}

	if got.RecentRevenue != model.StockValue*5 {
		t.Fatalf("recent revenue after first window: want %.0f, got %.0f", model.StockValue*5, got.RecentRevenue)
	}

	got = engine.Tick(got)
	if got.RecentRevenue != 0 {
		t.Fatalf("recent revenue after bucket expiry: want 0, got %.0f", got.RecentRevenue)
	}
}

func TestTick_AutoSellTriggersAtExactThreshold(t *testing.T) {
	m := model.Model{
		Money:             0,
		Stock:             5,
		AutoSellThreshold: 5,
		Plants:            []model.PlantSlot{},
	}

	got := engine.Tick(m)
	if got.Stock != 0 {
		t.Fatalf("stock after sell: want 0, got %d", got.Stock)
	}
	if got.Money != model.StockValue*5 {
		t.Fatalf("money after sell: want %.0f, got %.0f", model.StockValue*5, got.Money)
	}
}

func TestTick_AutoBuyUsesConfiguredMinimum(t *testing.T) {
	m := model.Model{
		Money:                  100,
		Seeds:                  4,
		SeedsPerPurchase:       5,
		AutoBuyEnabled:         true,
		AutoBuyMinimum:         5,
		AutoBuyMaxCashFraction: 0.30,
		AutoSellThreshold:      999,
		Plants:                 []model.PlantSlot{},
	}

	got := engine.Tick(m)

	if got.Seeds != 9 {
		t.Fatalf("seeds after auto-buy: want 9, got %d", got.Seeds)
	}
	if got.Money != 90 {
		t.Fatalf("money after auto-buy: want 90, got %.0f", got.Money)
	}
}
