// internal/engine/offline.go
package engine

import (
	"time"
	"farm-idle/internal/model"
)

const maxOfflineTicks = 24 * 60 * 60 // 24h em ticks (1 tick = 1 segundo)

func CalcOfflineProgress(m model.Model, now time.Time) (model.Model, model.OfflineResult) {
	elapsed := now.Sub(m.LastSave)
	if elapsed <= 0 {
		return m, model.OfflineResult{}
	}

	ticks := int(elapsed.Seconds())
	if ticks > maxOfflineTicks {
		ticks = maxOfflineTicks
	}

	moneyBefore := m.Money
	stockBefore := m.Stock

	for i := 0; i < ticks; i++ {
		m = Tick(m)
	}

	// revenue = dinheiro ganho + valor do estoque acumulado
	moneyGained := m.Money - moneyBefore
	stockValueGained := float64(m.Stock-stockBefore) * model.StockValue
	revenue := moneyGained + stockValueGained

	// harvests ≈ total produzido / valor por unidade
	harvests := int(revenue / model.StockValue)

	// eficiência: harvests reais vs máximo teórico
	maxPossible := 0
	if model.GrowTicks > 0 {
		maxPossible = (m.FieldSize * ticks) / model.GrowTicks
	}
	efficiency := 0.0
	if maxPossible > 0 {
		efficiency = float64(harvests) / float64(maxPossible)
		if efficiency > 1.0 {
			efficiency = 1.0
		}
	}

	return m, model.OfflineResult{
		Duration:   time.Duration(ticks) * time.Second,
		Harvests:   harvests,
		Revenue:    revenue,
		Efficiency: efficiency,
	}
}
