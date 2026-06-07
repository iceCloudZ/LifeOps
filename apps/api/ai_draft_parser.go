package main

import (
	"encoding/json"
	"regexp"
	"strings"
)

var allowedDraftTypes = map[string]bool{
	"event":         true,
	"task":          true,
	"shopping_item": true,
	"note":          true,
}

var markdownFenceRe = regexp.MustCompile("(?s)^```(?:json)?\\s*\n?(.*?)\n?```$")

type AIDraft struct {
	DraftType    string  `json:"draft_type"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Confidence   float64 `json:"confidence"`
	DueAt        *string `json:"due_at,omitempty"`
	AssigneeHint *string `json:"assignee_hint,omitempty"`
	Quantity     *string `json:"quantity,omitempty"`
	Topic        *string `json:"topic,omitempty"`
}

type AIDraftParseResult struct {
	OK            bool
	Drafts        []AIDraft
	FailureReason string
}

type aiDraftEnvelope struct {
	Drafts []AIDraft `json:"drafts"`
}

func sanitizeModelOutput(raw string) string {
	s := strings.TrimSpace(raw)
	if m := markdownFenceRe.FindStringSubmatch(s); len(m) == 2 {
		s = strings.TrimSpace(m[1])
	}
	return s
}

func ParseAIDrafts(modelOutput string) AIDraftParseResult {
	cleaned := sanitizeModelOutput(modelOutput)

	var envelope aiDraftEnvelope
	if err := json.Unmarshal([]byte(cleaned), &envelope); err != nil {
		return AIDraftParseResult{OK: false, FailureReason: "invalid_json"}
	}
	if len(envelope.Drafts) == 0 {
		return AIDraftParseResult{OK: false, FailureReason: "missing_drafts"}
	}

	for i := range envelope.Drafts {
		d := &envelope.Drafts[i]
		d.Title = strings.TrimSpace(d.Title)
		d.Description = strings.TrimSpace(d.Description)
		d.DraftType = strings.TrimSpace(d.DraftType)

		if d.DraftType == "" || d.Title == "" {
			return AIDraftParseResult{OK: false, FailureReason: "invalid_draft"}
		}
		if !allowedDraftTypes[d.DraftType] {
			return AIDraftParseResult{OK: false, FailureReason: "invalid_draft_type"}
		}
		if d.Confidence < 0 || d.Confidence > 1 {
			return AIDraftParseResult{OK: false, FailureReason: "invalid_confidence"}
		}
	}

	return AIDraftParseResult{OK: true, Drafts: envelope.Drafts}
}
