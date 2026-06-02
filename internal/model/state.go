// internal/model/state.go
package model

import "time"

// PlantState documents slot lifecycle stages.
type PlantState string
type FocusMode string

// Plant states
const (
	PlantEmpty   PlantState = "empty"
	PlantPlanted PlantState = "planted"
	PlantGrowing PlantState = "growing"
	PlantReady   PlantState = "ready"
)

const (
	FocusMenu  FocusMode = "menu"
	FocusField FocusMode = "field"
)

// Game constants
const (
	GrowTicks                             = 30
	SeedCost                      float64 = 10
	FieldExpandBaseCost           float64 = 100
	HarvestUpgradeBaseCost        float64 = 250
	StockValue                    float64 = 5
	DefaultAutoSellThreshold              = 5
	DefaultAutoBuyMaxCashFraction         = 0.30
	DefaultPlantType                      = "Trigo"
	MaxLogEntries                         = 20
	TicksPerDay                           = 60
)

const AutoSaveInterval = 60 * time.Second

type PlantSlot struct {
	State          PlantState `json:"state"`
	TicksRemaining int        `json:"ticks_remaining"`
	PlantType      string     `json:"plant_type,omitempty"`
}

type LogEntry struct {
	Timestamp time.Time
	Message   string
}

type OfflineResult struct {
	Duration   time.Duration
	Harvests   int
	Revenue    float64
	Efficiency float64 // harvests_reais / harvests_teoricos_maximos
}

type RevenueTracker struct {
	Buckets [TicksPerDay]float64
	Cursor  int
	Total   float64
}

type Model struct {
	// Resources
	Money float64
	Seeds int
	Stock int

	// Field
	FieldSize int
	Plants    []PlantSlot

	// Economy config
	SeedsPerPurchase       int
	AutoBuyEnabled         bool
	AutoBuyMinimum         int
	AutoBuyMaxCashFraction float64

	// Upgrades
	HarvestLevel      int
	AutoSellThreshold int

	// Meta
	Day            int
	TickCount      int
	LastSave       time.Time
	Log            []LogEntry
	MoneySnapshot  float64 // money há TicksPerDay ticks atrás, para lucro/min
	RevenueTracker RevenueTracker
	RecentRevenue  float64 // receita operacional recente; não é lucro líquido

	// UI state — não persistido
	Cursor        int
	FieldCursor   int
	FocusMode     FocusMode
	InputMode     bool
	InputBuffer   string
	OfflineReport *OfflineResult
	ViewWidth     int
	ViewHeight    int
}
