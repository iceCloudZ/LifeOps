package main

import "testing"

func TestParseValidDraftJSON(t *testing.T) {
	result := ParseAIDrafts(`{
		"drafts": [
			{
				"draft_type": "task",
				"title": "准备彩笔",
				"description": "周五孩子需要带彩笔",
				"confidence": 0.91
			}
		]
	}`)

	if !result.OK {
		t.Fatalf("expected parse success, got failure %q", result.FailureReason)
	}
	if len(result.Drafts) != 1 {
		t.Fatalf("expected one draft, got %d", len(result.Drafts))
	}
	draft := result.Drafts[0]
	if draft.DraftType != "task" {
		t.Fatalf("expected draft type task, got %q", draft.DraftType)
	}
	if draft.Title != "准备彩笔" {
		t.Fatalf("expected title 准备彩笔, got %q", draft.Title)
	}
	if draft.Confidence != 0.91 {
		t.Fatalf("expected confidence 0.91, got %v", draft.Confidence)
	}
}

func TestRejectNonJSONModelOutput(t *testing.T) {
	result := ParseAIDrafts("我已经帮你整理好了：记得周五带彩笔。")

	if result.OK {
		t.Fatal("expected parse failure")
	}
	if result.FailureReason != "invalid_json" {
		t.Fatalf("expected invalid_json, got %q", result.FailureReason)
	}
}

func TestRejectEmptyDrafts(t *testing.T) {
	result := ParseAIDrafts(`{"drafts":[]}`)

	if result.OK {
		t.Fatal("expected parse failure")
	}
	if result.FailureReason != "missing_drafts" {
		t.Fatalf("expected missing_drafts, got %q", result.FailureReason)
	}
}

func TestRejectMissingDraftType(t *testing.T) {
	result := ParseAIDrafts(`{"drafts":[{"title":"test","confidence":0.9}]}`)

	if result.OK {
		t.Fatal("expected parse failure")
	}
	if result.FailureReason != "invalid_draft" {
		t.Fatalf("expected invalid_draft, got %q", result.FailureReason)
	}
}

func TestRejectMissingTitle(t *testing.T) {
	result := ParseAIDrafts(`{"drafts":[{"draft_type":"task","confidence":0.9}]}`)

	if result.OK {
		t.Fatal("expected parse failure")
	}
	if result.FailureReason != "invalid_draft" {
		t.Fatalf("expected invalid_draft, got %q", result.FailureReason)
	}
}

func TestRejectInvalidDraftType(t *testing.T) {
	result := ParseAIDrafts(`{"drafts":[{"draft_type":"invalid","title":"test","confidence":0.9}]}`)

	if result.OK {
		t.Fatal("expected parse failure")
	}
	if result.FailureReason != "invalid_draft_type" {
		t.Fatalf("expected invalid_draft_type, got %q", result.FailureReason)
	}
}

func TestRejectConfidenceOutOfRange(t *testing.T) {
	result := ParseAIDrafts(`{"drafts":[{"draft_type":"task","title":"test","confidence":1.5}]}`)

	if result.OK {
		t.Fatal("expected parse failure")
	}
	if result.FailureReason != "invalid_confidence" {
		t.Fatalf("expected invalid_confidence, got %q", result.FailureReason)
	}
}

func TestRejectNegativeConfidence(t *testing.T) {
	result := ParseAIDrafts(`{"drafts":[{"draft_type":"task","title":"test","confidence":-0.1}]}`)

	if result.OK {
		t.Fatal("expected parse failure")
	}
	if result.FailureReason != "invalid_confidence" {
		t.Fatalf("expected invalid_confidence, got %q", result.FailureReason)
	}
}

func TestAcceptAllValidDraftTypes(t *testing.T) {
	types := []string{"event", "task", "shopping_item", "note"}
	for _, dt := range types {
		result := ParseAIDrafts(`{"drafts":[{"draft_type":"` + dt + `","title":"test","confidence":0.5}]}`)
		if !result.OK {
			t.Fatalf("expected %q to be valid, got %s", dt, result.FailureReason)
		}
	}
}

func TestStripsMarkdownFence(t *testing.T) {
	result := ParseAIDrafts("```json\n{\"drafts\":[{\"draft_type\":\"task\",\"title\":\"test\",\"confidence\":0.9}]}\n```")

	if !result.OK {
		t.Fatalf("expected parse success, got %q", result.FailureReason)
	}
	if result.Drafts[0].Title != "test" {
		t.Fatalf("expected title test, got %q", result.Drafts[0].Title)
	}
}

func TestTrimsWhitespace(t *testing.T) {
	result := ParseAIDrafts(`  {"drafts":[{"draft_type":"task","title":"  hello  ","confidence":0.9}]}  `)

	if !result.OK {
		t.Fatalf("expected parse success, got %q", result.FailureReason)
	}
	if result.Drafts[0].Title != "hello" {
		t.Fatalf("expected trimmed title 'hello', got %q", result.Drafts[0].Title)
	}
}

func TestAcceptBoundaryConfidence(t *testing.T) {
	for _, c := range []float64{0.0, 1.0} {
		result := ParseAIDrafts(`{"drafts":[{"draft_type":"task","title":"test","confidence":` + formatFloat(c) + `}]}`)
		if !result.OK {
			t.Fatalf("expected confidence %v to be valid, got %s", c, result.FailureReason)
		}
	}
}

func formatFloat(f float64) string {
	if f == 0 {
		return "0"
	}
	return "1"
}

func TestParseMultipleDrafts(t *testing.T) {
	result := ParseAIDrafts(`{
		"drafts": [
			{"draft_type":"task","title":"a","confidence":0.8},
			{"draft_type":"event","title":"b","confidence":0.7},
			{"draft_type":"shopping_item","title":"c","confidence":0.9},
			{"draft_type":"note","title":"d","confidence":0.6}
		]
	}`)

	if !result.OK {
		t.Fatalf("expected success, got %q", result.FailureReason)
	}
	if len(result.Drafts) != 4 {
		t.Fatalf("expected 4 drafts, got %d", len(result.Drafts))
	}
}

func TestOptionalFieldsPreserved(t *testing.T) {
	due := "2026-06-10"
	result := ParseAIDrafts(`{
		"drafts": [{
			"draft_type": "task",
			"title": "test",
			"confidence": 0.8,
			"due_at": "2026-06-10",
			"assignee_hint": "partner",
			"description": "some desc"
		}]
	}`)

	if !result.OK {
		t.Fatalf("expected success, got %q", result.FailureReason)
	}
	d := result.Drafts[0]
	if d.DueAt == nil || *d.DueAt != due {
		t.Fatalf("expected due_at %q, got %v", due, d.DueAt)
	}
	if d.AssigneeHint == nil || *d.AssigneeHint != "partner" {
		t.Fatalf("expected assignee_hint partner, got %v", d.AssigneeHint)
	}
	if d.Description != "some desc" {
		t.Fatalf("expected description 'some desc', got %q", d.Description)
	}
}
