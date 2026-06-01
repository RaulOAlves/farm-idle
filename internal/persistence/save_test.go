// internal/persistence/save_test.go
package persistence_test

import (
	"farm-idle/internal/persistence"
	"path/filepath"
	"testing"
)

func TestSaveLoad_Roundtrip(t *testing.T) {
	m := persistence.DefaultModel()
	m.Money = 500
	m.Seeds = 3
	m.Day = 7
	m.HarvestLevel = 2

	path := filepath.Join(t.TempDir(), "save.json")

	if err := persistence.Save(m, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := persistence.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Money != m.Money {
		t.Errorf("money: want %.0f, got %.0f", m.Money, loaded.Money)
	}
	if loaded.Seeds != m.Seeds {
		t.Errorf("seeds: want %d, got %d", m.Seeds, loaded.Seeds)
	}
	if loaded.Day != m.Day {
		t.Errorf("day: want %d, got %d", m.Day, loaded.Day)
	}
	if loaded.HarvestLevel != m.HarvestLevel {
		t.Errorf("harvest_level: want %d, got %d", m.HarvestLevel, loaded.HarvestLevel)
	}
	if loaded.FieldSize != m.FieldSize {
		t.Errorf("field_size: want %d, got %d", m.FieldSize, loaded.FieldSize)
	}
}

func TestLoad_NonExistentReturnsDefault(t *testing.T) {
	m, err := persistence.Load("/nonexistent/path/save.json")
	if err != nil {
		t.Fatalf("Load of non-existent file should not error, got: %v", err)
	}
	def := persistence.DefaultModel()
	if m.Money != def.Money {
		t.Errorf("money: want %.0f, got %.0f", def.Money, m.Money)
	}
	if m.FieldSize != def.FieldSize {
		t.Errorf("field_size: want %d, got %d", def.FieldSize, m.FieldSize)
	}
	if m.HarvestLevel != def.HarvestLevel {
		t.Errorf("harvest_level: want %d, got %d", def.HarvestLevel, m.HarvestLevel)
	}
}

func TestDefaultModel_HasCorrectInitialValues(t *testing.T) {
	m := persistence.DefaultModel()
	if m.Money != 100 {
		t.Errorf("money: want 100, got %.0f", m.Money)
	}
	if m.Seeds != 10 {
		t.Errorf("seeds: want 10, got %d", m.Seeds)
	}
	if m.FieldSize != 10 {
		t.Errorf("field_size: want 10, got %d", m.FieldSize)
	}
	if len(m.Plants) != 10 {
		t.Errorf("len(plants): want 10, got %d", len(m.Plants))
	}
	if m.HarvestLevel != 1 {
		t.Errorf("harvest_level: want 1, got %d", m.HarvestLevel)
	}
}
