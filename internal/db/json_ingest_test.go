package db

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestFlattenJSONExtractsArraysAndFlattensScalars(t *testing.T) {
	payload := map[string]any{
		"id": "CIF-1",
		"profile": map[string]any{
			"status": "active",
			"ok":     true,
		},
		"items": []any{
			map[string]any{"code": "A"},
			"raw",
		},
		"notes":     []any{},
		"Status OK": "yes",
		"Status-OK": "no",
	}

	flat, arrays := flattenJSON(payload)

	if got := flat["id"]; got != "CIF-1" {
		t.Fatalf("id mismatch: got=%v", got)
	}
	if got := flat["profile__status"]; got != "active" {
		t.Fatalf("profile__status mismatch: got=%v", got)
	}
	if got := flat["profile__ok"]; got != "true" {
		t.Fatalf("profile__ok mismatch: got=%v", got)
	}
	v1, ok1 := flat["status_o_k"]
	v2, ok2 := flat["status_o_k_2"]
	if !ok1 || !ok2 {
		t.Fatalf("expected status_o_k and status_o_k_2 keys, got map=%v", flat)
	}
	gotSet := map[any]struct{}{v1: {}, v2: {}}
	if _, ok := gotSet["yes"]; !ok {
		t.Fatalf("expected yes in status collision set, got=%v", gotSet)
	}
	if _, ok := gotSet["no"]; !ok {
		t.Fatalf("expected no in status collision set, got=%v", gotSet)
	}
	if _, ok := flat["items"]; ok {
		t.Fatalf("array key items should not be present in parent flat map")
	}

	if got := len(arrays["items"]); got != 2 {
		t.Fatalf("items array length mismatch: got=%d", got)
	}
	if got := len(arrays["notes"]); got != 0 {
		t.Fatalf("notes array length mismatch: got=%d", got)
	}
}

func TestBuildChildDataHandlesObjectsAndPrimitives(t *testing.T) {
	arrays := map[string][]any{
		"items": {
			map[string]any{
				"code":   "A",
				"nested": []any{1, 2},
			},
			"raw",
			5.0,
		},
	}

	children := buildChildData("raw_cif", arrays)
	if len(children) != 1 {
		t.Fatalf("children length mismatch: got=%d", len(children))
	}

	child := children[0]
	if child.table != "raw_cif__items" {
		t.Fatalf("child table mismatch: got=%s", child.table)
	}
	if !reflect.DeepEqual(child.columns, []string{"code", "nested", "value"}) {
		t.Fatalf("child columns mismatch: got=%v", child.columns)
	}
	if len(child.rows) != 3 {
		t.Fatalf("child rows mismatch: got=%d", len(child.rows))
	}

	if got := child.rows[0].values["code"]; got != "A" {
		t.Fatalf("row0 code mismatch: got=%v", got)
	}
	if got := child.rows[0].values["nested"]; got != "[1,2]" {
		t.Fatalf("row0 nested mismatch: got=%v", got)
	}
	if got := child.rows[1].values["value"]; got != "raw" {
		t.Fatalf("row1 value mismatch: got=%v", got)
	}
	if got := child.rows[2].values["value"]; got != "5" {
		t.Fatalf("row2 value mismatch: got=%v", got)
	}
}

func TestExtractRowKeyPriority(t *testing.T) {
	flat := map[string]any{
		"id":         "ID-1",
		"norekening": "REK-1",
		"nocif":      "CIF-1",
		"custom":     "CUSTOM-1",
	}

	if got := extractRowKey(flat, []string{"custom"}); got != "CUSTOM-1" {
		t.Fatalf("key hint mismatch: got=%s", got)
	}

	if got := extractRowKey(flat, nil); got != "ID-1" {
		t.Fatalf("default key mismatch: got=%s", got)
	}
}

func TestBuildOverflowJSONWithSchemaFullStyleInputs(t *testing.T) {
	columns := []string{"a", "b"}
	values := map[string]any{
		"a": "x",
		"b": "y",
	}
	existing := map[string]struct{}{
		"a":              {},
		"_overflow_json": {},
	}

	persisted := selectPersistedColumns(columns, existing)
	if !reflect.DeepEqual(persisted, []string{"a"}) {
		t.Fatalf("persisted columns mismatch: got=%v", persisted)
	}

	overflowJSON, err := buildOverflowJSON(columns, values, existing, true)
	if err != nil {
		t.Fatalf("build overflow failed: %v", err)
	}
	if overflowJSON == nil {
		t.Fatalf("overflow JSON should not be nil")
	}

	var overflow map[string]any
	if err := json.Unmarshal([]byte(overflowJSON.(string)), &overflow); err != nil {
		t.Fatalf("unmarshal overflow failed: %v", err)
	}

	if !reflect.DeepEqual(overflow, map[string]any{"b": "y"}) {
		t.Fatalf("overflow mismatch: got=%v", overflow)
	}
}

func TestIsConcurrentDDLError(t *testing.T) {
	if !isConcurrentDDLError(&mysql.MySQLError{Number: 1684}) {
		t.Fatalf("expected true for mysql 1684")
	}
	if isConcurrentDDLError(&mysql.MySQLError{Number: 1213}) {
		t.Fatalf("expected false for non-1684 code")
	}
}
