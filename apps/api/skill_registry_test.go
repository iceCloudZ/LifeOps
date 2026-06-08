package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewSkillRegistryEmpty(t *testing.T) {
	reg, err := NewSkillRegistry("")
	if err != nil {
		t.Fatalf("empty path should not error: %v", err)
	}
	if reg.Available() {
		t.Fatal("empty path should not be available")
	}
	if lenses := reg.ListLenses("finance"); len(lenses) != 0 {
		t.Fatalf("expected 0 lenses, got %d", len(lenses))
	}
}

func TestNewSkillRegistryFromPath(t *testing.T) {
	dir := t.TempDir()

	regData := map[string]interface{}{
		"schema_version": "0.2.0",
		"skills": []map[string]interface{}{
			{
				"id":               "test-lens",
				"name":             "Test Lens",
				"skill_type":       "lens",
				"domain":           "finance",
				"path":             "lenses/finance/test-lens.md",
				"description":      "A test lens",
				"best_for":         []string{"testing"},
				"avoid_if":         []string{"production"},
				"required_context": []string{},
				"style":            []string{"test"},
				"risk_level":       "low",
			},
		},
	}
	regJSON, _ := json.Marshal(regData)
	os.WriteFile(filepath.Join(dir, "registry.json"), regJSON, 0644)

	lensDir := filepath.Join(dir, "lenses", "finance")
	os.MkdirAll(lensDir, 0755)
	os.WriteFile(filepath.Join(lensDir, "test-lens.md"), []byte("# Test Lens\n\nContent here."), 0644)

	reg, err := NewSkillRegistry(dir)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	if !reg.Available() {
		t.Fatal("should be available")
	}
	if len(reg.ListLenses("")) != 1 {
		t.Fatalf("expected 1 lens, got %d", len(reg.ListLenses("")))
	}
	if len(reg.ListLenses("finance")) != 1 {
		t.Fatalf("expected 1 finance lens, got %d", len(reg.ListLenses("finance")))
	}
	if len(reg.ListLenses("health")) != 0 {
		t.Fatalf("expected 0 health lenses, got %d", len(reg.ListLenses("health")))
	}

	lens := reg.GetLens("test-lens")
	if lens == nil {
		t.Fatal("expected to find test-lens")
	}
	if lens.Domain != "finance" {
		t.Fatalf("expected finance domain, got %s", lens.Domain)
	}

	content := reg.GetLensContent("test-lens")
	if content == "" {
		t.Fatal("expected lens content")
	}

	domains := reg.Domains()
	if len(domains) != 1 || domains[0] != "finance" {
		t.Fatalf("expected [finance], got %v", domains)
	}
}

func TestSkillRegistryLensIndexText(t *testing.T) {
	reg, _ := NewSkillRegistry("")
	if text := reg.LensIndexText("finance"); text != "" {
		t.Fatalf("expected empty text for empty registry, got %q", text)
	}
}
