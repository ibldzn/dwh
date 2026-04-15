package main

import (
	"reflect"
	"testing"
	"time"
)

func TestDefaultDateWindow(t *testing.T) {
	now := time.Date(2026, 2, 26, 14, 30, 15, 0, time.FixedZone("UTC+7", 7*60*60))

	t.Setenv(envIngestStartDate, "")
	t.Setenv(envIngestEndDate, "")

	w, err := loadDateWindow(now)
	if err != nil {
		t.Fatalf("loadDateWindow failed: %v", err)
	}

	wantStart := now.AddDate(0, 0, -7)
	wantEnd := now.AddDate(0, 0, -1)

	if !w.start.Equal(wantStart) {
		t.Fatalf("unexpected start: got %v, want %v", w.start, wantStart)
	}
	if !w.end.Equal(wantEnd) {
		t.Fatalf("unexpected end: got %v, want %v", w.end, wantEnd)
	}

	wantAsOf := now.Format("2006-01-02")
	if w.asOf != wantAsOf {
		t.Fatalf("unexpected asOf: got %q, want %q", w.asOf, wantAsOf)
	}

	if got := daysInWindow(w); got != 7 {
		t.Fatalf("unexpected window days: got %d, want 7", got)
	}
}

func TestLoadDateWindow(t *testing.T) {
	loc := time.FixedZone("UTC+7", 7*60*60)
	now := time.Date(2026, 2, 26, 14, 30, 15, 0, loc)

	testCases := []struct {
		name       string
		startDate  string
		endDate    string
		wantWindow *dateWindow
		wantErr    string
	}{
		{
			name: "no overrides uses default server local window",
			wantWindow: &dateWindow{
				start: now.AddDate(0, 0, -7),
				end:   now.AddDate(0, 0, -1),
				asOf:  now.Format("2006-01-02"),
			},
		},
		{
			name:      "valid overrides use explicit range",
			startDate: "2026-02-01",
			endDate:   "2026-02-05",
			wantWindow: &dateWindow{
				start: time.Date(2026, 2, 1, 0, 0, 0, 0, loc),
				end:   time.Date(2026, 2, 5, 0, 0, 0, 0, loc),
				asOf:  now.Format("2006-01-02"),
			},
		},
		{
			name:      "only start date set returns error",
			startDate: "2026-02-01",
			wantErr:   "INGEST_START_DATE and INGEST_END_DATE must both be set together",
		},
		{
			name:    "only end date set returns error",
			endDate: "2026-02-05",
			wantErr: "INGEST_START_DATE and INGEST_END_DATE must both be set together",
		},
		{
			name:      "invalid start date format returns error",
			startDate: "2026/02/01",
			endDate:   "2026-02-05",
			wantErr:   `invalid INGEST_START_DATE "2026/02/01": expected YYYY-MM-DD`,
		},
		{
			name:      "invalid end date format returns error",
			startDate: "2026-02-01",
			endDate:   "2026/02/05",
			wantErr:   `invalid INGEST_END_DATE "2026/02/05": expected YYYY-MM-DD`,
		},
		{
			name:      "end before start returns error",
			startDate: "2026-02-05",
			endDate:   "2026-02-01",
			wantErr:   "INGEST_END_DATE must be on or after INGEST_START_DATE",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(envIngestStartDate, tc.startDate)
			t.Setenv(envIngestEndDate, tc.endDate)

			got, err := loadDateWindow(now)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.wantErr)
				}
				if err.Error() != tc.wantErr {
					t.Fatalf("unexpected error: got %q, want %q", err.Error(), tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("loadDateWindow failed: %v", err)
			}
			if !got.start.Equal(tc.wantWindow.start) {
				t.Fatalf("unexpected start: got %v, want %v", got.start, tc.wantWindow.start)
			}
			if !got.end.Equal(tc.wantWindow.end) {
				t.Fatalf("unexpected end: got %v, want %v", got.end, tc.wantWindow.end)
			}
			if got.asOf != tc.wantWindow.asOf {
				t.Fatalf("unexpected asOf: got %q, want %q", got.asOf, tc.wantWindow.asOf)
			}
		})
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
