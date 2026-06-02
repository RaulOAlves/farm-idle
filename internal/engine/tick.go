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
	m = normalizeStock(m)
	m = normalizeSeeds(m)
	advanceRevenueWindow(&m)

	if shouldAutoBuy(m) {
		var err error
		m, err = BuySeeds(m)
		if err == nil {
			m = addLog(m, fmt.Sprintf("⚙ Auto-compra: +%d sementes", m.SeedsPerPurchase))
		}
	}

	// copia Plants para não mutar o backing array do caller
	plants := make([]model.PlantSlot, len(m.Plants))
	copy(plants, m.Plants)
	m.Plants = plants

	for i := range m.Plants {
		switch m.Plants[i].State {
		case model.PlantEmpty:
			crop := model.CropByName(m.SelectedCrop)
			if m.SeedsByCrop[crop.Name] > 0 {
				m.SeedsByCrop[crop.Name]--
				m.Seeds = totalSeeds(m.SeedsByCrop)
				m.Plants[i].State = model.PlantPlanted
				m.Plants[i].TicksRemaining = crop.GrowTicks
				m.Plants[i].PlantType = crop.Name
			}
		case model.PlantPlanted:
			m, i = advancePlantedSlot(m, i)
		case model.PlantGrowing:
			m, i = advancePlantedSlot(m, i)
		case model.PlantReady:
			m, i = advancePlantedSlot(m, i)
		}
	}

	if m.AutoSellThreshold > 0 && m.Stock >= m.AutoSellThreshold {
		var earned float64
		m, earned = sellStock(m, "Auto-venda")
		if earned > 0 {
			recordRevenue(&m, earned)
		}
	}

	m.TickCount++
	if m.TickCount%model.TicksPerDay == 0 {
		m.Day++
		m.MoneySnapshot = m.Money
	}

	m.RecentRevenue = m.RevenueTracker.Total
	return m
}

