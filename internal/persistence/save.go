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
	Stock                  int                  `json:"stock"`
	FieldSize              int                  `json:"field_size"`
	Plants                 []model.PlantSlot    `json:"plants"`
	SeedsPerPurchase       int                  `json:"seeds_per_purchase"`
	AutoBuyEnabled         bool                 `json:"auto_buy_enabled"`
	AutoBuyMinimum         int                  `json:"auto_buy_minimum"`
	AutoBuyMaxCashFraction float64              `json:"auto_buy_max_cash_fraction"`
	HarvestLevel           int                  `json:"harvest_level"`
	AutoSellThreshold      int                  `json:"auto_sell_threshold"`
	Day                    int                  `json:"day"`
	TickCount              int                  `json:"tick_count"`
	LastSave               time.Time            `json:"last_save"`
	MoneySnapshot          float64              `json:"money_snapshot"`
	RevenueTracker         model.RevenueTracker `json:"revenue_tracker"`
	RecentRevenue          float64              `json:"recent_revenue"`
}

func DefaultModel() model.Model {
	plants := make([]model.PlantSlot, 10)
	for i := range plants {
		plants[i] = model.PlantSlot{State: model.PlantEmpty}
	}

	return model.Model{
		Money:                  100,
		Seeds:                  10,
		FieldSize:              10,
		Plants:                 plants,
		SeedsPerPurchase:       5,
		AutoBuyMinimum:         5,
		AutoBuyMaxCashFraction: model.DefaultAutoBuyMaxCashFraction,
		HarvestLevel:           1,
		AutoSellThreshold:      model.DefaultAutoSellThreshold,
		Day:                    1,
		MoneySnapshot:          100,
	}
}

func Save(m model.Model, path string) error {
	data := SaveData{
		Money:                  m.Money,
		Seeds:                  m.Seeds,
		Stock:                  m.Stock,
		FieldSize:              m.FieldSize,
		Plants:                 m.Plants,
		SeedsPerPurchase:       m.SeedsPerPurchase,
		AutoBuyEnabled:         m.AutoBuyEnabled,
		AutoBuyMinimum:         m.AutoBuyMinimum,
		AutoBuyMaxCashFraction: m.AutoBuyMaxCashFraction,
		HarvestLevel:           m.HarvestLevel,
		AutoSellThreshold:      m.AutoSellThreshold,
		Day:                    m.Day,
		TickCount:              m.TickCount,
		LastSave:               m.LastSave,
		MoneySnapshot:          m.MoneySnapshot,
		RevenueTracker:         m.RevenueTracker,
		RecentRevenue:          m.RecentRevenue,
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
	m.Stock = data.Stock
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
	m.RecentRevenue = data.RecentRevenue
	if m.MoneySnapshot == 0 && m.TickCount == 0 {
		m.MoneySnapshot = m.Money
	}
	if m.RecentRevenue == 0 && m.RevenueTracker.Total > 0 {
		m.RecentRevenue = m.RevenueTracker.Total
	}

	return m, nil
}
