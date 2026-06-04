package main

import "encoding/json"

type AIDraft struct {
	DraftType   string  `json:"draft_type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

type AIDraftParseResult struct {
	OK            bool
	Drafts        []AIDraft
	FailureReason string
}

type aiDraftEnvelope struct {
	Drafts []AIDraft `json:"drafts"`
}

func ParseAIDrafts(modelOutput string) AIDraftParseResult {
	var envelope aiDraftEnvelope
	if err := json.Unmarshal([]byte(modelOutput), &envelope); err != nil {
		return AIDraftParseResult{OK: false, FailureReason: "invalid_json"}
	}
	if len(envelope.Drafts) == 0 {
		return AIDraftParseResult{OK: false, FailureReason: "missing_drafts"}
	}
	for _, draft := range envelope.Drafts {
		if draft.DraftType == "" || draft.Title == "" {
			return AIDraftParseResult{OK: false, FailureReason: "invalid_draft"}
		}
	}
	return AIDraftParseResult{OK: true, Drafts: envelope.Drafts}
}