func addLog(m model.Model, msg string) model.Model {
	entry := model.LogEntry{Timestamp: time.Now(), Message: msg}
	logs := make([]model.LogEntry, len(m.Log), len(m.Log)+1)
	copy(logs, m.Log)
	m.Log = append(logs, entry)
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

// HarvestUpgradeCost cresce por nível para segurar aceleração econômica cedo demais.
func HarvestUpgradeCost(level int) float64 {
	if level < 1 {
		level = 1
	}
	return model.HarvestUpgradeBaseCost * float64(level)
}

func BuySeeds(m model.Model) (model.Model, error) {
	m = normalizeSeeds(m)
	if m.Money < model.SeedCost {
		return m, ErrInsufficientFunds
	}
	if m.SeedsPerPurchase <= 0 {
		m.SeedsPerPurchase = 5
	}
	crop := model.CropByName(m.SelectedCrop)
	m.Money -= model.SeedCost
	m.SeedsByCrop[crop.Name] += m.SeedsPerPurchase
	m.Seeds = totalSeeds(m.SeedsByCrop)
	return addLog(m, fmt.Sprintf("🌱 Comprou %d sementes de %s", m.SeedsPerPurchase, crop.Name)), nil
}

func ExpandField(m model.Model) (model.Model, error) {
	cost := ExpandFieldCost(m.FieldSize)
	if m.Money < cost {
		return m, ErrInsufficientFunds
	}
	m.Money -= cost
	plants := make([]model.PlantSlot, len(m.Plants)+5)
	copy(plants, m.Plants)
	for i := len(m.Plants); i < len(plants); i++ {
		plants[i] = model.PlantSlot{State: model.PlantEmpty}
	}
	m.Plants = plants
	m.FieldSize += 5
	return addLog(m, fmt.Sprintf("🚜 Campo: +5 slots ($%.0f)", cost)), nil
}

func UpgradeHarvest(m model.Model) (model.Model, error) {
	cost := HarvestUpgradeCost(m.HarvestLevel)
	if m.Money < cost {
		return m, ErrInsufficientFunds
	}
	m.Money -= cost
	m.HarvestLevel++
	return addLog(m, fmt.Sprintf("📈 Colheita nível %d ($%.0f)", m.HarvestLevel, cost)), nil
}

func SellAll(m model.Model) (model.Model, float64) {
	m = normalizeStock(m)
	if m.Stock == 0 {
		return m, 0
	}
	var earned float64
	m, earned = sellStock(m, "Venda total")
	if earned > 0 {
		recordRevenue(&m, earned)
	}
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

func CycleSelectedCrop(m model.Model) model.Model {
	next := model.NextCrop(m.SelectedCrop)
	m.SelectedCrop = next.Name
	return addLog(m, fmt.Sprintf("🌿 Cultura ativa: %s (%ds)", next.Name, next.GrowTicks))
}

func advancePlantedSlot(m model.Model, i int) (model.Model, int) {
	m.Plants[i].TicksRemaining--
	if m.Plants[i].TicksRemaining <= 0 {
		return harvestSlot(m, i), i
	}

	crop := model.CropByName(m.Plants[i].PlantType)
	m.Plants[i].State = growthStage(crop, m.Plants[i].TicksRemaining)
	return m, i
}

func growthStage(crop model.CropProfile, ticksRemaining int) model.PlantState {
	if ticksRemaining <= 0 {
		return model.PlantReady
	}
	firstStageEnd := (crop.GrowTicks * 2) / 3
	secondStageEnd := crop.GrowTicks / 3
	switch {
	case ticksRemaining > firstStageEnd:
		return model.PlantPlanted
	case ticksRemaining > secondStageEnd:
		return model.PlantGrowing
	default:
		return model.PlantReady
	}
}

func harvestSlot(m model.Model, i int) model.Model {
	crop := model.CropByName(m.Plants[i].PlantType)
	yield := crop.Yield
	if yield <= 0 {
		yield = 1
	}
	harvested := yield * m.HarvestLevel
	m.StockByCrop[crop.Name] += harvested
	m.Stock = totalStock(m.StockByCrop)
	m.Plants[i].State = model.PlantEmpty
	m.Plants[i].TicksRemaining = 0
	m.Plants[i].PlantType = ""
	return addLog(m, fmt.Sprintf("🌾 Colheita %s: +%d estoques", crop.Name, harvested))
}

func shouldAutoBuy(m model.Model) bool {
	if !m.AutoBuyEnabled || selectedCropSeeds(m) >= m.AutoBuyMinimum {
		return false
	}
	if m.AutoBuyMaxCashFraction <= 0 {
		m.AutoBuyMaxCashFraction = model.DefaultAutoBuyMaxCashFraction
	}
	return model.SeedCost <= m.Money && model.SeedCost <= m.Money*m.AutoBuyMaxCashFraction
}

func advanceRevenueWindow(m *model.Model) {
	next := (m.RevenueTracker.Cursor + 1) % len(m.RevenueTracker.Buckets)
	m.RevenueTracker.Cursor = next
	m.RevenueTracker.Total -= m.RevenueTracker.Buckets[next]
	m.RevenueTracker.Buckets[next] = 0
	if m.RevenueTracker.Total < 0 {
		m.RevenueTracker.Total = 0
	}
}

func recordRevenue(m *model.Model, earned float64) {
	idx := m.RevenueTracker.Cursor
	m.RevenueTracker.Buckets[idx] += earned
	m.RevenueTracker.Total += earned
	m.RecentRevenue = m.RevenueTracker.Total
}

func normalizeStock(m model.Model) model.Model {
	stocks := make(map[string]int, len(m.StockByCrop)+1)
	for crop, qty := range m.StockByCrop {
		if qty > 0 {
			stocks[model.CropByName(crop).Name] += qty
		}
	}
	if len(stocks) == 0 && m.Stock > 0 {
		stocks[model.DefaultPlantType] = m.Stock
	}
	m.StockByCrop = stocks
	m.Stock = totalStock(stocks)
	return m
}

func normalizeSeeds(m model.Model) model.Model {
	seeds := make(map[string]int, len(m.SeedsByCrop)+1)
	for crop, qty := range m.SeedsByCrop {
		if qty > 0 {
			seeds[model.CropByName(crop).Name] += qty
		}
	}
	if len(seeds) == 0 && m.Seeds > 0 {
		seeds[model.DefaultPlantType] = m.Seeds
	}
	m.SeedsByCrop = seeds
	m.Seeds = totalSeeds(seeds)
	return m
}

func selectedCropSeeds(m model.Model) int {
	crop := model.CropByName(m.SelectedCrop)
	return m.SeedsByCrop[crop.Name]
}

func totalSeeds(seeds map[string]int) int {
	total := 0
	for _, qty := range seeds {
		total += qty
	}
	return total
}

func totalStock(stocks map[string]int) int {
	total := 0
	for _, qty := range stocks {
		total += qty
	}
	return total
}

func stockValue(stocks map[string]int) float64 {
	var total float64
	for cropName, qty := range stocks {
		crop := model.CropByName(cropName)
		total += float64(qty) * crop.SellPrice
	}
	return total
}

func StockMarketValue(m model.Model) float64 {
	m = normalizeStock(m)
	return stockValue(m.StockByCrop)
}

func sellStock(m model.Model, label string) (model.Model, float64) {
	m = normalizeStock(m)
	if m.Stock == 0 {
		return m, 0
	}
	earned := stockValue(m.StockByCrop)
	stockCount := m.Stock
	m.Money += earned
	m.Stock = 0
	m.StockByCrop = map[string]int{}
	return addLog(m, fmt.Sprintf("💰 %s: +$%.0f (%d estoques)", label, earned, stockCount)), earned
}
