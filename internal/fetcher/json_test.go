package fetcher

import "testing"

func TestWrapSavingStatementsPayloadAddsNoRekeningAndPreservesRootFields(t *testing.T) {
	payload := map[string]any{
		"saldoawal": "1000",
		"mutasi": []any{
			map[string]any{"keterangan": "setor"},
		},
	}

	got, ok := wrapSavingStatementsPayload("REK-001", payload).(map[string]any)
	if !ok {
		t.Fatalf("wrapped payload type mismatch: got %T", got)
	}

	if got["norekening"] != "REK-001" {
		t.Fatalf("norekening mismatch: got=%v", got["norekening"])
	}
	if got["saldoawal"] != "1000" {
		t.Fatalf("saldoawal mismatch: got=%v", got["saldoawal"])
	}

	mutasi, ok := got["mutasi"].([]any)
	if !ok {
		t.Fatalf("mutasi type mismatch: got %T", got["mutasi"])
	}
	if len(mutasi) != 1 {
		t.Fatalf("mutasi length mismatch: got=%d", len(mutasi))
	}

	if _, exists := payload["norekening"]; exists {
		t.Fatal("original payload should not be mutated")
	}
}

func TestWrapSavingStatementsPayloadWrapsNonObjectValue(t *testing.T) {
	got, ok := wrapSavingStatementsPayload("REK-001", []any{"raw"}).(map[string]any)
	if !ok {
		t.Fatalf("wrapped payload type mismatch: got %T", got)
	}

	if got["norekening"] != "REK-001" {
		t.Fatalf("norekening mismatch: got=%v", got["norekening"])
	}

	value, ok := got["value"].([]any)
	if !ok {
		t.Fatalf("value type mismatch: got %T", got["value"])
	}
	if len(value) != 1 || value[0] != "raw" {
		t.Fatalf("value mismatch: got=%v", value)
	}
}
