package persistence

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"farm-idle/internal/model"
)

type SaveData struct {
	Money                  float64              `json:"money"`
	Seeds                  int                  `json:"seeds"`
	SeedsByCrop            map[string]int       `json:"seeds_by_crop"`
	Stock                  int                  `json:"stock"`
	StockByCrop            map[string]int       `json:"stock_by_crop"`
	FieldSize              int                  `json:"field_size"`
	Plants                 []model.PlantSlot    `json:"plants"`
	SeedsPerPurchase       int                  `json:"seeds_per_purchase"`
	AutoBuyEnabled         bool                 `json:"auto_buy_enabled"`
	AutoBuyMinimum         int                  `json:"auto_buy_minimum"`
	AutoBuyMaxCashFraction float64              `json:"auto_buy_max_cash_fraction"`
	SelectedCrop           string               `json:"selected_crop"`
	HarvestLevel           int                  `json:"harvest_level"`
	AutoSellThreshold      int                  `json:"auto_sell_threshold"`
	Day                    int                  `json:"day"`
	TickCount              int                  `json:"tick_count"`
	LastSave               time.Time            `json:"last_save"`
	MoneySnapshot          float64              `json:"money_snapshot"`
	RevenueTracker         model.RevenueTracker `json:"revenue_tracker"`
}

func DefaultModel() model.Model {
	plants := make([]model.PlantSlot, 10)
	for i := range plants {
		plants[i] = model.PlantSlot{State: model.PlantEmpty}
	}

	return model.Model{
		Money:                  100,
		Seeds:                  10,
		SeedsByCrop:            map[string]int{model.DefaultPlantType: 10},
		StockByCrop:            map[string]int{},
		FieldSize:              10,
		Plants:                 plants,
		SeedsPerPurchase:       5,
		AutoBuyMinimum:         5,
		AutoBuyMaxCashFraction: model.DefaultAutoBuyMaxCashFraction,
		SelectedCrop:           model.DefaultPlantType,
		HarvestLevel:           1,
		AutoSellThreshold:      model.DefaultAutoSellThreshold,
		Day:                    1,
		MoneySnapshot:          100,
	}
}

func Save(m model.Model, path string) error {
	seedsByCrop := normalizeSeedsByCrop(m.SeedsByCrop)
	seeds := sumSeedsByCrop(seedsByCrop)
	if seeds == 0 && m.Seeds > 0 {
		seeds = m.Seeds
		crop := model.CropByName(m.SelectedCrop)
		seedsByCrop = map[string]int{crop.Name: m.Seeds}
	}
	stockByCrop := normalizeStockByCrop(m.StockByCrop)
	stock := sumStockByCrop(stockByCrop)
	if stock == 0 && m.Stock > 0 {
		stock = m.Stock
		stockByCrop = map[string]int{model.DefaultPlantType: m.Stock}
	}

	data := SaveData{
		Money:                  m.Money,
		Seeds:                  seeds,
		SeedsByCrop:            seedsByCrop,
		Stock:                  stock,
		StockByCrop:            stockByCrop,
		FieldSize:              m.FieldSize,
		Plants:                 m.Plants,
		SeedsPerPurchase:       m.SeedsPerPurchase,
		AutoBuyEnabled:         m.AutoBuyEnabled,
		AutoBuyMinimum:         m.AutoBuyMinimum,
		AutoBuyMaxCashFraction: m.AutoBuyMaxCashFraction,
		SelectedCrop:           m.SelectedCrop,
		HarvestLevel:           m.HarvestLevel,
		AutoSellThreshold:      m.AutoSellThreshold,
		Day:                    m.Day,
		TickCount:              m.TickCount,
		LastSave:               m.LastSave,
		MoneySnapshot:          m.MoneySnapshot,
		RevenueTracker:         m.RevenueTracker,
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0o644)
}

func Load(path string) (model.Model, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultModel(), nil
		}
		return DefaultModel(), err
	}

	var data SaveData
	if err := json.Unmarshal(b, &data); err != nil {
		return DefaultModel(), err
	}

	m := DefaultModel()
	m.Money = data.Money
	m.Seeds = data.Seeds
	if len(data.SeedsByCrop) > 0 {
		m.SeedsByCrop = normalizeSeedsByCrop(data.SeedsByCrop)
		m.Seeds = sumSeedsByCrop(m.SeedsByCrop)
	} else if data.Seeds > 0 {
		legacyCrop := model.CropByName(data.SelectedCrop)
		m.SeedsByCrop = map[string]int{legacyCrop.Name: data.Seeds}
	}
	m.Stock = data.Stock
	if len(data.StockByCrop) > 0 {
		m.StockByCrop = normalizeStockByCrop(data.StockByCrop)
		m.Stock = sumStockByCrop(m.StockByCrop)
	} else if data.Stock > 0 {
		m.StockByCrop = map[string]int{model.DefaultPlantType: data.Stock}
	}
	if data.FieldSize > 0 {
		m.FieldSize = data.FieldSize
	}
	if len(data.Plants) > 0 {
		m.Plants = data.Plants
	}
	if data.SeedsPerPurchase > 0 {
		m.SeedsPerPurchase = data.SeedsPerPurchase
	}
	if data.AutoBuyEnabled {
		m.AutoBuyEnabled = data.AutoBuyEnabled
	}
	if data.AutoBuyMinimum > 0 {
		m.AutoBuyMinimum = data.AutoBuyMinimum
	}
	if data.AutoBuyMaxCashFraction > 0 {
		m.AutoBuyMaxCashFraction = data.AutoBuyMaxCashFraction
	}
	if data.SelectedCrop != "" {
		m.SelectedCrop = model.CropByName(data.SelectedCrop).Name
	}
	m.HarvestLevel = data.HarvestLevel
	if m.HarvestLevel == 0 {
		m.HarvestLevel = 1
	}
	if data.AutoSellThreshold >= 0 {
		m.AutoSellThreshold = data.AutoSellThreshold
	}
	if data.Day > 0 {
		m.Day = data.Day
	}
	m.TickCount = data.TickCount
	m.LastSave = data.LastSave
	m.MoneySnapshot = data.MoneySnapshot
	m.RevenueTracker = data.RevenueTracker
	if m.MoneySnapshot == 0 && m.TickCount == 0 {
		m.MoneySnapshot = m.Money
	}
	m.RecentRevenue = m.RevenueTracker.Total

	return m, nil
}

func normalizeStockByCrop(stocks map[string]int) map[string]int {
	out := make(map[string]int, len(stocks))
	for crop, qty := range stocks {
		if qty > 0 {
			out[model.CropByName(crop).Name] += qty
		}
	}
	return out
}

func normalizeSeedsByCrop(seeds map[string]int) map[string]int {
	out := make(map[string]int, len(seeds))
	for crop, qty := range seeds {
		if qty > 0 {
			out[model.CropByName(crop).Name] += qty
		}
	}
	return out
}

func sumStockByCrop(stocks map[string]int) int {
	total := 0
	for _, qty := range stocks {
		total += qty
	}
	return total
}

func sumSeedsByCrop(seeds map[string]int) int {
	total := 0
	for _, qty := range seeds {
		total += qty
	}
	return total
}
