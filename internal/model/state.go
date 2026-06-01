// internal/model/state.go
package model

import "time"

// PlantState documents slot lifecycle stages.
type PlantState string

// Plant states
const (
	PlantEmpty   PlantState = "empty"
	PlantPlanted PlantState = "planted"
	PlantGrowing PlantState = "growing"
	PlantReady   PlantState = "ready"
)

// Game constants
const (
	GrowTicks                             = 30
	SeedCost                      float64 = 10
	FieldExpandBaseCost           float64 = 100
	HarvestUpgradeCost            float64 = 250
	StockValue                    float64 = 5
	DefaultAutoSellThreshold              = 5
	DefaultAutoBuyMaxCashFraction         = 0.30
	MaxLogEntries                         = 20
	TicksPerDay                           = 60
)

const AutoSaveInterval = 60 * time.Second

type PlantSlot struct {
	State          PlantState `json:"state"`
	TicksRemaining int        `json:"ticks_remaining"`
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
	InputMode     bool
	InputBuffer   string
	OfflineReport *OfflineResult
}

type SaveData struct {
	Money             float64     `json:"money"`
	Seeds             int         `json:"seeds"`
	Stock             int         `json:"stock"`
	FieldSize         int         `json:"field_size"`
	Plants            []PlantSlot `json:"plants"`
	HarvestLevel      int         `json:"harvest_level"`
	AutoSellThreshold int         `json:"auto_sell_threshold"`
	Day               int         `json:"day"`
	TickCount         int         `json:"tick_count"`
	LastSave          time.Time   `json:"last_save"`
	MoneySnapshot     float64     `json:"money_snapshot"`
}
