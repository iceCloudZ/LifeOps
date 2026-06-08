package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LensMeta struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	SkillType       string   `json:"skill_type"`
	Domain          string   `json:"domain"`
	Path            string   `json:"path"`
	Description     string   `json:"description"`
	BestFor         []string `json:"best_for"`
	AvoidIf         []string `json:"avoid_if"`
	RequiredContext []string `json:"required_context"`
	Style           []string `json:"style"`
	RiskLevel       string   `json:"risk_level"`
}

type Registry struct {
	SchemaVersion string     `json:"schema_version"`
	Skills        []LensMeta `json:"skills"`
}

type SkillRegistry struct {
	registry   Registry
	lensContent map[string]string
}

func NewSkillRegistry(skillsPath string) (*SkillRegistry, error) {
	if skillsPath == "" {
		return &SkillRegistry{}, nil
	}

	regPath := filepath.Join(skillsPath, "registry.json")
	data, err := os.ReadFile(regPath)
	if err != nil {
		return nil, fmt.Errorf("read registry: %w", err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}

	content := make(map[string]string, len(reg.Skills))
	for i := range reg.Skills {
		skill := &reg.Skills[i]
		if skill.SkillType != "lens" {
			continue
		}
		lensPath := filepath.Join(skillsPath, filepath.FromSlash(skill.Path))
		md, err := os.ReadFile(lensPath)
		if err != nil {
			continue
		}
		content[skill.ID] = string(md)
	}

	return &SkillRegistry{registry: reg, lensContent: content}, nil
}

func (sr *SkillRegistry) Available() bool {
	return len(sr.registry.Skills) > 0
}

func (sr *SkillRegistry) EntrySkill() *LensMeta {
	for i := range sr.registry.Skills {
		if sr.registry.Skills[i].SkillType == "entry" {
			return &sr.registry.Skills[i]
		}
	}
	return nil
}

func (sr *SkillRegistry) ListLenses(domain string) []LensMeta {
	var result []LensMeta
	for _, s := range sr.registry.Skills {
		if s.SkillType != "lens" {
			continue
		}
		if domain != "" && s.Domain != domain {
			continue
		}
		result = append(result, s)
	}
	return result
}

func (sr *SkillRegistry) GetLens(id string) *LensMeta {
	for i := range sr.registry.Skills {
		if sr.registry.Skills[i].ID == id {
			return &sr.registry.Skills[i]
		}
	}
	return nil
}

func (sr *SkillRegistry) GetLensContent(id string) string {
	return sr.lensContent[id]
}

func (sr *SkillRegistry) LensIndexText(domain string) string {
	lenses := sr.ListLenses(domain)
	if len(lenses) == 0 {
		return ""
	}
	var b strings.Builder
	for _, l := range lenses {
		fmt.Fprintf(&b, "- %s: %s\n", l.ID, l.Description)
		if len(l.BestFor) > 0 {
			fmt.Fprintf(&b, "  best_for: %s\n", strings.Join(l.BestFor, ", "))
		}
		if len(l.AvoidIf) > 0 {
			fmt.Fprintf(&b, "  avoid_if: %s\n", strings.Join(l.AvoidIf, ", "))
		}
	}
	return b.String()
}

func (sr *SkillRegistry) Domains() []string {
	seen := make(map[string]bool)
	var domains []string
	for _, s := range sr.registry.Skills {
		if s.SkillType == "lens" && !seen[s.Domain] {
			seen[s.Domain] = true
			domains = append(domains, s.Domain)
		}
	}
	return domains
}
