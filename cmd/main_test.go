package main

import (
	"reflect"
	"testing"
	"time"
)

func TestDefaultDateWindow(t *testing.T) {
	now := time.Date(2026, 2, 26, 14, 30, 15, 0, time.FixedZone("UTC+7", 7*60*60))

	w := defaultDateWindow(now)
	wantStart := now.UTC().AddDate(0, 0, -3)
	wantEnd := now.UTC().AddDate(0, 0, -1)

	if !w.start.Equal(wantStart) {
		t.Fatalf("unexpected start: got %v, want %v", w.start, wantStart)
	}
	if !w.end.Equal(wantEnd) {
		t.Fatalf("unexpected end: got %v, want %v", w.end, wantEnd)
	}

	wantAsOf := now.UTC().Format("2006-01-02")
	if w.asOf != wantAsOf {
		t.Fatalf("unexpected asOf: got %q, want %q", w.asOf, wantAsOf)
	}

	if got := daysInWindow(w); got != 3 {
		t.Fatalf("unexpected window days: got %d, want 3", got)
	}
}

func TestLoadRuntimeConfig(t *testing.T) {
	t.Run("parses flags and normalizes concurrency minimum", func(t *testing.T) {
		setBaseRuntimeEnv(t)
		t.Setenv("JSON_INGEST", "true")
		t.Setenv("INGEST_EOD", "1")
		t.Setenv("FETCH_JOURNAL_TRX", "TRUE")
		t.Setenv("FETCH_BALANCE_SHEET", "true")
		t.Setenv("FETCH_VAULT_MUTATIONS", "1")
		t.Setenv("FETCH_TELLER_MUTATIONS", "TRUE")
		t.Setenv("FETCH_COA_MOVEMENTS", "t")
		t.Setenv("FETCH_MASTER_DATA", "invalid-bool")
		t.Setenv("FETCH_CIF_ALL", "true")
		t.Setenv("FETCH_LOAN_ALL", "0")
		t.Setenv("FETCH_SAVINGS_ALL", "TRUE")
		t.Setenv("FETCH_TIME_DEPOSIT_ALL", "1")
		t.Setenv("INGEST_CONCURRENCY", "0")

		got := loadRuntimeConfig()

		if !got.jsonIngest {
			t.Fatalf("jsonIngest should be true")
		}
		if !got.ingestEOD || !got.fetchJournalTrx || !got.fetchBalanceSheet || !got.fetchVaultMut || !got.fetchTellerMut || !got.fetchCOAMovements {
			t.Fatalf("expected ingestEOD, fetchJournalTrx, fetchBalanceSheet, fetchVaultMut, fetchTellerMut, fetchCOAMovements to be true: %+v", got)
		}
		if got.fetchMasterData {
			t.Fatalf("fetchMasterData should be false for invalid bool value")
		}
		if !got.fetchCIFAll || got.fetchLoanAll != false || !got.fetchSavingsAll || !got.fetchTimeDepositAll {
			t.Fatalf("unexpected feature flags: %+v", got)
		}
		if got.ingestConcurrency != 1 {
			t.Fatalf("unexpected concurrency: got %d, want 1", got.ingestConcurrency)
		}
	})

	t.Run("uses fallback concurrency when env is invalid integer", func(t *testing.T) {
		setBaseRuntimeEnv(t)
		t.Setenv("INGEST_CONCURRENCY", "invalid")

		got := loadRuntimeConfig()
		if got.ingestConcurrency != 10 {
			t.Fatalf("unexpected concurrency fallback: got %d, want 10", got.ingestConcurrency)
		}
	})
}

func TestEnabledJobNamesOrder(t *testing.T) {
	cfg := runtimeConfig{
		ingestEOD:           true,
		fetchJournalTrx:     true,
		fetchBalanceSheet:   true,
		fetchVaultMut:       true,
		fetchTellerMut:      true,
		fetchCIFAll:         true,
		fetchLoanAll:        true,
		fetchTimeDepositAll: true,
	}

	got := enabledJobNames(cfg)
	want := []string{
		jobIngestEOD,
		jobFetchJournalTrx,
		jobFetchBalanceSheet,
		jobFetchVaultMut,
		jobFetchTellerMut,
		jobFetchCIFAll,
		jobFetchLoanAll,
		jobFetchTimeDeposit,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected job order: got %v, want %v", got, want)
	}
}

func setBaseRuntimeEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"JSON_INGEST",
		jobIngestEOD,
		jobFetchJournalTrx,
		jobFetchBalanceSheet,
		jobFetchVaultMut,
		jobFetchTellerMut,
		jobFetchCOAMovements,
		jobFetchMasterData,
		jobFetchCIFAll,
		jobFetchLoanAll,
		jobFetchSavingsAll,
		jobFetchTimeDeposit,
	}

	for _, key := range keys {
		t.Setenv(key, "false")
	}
	t.Setenv("INGEST_CONCURRENCY", "10")
}
