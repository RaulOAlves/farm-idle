// internal/engine/tick.go
package engine

import (
	"errors"
	"fmt"
	"time"

	"farm-idle/internal/model"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

func Tick(m model.Model) model.Model {
	// copia Plants para não mutar o backing array do caller
	plants := make([]model.PlantSlot, len(m.Plants))
	copy(plants, m.Plants)
	m.Plants = plants

	for i := range m.Plants {
		switch m.Plants[i].State {
		case model.PlantEmpty:
			if m.Seeds > 0 {
				m.Seeds--
				m.Plants[i].State = model.PlantPlanted
				m.Plants[i].TicksRemaining = model.GrowTicks
			}
		case model.PlantPlanted:
			m.Plants[i].State = model.PlantGrowing
			m.Plants[i].TicksRemaining--
			if m.Plants[i].TicksRemaining <= 0 {
				m.Plants[i].State = model.PlantReady
			}
		case model.PlantGrowing:
			m.Plants[i].TicksRemaining--
			if m.Plants[i].TicksRemaining <= 0 {
				m.Plants[i].State = model.PlantReady
			}
		case model.PlantReady:
			m.Stock += m.HarvestLevel
			m.Plants[i].State = model.PlantEmpty
			m = addLog(m, fmt.Sprintf("🌾 Colheita: +%d estoques", m.HarvestLevel))
		}
	}

	if m.Stock > m.AutoSellThreshold {
		toSell := m.Stock - m.AutoSellThreshold
		earned := float64(toSell) * model.StockValue
		m.Stock -= toSell
		m.Money += earned
		m = addLog(m, fmt.Sprintf("💰 Auto-venda: +$%.0f", earned))
	}

	m.TickCount++
	if m.TickCount%model.TicksPerDay == 0 {
		m.Day++
		m.MoneySnapshot = m.Money
	}

	return m
}

func addLog(m model.Model, msg string) model.Model {
	entry := model.LogEntry{Timestamp: time.Now(), Message: msg}
	m.Log = append(m.Log, entry)
	if len(m.Log) > model.MaxLogEntries {
		m.Log = m.Log[len(m.Log)-model.MaxLogEntries:]
	}
	return m
}

// ExpandFieldCost é exportado para uso no renderer.
func ExpandFieldCost(fieldSize int) float64 {
	multiplier := fieldSize / 10
	if multiplier < 1 {
		multiplier = 1
	}
	return model.FieldExpandBaseCost * float64(multiplier)
}

func BuySeeds(m model.Model) (model.Model, error) {
	if m.Money < model.SeedCost {
		return m, ErrInsufficientFunds
	}
	m.Money -= model.SeedCost
	m.Seeds++
	return addLog(m, "🌱 Comprou 1 semente"), nil
}

func ExpandField(m model.Model) (model.Model, error) {
	cost := ExpandFieldCost(m.FieldSize)
	if m.Money < cost {
		return m, ErrInsufficientFunds
	}
	m.Money -= cost
	newSlots := make([]model.PlantSlot, 5)
	for i := range newSlots {
		newSlots[i] = model.PlantSlot{State: model.PlantEmpty}
	}
	m.Plants = append(m.Plants, newSlots...)
	m.FieldSize += 5
	return addLog(m, fmt.Sprintf("🚜 Campo: +5 slots ($%.0f)", cost)), nil
}

func UpgradeHarvest(m model.Model) (model.Model, error) {
	if m.Money < model.HarvestUpgradeCost {
		return m, ErrInsufficientFunds
	}
	m.Money -= model.HarvestUpgradeCost
	m.HarvestLevel++
	return addLog(m, fmt.Sprintf("📈 Colheita nível %d", m.HarvestLevel)), nil
}

func SellAll(m model.Model) (model.Model, float64) {
	if m.Stock == 0 {
		return m, 0
	}
	earned := float64(m.Stock) * model.StockValue
	m = addLog(m, fmt.Sprintf("💰 Venda total: +$%.0f (%d estoques)", earned, m.Stock))
	m.Money += earned
	m.Stock = 0
	return m, earned
}

func SetAutoSellThreshold(m model.Model, v int) model.Model {
	if v < 0 {
		v = 0
	}
	m.AutoSellThreshold = v
	if v == 0 {
		return addLog(m, "⚙ Auto-venda desativada")
	}
	return addLog(m, fmt.Sprintf("⚙ Auto-venda: threshold=%d", v))
}
