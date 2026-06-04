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
