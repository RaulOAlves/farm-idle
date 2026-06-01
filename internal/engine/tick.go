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
