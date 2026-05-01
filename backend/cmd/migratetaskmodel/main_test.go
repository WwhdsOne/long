package main

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestParseArgsRequiresExplicitMongoConfig(t *testing.T) {
	_, err := parseArgs([]string{"migrate", "--db", "long"})
	if err == nil {
		t.Fatal("expected parseArgs to reject missing mongo uri")
	}
}

func TestNormalizeTaskDefinitionDocumentFromLegacyFields(t *testing.T) {
	doc := bson.M{
		"task_id":        "daily_click_100",
		"task_type":      "daily",
		"condition_kind": "daily_clicks",
	}

	update, changed := normalizeTaskDefinitionDocument(doc)
	if !changed {
		t.Fatal("expected legacy task definition to require migration")
	}
	if update["event_kind"] != "click" {
		t.Fatalf("expected event_kind click, got %v", update["event_kind"])
	}
	if update["window_kind"] != "daily" {
		t.Fatalf("expected window_kind daily, got %v", update["window_kind"])
	}
}

func TestNormalizeTaskArchiveDocumentSupportsCustomCollections(t *testing.T) {
	doc := bson.M{
		"task_id":        "campaign_click_100",
		"task_type":      "limited",
		"condition_kind": "daily_clicks",
		"start_at":       int64(10),
		"end_at":         int64(20),
	}

	update, changed := normalizeTaskArchiveDocument(doc)
	if !changed {
		t.Fatal("expected legacy archive to require migration")
	}
	if update["event_kind"] != "click" {
		t.Fatalf("expected event_kind click, got %v", update["event_kind"])
	}
	if update["window_kind"] != "fixed_range" {
		t.Fatalf("expected window_kind fixed_range, got %v", update["window_kind"])
	}
}
