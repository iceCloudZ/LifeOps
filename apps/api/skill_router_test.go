package main

import (
	"strings"
	"testing"
)

func TestSkillRouterFallbackPrompts(t *testing.T) {
	router := NewSkillRouter(nil, nil)

	for _, domain := range []string{"finance", "health", "movement", "family"} {
		prompt := router.fallbackPrompt(domain)
		if prompt == "" {
			t.Fatalf("expected fallback prompt for %s", domain)
		}
		if !strings.Contains(prompt, "中文") {
			t.Fatalf("expected Chinese in fallback prompt for %s", domain)
		}
	}

	unknown := router.fallbackPrompt("unknown")
	if unknown == "" {
		t.Fatal("expected fallback for unknown domain")
	}
}

func TestSkillRouterBuildSystemPromptWithoutLens(t *testing.T) {
	router := NewSkillRouter(nil, nil)
	prompt := router.BuildSystemPrompt("finance", nil)
	if !strings.Contains(prompt, "财务") {
		t.Fatalf("expected fallback finance prompt, got: %s", prompt)
	}
}

func TestSkillRouterBuildSystemPromptWithMissingLens(t *testing.T) {
	reg, _ := NewSkillRegistry("")
	router := NewSkillRouter(reg, nil)
	result := &RoutingResult{PrimaryLensID: "nonexistent"}
	prompt := router.BuildSystemPrompt("finance", result)
	if !strings.Contains(prompt, "财务") {
		t.Fatalf("expected fallback when lens not found, got: %s", prompt)
	}
}

func TestFirstNLines(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5"
	result := firstNLines(input, 3)
	if !strings.HasPrefix(result, "line1\nline2\nline3") {
		t.Fatalf("expected first 3 lines, got: %s", result)
	}
	if !strings.HasSuffix(result, "...") {
		t.Fatal("expected truncation marker")
	}

	short := "line1\nline2"
	result2 := firstNLines(short, 5)
	if result2 != short {
		t.Fatalf("expected no truncation, got: %s", result2)
	}
}
