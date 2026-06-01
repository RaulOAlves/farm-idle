package persistence

import (
	"encoding/json"
	"errors"
	"os"

	"farm-idle/internal/model"
)

func DefaultModel() model.Model {
	plants := make([]model.PlantSlot, 10)
	for i := range plants {
		plants[i] = model.PlantSlot{State: model.PlantEmpty}
	}

	return model.Model{
		Money:             100,
		Seeds:             10,
		FieldSize:         10,
		Plants:            plants,
		HarvestLevel:      1,
		AutoSellThreshold: model.DefaultAutoSellThreshold,
		Day:               1,
	}
}

func Save(m model.Model, path string) error {
	data := model.SaveData{
		Money:             m.Money,
		Seeds:             m.Seeds,
		Stock:             m.Stock,
		FieldSize:         m.FieldSize,
		Plants:            m.Plants,
		HarvestLevel:      m.HarvestLevel,
		AutoSellThreshold: m.AutoSellThreshold,
		Day:               m.Day,
		TickCount:         m.TickCount,
		LastSave:          m.LastSave,
		MoneySnapshot:     m.MoneySnapshot,
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

	var data model.SaveData
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

	return m, nil
}
