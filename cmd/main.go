package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ibldzn/dwh-v2/internal/db"
	"github.com/ibldzn/dwh-v2/internal/fetcher"
	"github.com/ibldzn/dwh-v2/internal/fincloud"
	"github.com/joho/godotenv"
	"github.com/schollz/progressbar/v3"
)

const (
	jobIngestEOD         = "INGEST_EOD"
	jobFetchJournalTrx   = "FETCH_JOURNAL_TRX"
	jobFetchBalanceSheet = "FETCH_BALANCE_SHEET"
	jobFetchVaultMut     = "FETCH_VAULT_MUTATIONS"
	jobFetchTellerMut    = "FETCH_TELLER_MUTATIONS"
	jobFetchCOAMovements = "FETCH_COA_MOVEMENTS"
	jobFetchMasterData   = "FETCH_MASTER_DATA"
	jobFetchCIFAll       = "FETCH_CIF_ALL"
	jobFetchLoanAll      = "FETCH_LOAN_ALL"
	jobFetchSavingsAll   = "FETCH_SAVINGS_ALL"
	jobFetchTimeDeposit  = "FETCH_TIME_DEPOSIT_ALL"

	envIngestStartDate = "INGEST_START_DATE"
	envIngestEndDate   = "INGEST_END_DATE"
)

type runtimeConfig struct {
	jsonIngest          bool
	ingestEOD           bool
	fetchJournalTrx     bool
	fetchBalanceSheet   bool
	fetchVaultMut       bool
	fetchTellerMut      bool
	fetchCOAMovements   bool
	fetchMasterData     bool
	fetchCIFAll         bool
	fetchLoanAll        bool
	fetchSavingsAll     bool
	fetchTimeDepositAll bool
	ingestConcurrency   int
}

type runtimeDeps struct {
	client *fincloud.Client
	fetch  *fetcher.Fetcher
	store  *db.Store
}

type dateWindow struct {
	start time.Time
	end   time.Time
	asOf  string
}

type eodData struct {
	file string
	date string
	data string
}

func main() {
	if err := run(); err != nil {
		errorExit("%v", err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	cfg := loadRuntimeConfig()
	window, err := loadDateWindow(time.Now())
	if err != nil {
		return fmt.Errorf("failed to load date window: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	client, err := fincloud.NewClient(fincloud.Config{})
	if err != nil {
		return fmt.Errorf("failed to create fincloud client: %w", err)
	}

	username, err := requireEnv("FINCLOUD_USERNAME")
	if err != nil {
		return err
	}

	password, err := requireEnv("FINCLOUD_PASSWORD")
	if err != nil {
		return err
	}

	roleID, err := requireEnv("FINCLOUD_ROLE_ID")
	if err != nil {
		return err
	}

	locationID, err := requireEnv("FINCLOUD_LOCATION_ID")
	if err != nil {
		return err
	}

	session, err := client.Login(ctx, fincloud.Credentials{
		Username:   username,
		Password:   password,
		RoleID:     roleID,
		LocationID: locationID,
	})
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	ctx = fincloud.WithFincloudSessionID(ctx, session.ID)

	fetch, err := fetcher.NewFetcher(client)
	if err != nil {
		return fmt.Errorf("failed to create fetcher: %w", err)
	}

	dsn, err := requireEnv("MYSQL_DSN")
	if err != nil {
		return err
	}

	storeDB, err := db.Open(dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer storeDB.Close()

	if err := db.Migrate(ctx, storeDB); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	store, err := db.NewStore(storeDB)
	if err != nil {
		return fmt.Errorf("failed to create db store: %w", err)
	}

	deps := runtimeDeps{
		client: client,
		fetch:  fetch,
		store:  store,
	}

	if err := runEnabledJobs(ctx, deps, cfg, window); err != nil {
		return err
	}

	return nil
}

func loadRuntimeConfig() runtimeConfig {
	concurrency := max(envInt("INGEST_CONCURRENCY", 10), 1)

	return runtimeConfig{
		jsonIngest:          envBool("JSON_INGEST"),
		ingestEOD:           envBool(jobIngestEOD),
		fetchJournalTrx:     envBool(jobFetchJournalTrx),
		fetchBalanceSheet:   envBool(jobFetchBalanceSheet),
		fetchVaultMut:       envBool(jobFetchVaultMut),
		fetchTellerMut:      envBool(jobFetchTellerMut),
		fetchCOAMovements:   envBool(jobFetchCOAMovements),
		fetchMasterData:     envBool(jobFetchMasterData),
		fetchCIFAll:         envBool(jobFetchCIFAll),
		fetchLoanAll:        envBool(jobFetchLoanAll),
		fetchSavingsAll:     envBool(jobFetchSavingsAll),
		fetchTimeDepositAll: envBool(jobFetchTimeDeposit),
		ingestConcurrency:   concurrency,
	}
}

func loadDateWindow(now time.Time) (dateWindow, error) {
	startDate := strings.TrimSpace(os.Getenv(envIngestStartDate))
	endDate := strings.TrimSpace(os.Getenv(envIngestEndDate))

	if startDate == "" && endDate == "" {
		return dateWindow{
			start: now.AddDate(0, 0, -7),
			end:   now.AddDate(0, 0, -1),
			asOf:  now.Format("2006-01-02"),
		}, nil
	}

	if startDate == "" || endDate == "" {
		return dateWindow{}, fmt.Errorf("%s and %s must both be set together", envIngestStartDate, envIngestEndDate)
	}

	start, err := parseDateWindowOverride(envIngestStartDate, startDate, now.Location())
	if err != nil {
		return dateWindow{}, err
	}
	end, err := parseDateWindowOverride(envIngestEndDate, endDate, now.Location())
	if err != nil {
		return dateWindow{}, err
	}
	if end.Before(start) {
		return dateWindow{}, errors.New("INGEST_END_DATE must be on or after INGEST_START_DATE")
	}

	return dateWindow{
		start: start,
		end:   end,
		asOf:  now.Format("2006-01-02"),
	}, nil
}

func parseDateWindowOverride(envKey, value string, loc *time.Location) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", value, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s %q: expected YYYY-MM-DD", envKey, value)
	}
	return parsed, nil
}

func runEnabledJobs(ctx context.Context, deps runtimeDeps, cfg runtimeConfig, w dateWindow) error {
	for _, job := range enabledJobNames(cfg) {
		var err error
		switch job {
		case jobIngestEOD:
			err = runIngestEOD(ctx, deps, cfg, w)
		case jobFetchJournalTrx:
			err = runFetchJournalTrx(ctx, deps, w)
		case jobFetchBalanceSheet:
			err = runFetchBalanceSheet(ctx, deps, cfg, w)
		case jobFetchVaultMut:
			err = runFetchVaultMutations(ctx, deps, cfg, w)
		case jobFetchTellerMut:
			err = runFetchTellerMutations(ctx, deps, cfg, w)
		case jobFetchCOAMovements:
			err = runFetchCOAMovements(ctx, deps, cfg, w)
		case jobFetchMasterData:
			err = runFetchMasterData(ctx, deps, w)
		case jobFetchCIFAll:
			err = runFetchCIFAll(ctx, deps, cfg, w)
		case jobFetchLoanAll:
			err = runFetchLoanAll(ctx, deps, cfg, w)
		case jobFetchSavingsAll:
			err = runFetchSavingsAll(ctx, deps, cfg, w)
		case jobFetchTimeDeposit:
			err = runFetchTimeDepositAll(ctx, deps, cfg, w)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func runIngestEOD(ctx context.Context, deps runtimeDeps, cfg runtimeConfig, w dateWindow) error {
	bar := progressbar.Default(int64(daysInWindow(w)), "fetching and ingesting EOD files")
	sem := make(chan struct{}, cfg.ingestConcurrency)
	eodCh := make(chan eodData)
	eodDone := make(chan struct{})
	upsertFailed := atomic.Int32{}
	fetchFailed := atomic.Int32{}

	go func() {
		defer close(eodDone)
		for eod := range eodCh {
			if _, err := deps.store.UpsertEODCSV(ctx, eod.file, eod.date, eod.data); err != nil {
				fmt.Fprintf(os.Stderr, "failed to ingest EOD %s: %v\n", eod.file, err)
				upsertFailed.Add(1)
			}
		}
	}()

	var wg sync.WaitGroup

	for d := w.start; !d.After(w.end); d = d.AddDate(0, 0, 1) {
		currentDate := d
		sem <- struct{}{}
		wg.Go(func() {
			defer func() {
				<-sem
				_ = bar.Add(1)
			}()

			dateStr := currentDate.Format("2006-01-02")
			eodFiles, err := deps.fetch.FetchEODFiles(ctx, dateStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to fetch EOD files for %s: %v\n", dateStr, err)
				fetchFailed.Add(1)
				return
			}

			files := make([]string, 0, len(eodFiles))
			for file := range eodFiles {
				files = append(files, file)
			}
			sort.Strings(files)

			for _, file := range files {
				eodCh <- eodData{
					file: file,
					date: dateStr,
					data: eodFiles[file],
				}
			}
		})
	}

	wg.Wait()
	close(eodCh)
	<-eodDone

	fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	return nil
}

func runFetchJournalTrx(ctx context.Context, deps runtimeDeps, w dateWindow) error {
	bar := progressbar.Default(int64(daysInWindow(w)), "fetching and ingesting journal transaction reports")

	for d := w.start; !d.After(w.end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		journal, err := deps.fetch.FetchJournalTransactionReport(ctx, "", dateStr, dateStr)
		if err != nil {
			return fmt.Errorf("failed to fetch journal transaction report for %s: %w", dateStr, err)
		}

		if _, err := deps.store.UpsertCSV(ctx, "journal_transactions", "Journal Transaction Report csv", dateStr, journal); err != nil {
			return fmt.Errorf("failed to ingest journal transaction report for %s: %w", dateStr, err)
		}

		_ = bar.Add(1)
	}

	return nil
}

func runFetchBalanceSheet(ctx context.Context, deps runtimeDeps, cfg runtimeConfig, w dateWindow) error {
	branches := reportBranches()
	bar := progressbar.Default(int64(len(branches)*daysInWindow(w)), "fetching and ingesting balance sheet reports")

	upsertFailed := atomic.Int32{}
	fetchFailed := atomic.Int32{}

	sem := make(chan struct{}, cfg.ingestConcurrency)
	var wg sync.WaitGroup

balanceSheetLoop:
	for d := w.start; !d.After(w.end); d = d.AddDate(0, 0, 1) {
		currentDateStr := d.Format("2006-01-02")

		for _, branch := range branches {
			currentBranch := branch
			currentDate := currentDateStr

			select {
			case <-ctx.Done():
				break balanceSheetLoop
			case sem <- struct{}{}:
				wg.Go(func() {
					defer func() {
						<-sem
						_ = bar.Add(1)
					}()

					report, err := deps.fetch.FetchBalanceSheetReport(ctx, currentBranch, currentDate)
					if err != nil {
						fetchFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to fetch balance sheet report for branch %s on %s: %v\n", currentBranch, currentDate, err)
						return
					}

					if _, err := deps.store.UpsertBalanceSheetCSV(ctx, "Balance Sheet Report csv", currentDate, currentBranch, report); err != nil {
						upsertFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to upsert balance sheet report for branch %s on %s: %v\n", currentBranch, currentDate, err)
					}
				})
			}
		}
	}

	wg.Wait()
	fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	return nil
}

func runFetchVaultMutations(ctx context.Context, deps runtimeDeps, cfg runtimeConfig, w dateWindow) error {
	return runFetchBranchScopedReport(
		ctx,
		deps,
		cfg,
		w,
		"fetching and ingesting vault mutation reports",
		"Vault Mutation Report csv",
		"vault_mutations",
		func(ctx context.Context, branch, date string) (string, error) {
			return deps.fetch.FetchVaultMutationReport(ctx, branch, date, date)
		},
	)
}

func runFetchTellerMutations(ctx context.Context, deps runtimeDeps, cfg runtimeConfig, w dateWindow) error {
	return runFetchBranchScopedReport(
		ctx,
		deps,
		cfg,
		w,
		"fetching and ingesting teller mutation reports",
		"Teller Mutation Report (Teller's Blotter) csv",
		"teller_mutations",
		func(ctx context.Context, branch, date string) (string, error) {
			return deps.fetch.FetchTellerMutationReport(ctx, branch, date, date)
		},
	)
}

func runFetchBranchScopedReport(
	ctx context.Context,
	deps runtimeDeps,
	cfg runtimeConfig,
	w dateWindow,
	progressLabel, sourceFile, tableName string,
	fetchReport func(context.Context, string, string) (string, error),
) error {
	branches := reportBranches()
	bar := progressbar.Default(int64(len(branches)*daysInWindow(w)), progressLabel)

	upsertFailed := atomic.Int32{}
	fetchFailed := atomic.Int32{}

	sem := make(chan struct{}, cfg.ingestConcurrency)
	var wg sync.WaitGroup

branchReportLoop:
	for d := w.start; !d.After(w.end); d = d.AddDate(0, 0, 1) {
		currentDateStr := d.Format("2006-01-02")

		for _, branch := range branches {
			currentBranch := branch
			currentDate := currentDateStr

			select {
			case <-ctx.Done():
				break branchReportLoop
			case sem <- struct{}{}:
				wg.Go(func() {
					defer func() {
						<-sem
						_ = bar.Add(1)
					}()

					report, err := fetchReport(ctx, currentBranch, currentDate)
					if err != nil {
						fetchFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to fetch %s for branch %s on %s: %v\n", sourceFile, currentBranch, currentDate, err)
						return
					}

					if _, err := deps.store.UpsertBranchScopedCSV(ctx, tableName, sourceFile, currentDate, currentBranch, report); err != nil {
						upsertFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to upsert %s for branch %s on %s: %v\n", sourceFile, currentBranch, currentDate, err)
					}
				})
			}
		}
	}

	wg.Wait()
	fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	return nil
}

func runFetchCOAMovements(ctx context.Context, deps runtimeDeps, cfg runtimeConfig, w dateWindow) error {
	accounts, err := deps.client.FetchAccountCodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch account codes: %w", err)
	}

	bar := progressbar.Default(int64(len(accounts)*daysInWindow(w)), "fetching and ingesting COA movements")

	upsertFailed := atomic.Int32{}
	fetchFailed := atomic.Int32{}

	sem := make(chan struct{}, cfg.ingestConcurrency)
	var wg sync.WaitGroup

coaLoop:
	for d := w.start; !d.After(w.end); d = d.AddDate(0, 0, 1) {
		currentDate := d
		dateStr := currentDate.Format("2006-01-02")

		for accCode := range accounts {
			currentAccCode := accCode

			select {
			case <-ctx.Done():
				break coaLoop
			case sem <- struct{}{}:
				wg.Go(func() {
					defer func() {
						<-sem
						_ = bar.Add(1)
					}()

					coaData, err := deps.fetch.FetchCoAMovementReport(ctx, currentAccCode, "", dateStr, dateStr)
					if err != nil {
						fetchFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to fetch COA movements for %s: %v\n", currentAccCode, err)
						return
					}

					if _, err := deps.store.UpsertCSV(ctx, "coa_movements", fmt.Sprintf("COA Movement Report csv - %s", currentAccCode), dateStr, coaData); err != nil {
						upsertFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to upsert COA movements for %s: %v\n", currentAccCode, err)
					}
				})
			}
		}
	}

	wg.Wait()
	fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	return nil
}

func runFetchMasterData(ctx context.Context, deps runtimeDeps, w dateWindow) error {
	fmt.Printf("fetching and ingesting master data...\n")

	cifsMasterData, err := deps.fetch.FetchCIFMasterDataRaw(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch CIF master data: %w", err)
	}

	if _, err := deps.store.UpsertJSON(ctx, "raw_cif_master_data", "/tabungan/inquiry/rekening//listvalues", w.asOf, cifsMasterData, nil); err != nil {
		return fmt.Errorf("failed to upsert CIF master data: %w", err)
	}

	timeDepositMasterData, err := deps.fetch.FetchTimeDepositMasterDataRaw(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch time deposit master data: %w", err)
	}

	if _, err := deps.store.UpsertJSON(ctx, "raw_time_deposit_master_data", "/deposito/inquiry/rekening//listvalues", w.asOf, timeDepositMasterData, nil); err != nil {
		return fmt.Errorf("failed to upsert time deposit master data: %w", err)
	}

	loanMasterData, err := deps.fetch.FetchLoanMasterDataRaw(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch loan master data: %w", err)
	}

	if _, err := deps.store.UpsertJSON(ctx, "raw_loan_master_data", "/pinjaman/inquiry/rekening//listvalues", w.asOf, loanMasterData, nil); err != nil {
		return fmt.Errorf("failed to upsert loan master data: %w", err)
	}

	fmt.Printf("done\n")
	return nil
}

func runFetchCIFAll(ctx context.Context, deps runtimeDeps, cfg runtimeConfig, w dateWindow) error {
	cifs, err := deps.fetch.FetchCIFList(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch CIF list: %w", err)
	}

	bar := progressbar.Default(int64(len(cifs)), "fetching CIFs")
	fetchFailed := atomic.Int32{}
	upsertFailed := atomic.Int32{}

	sem := make(chan struct{}, cfg.ingestConcurrency)
	var wg sync.WaitGroup

cifLoop:
	for _, cifNo := range cifs {
		currentCIFNo := cifNo

		select {
		case <-ctx.Done():
			break cifLoop
		case sem <- struct{}{}:
			wg.Go(func() {
				defer func() {
					<-sem
					_ = bar.Add(1)
				}()

				if cfg.jsonIngest {
					payload, err := deps.fetch.FetchCIFDetailRaw(ctx, currentCIFNo)
					if err != nil {
						fetchFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to fetch CIF %s: %v\n", currentCIFNo, err)
						return
					}

					if _, err := deps.store.UpsertJSON(ctx, "raw_cif", "/cif/inquiry/cif/cif", w.asOf, payload, []string{"nocif", "no_cif", "id"}); err != nil {
						upsertFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to upsert CIF %s: %v\n", currentCIFNo, err)
					}
					return
				}

				cif, err := deps.fetch.FetchCIFDetail(ctx, currentCIFNo)
				if err != nil {
					fetchFailed.Add(1)
					fmt.Fprintf(os.Stderr, "failed to fetch CIF %s: %v\n", currentCIFNo, err)
					return
				}

				if err := deps.store.UpsertCIF(ctx, cif); err != nil {
					upsertFailed.Add(1)
					fmt.Fprintf(os.Stderr, "failed to upsert CIF %s: %v\n", currentCIFNo, err)
				}
			})
		}
	}

	wg.Wait()
	fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	return nil
}

func runFetchLoanAll(ctx context.Context, deps runtimeDeps, cfg runtimeConfig, w dateWindow) error {
	loanAccounts := make([]string, 0, 20000) // preallocate with an estimated size to reduce reallocations
	statuses := []string{"Aktif", "Closed", "HT", "WO"}
	for _, status := range statuses {
		accounts, err := deps.fetch.FetchLoanAccounts(ctx, status)
		if err != nil {
			return fmt.Errorf("failed to fetch loan accounts with status %s: %w", status, err)
		}
		loanAccounts = append(loanAccounts, accounts...)
	}

	bar := progressbar.Default(int64(len(loanAccounts)), "fetching loans")
	fetchFailed := atomic.Int32{}
	upsertFailed := atomic.Int32{}

	sem := make(chan struct{}, cfg.ingestConcurrency)
	var wg sync.WaitGroup

loanLoop:
	for _, loanID := range loanAccounts {
		currentLoanID := loanID

		select {
		case <-ctx.Done():
			break loanLoop
		case sem <- struct{}{}:
			wg.Go(func() {
				defer func() {
					<-sem
					_ = bar.Add(1)
				}()

				if cfg.jsonIngest {
					payload, err := deps.fetch.FetchLoanDetailRaw(ctx, currentLoanID)
					if err != nil {
						fetchFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to fetch loan %s: %v\n", currentLoanID, err)
						return
					}

					if _, err := deps.store.UpsertJSON(ctx, "raw_loans", "/pinjaman/inquiry/rekening/pinjaman", w.asOf, payload, []string{"id", "nopk", "no_pk"}); err != nil {
						upsertFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to upsert loan %s: %v\n", currentLoanID, err)
					}
					return
				}

				loan, err := deps.fetch.FetchLoansDetail(ctx, currentLoanID)
				if err != nil {
					fetchFailed.Add(1)
					fmt.Fprintf(os.Stderr, "failed to fetch loan %s: %v\n", currentLoanID, err)
					return
				}

				if err := deps.store.UpsertLoan(ctx, loan); err != nil {
					upsertFailed.Add(1)
					fmt.Fprintf(os.Stderr, "failed to upsert loan %s: %v\n", currentLoanID, err)
				}
			})
		}
	}

	wg.Wait()
	fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	return nil
}

func runFetchSavingsAll(ctx context.Context, deps runtimeDeps, cfg runtimeConfig, w dateWindow) error {
	savingAccounts, err := deps.fetch.FetchSavingsAccounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch savings accounts: %w", err)
	}

	bar := progressbar.Default(int64(len(savingAccounts)), "fetching savings")
	fetchFailed := atomic.Int32{}
	upsertFailed := atomic.Int32{}

	sem := make(chan struct{}, cfg.ingestConcurrency)
	var wg sync.WaitGroup

savingLoop:
	for _, savingID := range savingAccounts {
		currentSavingID := savingID

		select {
		case <-ctx.Done():
			break savingLoop
		case sem <- struct{}{}:
			wg.Go(func() {
				defer func() {
					<-sem
					_ = bar.Add(1)
				}()

				if cfg.jsonIngest {
					payload, err := deps.fetch.FetchSavingDetailRaw(ctx, currentSavingID)
					if err != nil {
						fetchFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to fetch saving %s: %v\n", currentSavingID, err)
						return
					}

					if _, err := deps.store.UpsertJSON(ctx, "raw_savings", "/tabungan/inquiry/rekening/tabungan", w.asOf, payload, []string{"norekening", "no_rekening", "id"}); err != nil {
						upsertFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to upsert saving %s: %v\n", currentSavingID, err)
					}
					return
				}

				saving, err := deps.fetch.FetchSavingsDetail(ctx, currentSavingID)
				if err != nil {
					fetchFailed.Add(1)
					fmt.Fprintf(os.Stderr, "failed to fetch saving %s: %v\n", currentSavingID, err)
					return
				}

				if err := deps.store.UpsertSaving(ctx, saving); err != nil {
					upsertFailed.Add(1)
					fmt.Fprintf(os.Stderr, "failed to upsert saving %s: %v\n", currentSavingID, err)
				}
			})
		}
	}

	wg.Wait()
	fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	return nil
}

func runFetchTimeDepositAll(ctx context.Context, deps runtimeDeps, cfg runtimeConfig, w dateWindow) error {
	if !cfg.jsonIngest {
		return fmt.Errorf("JSON_INGEST must be true to ingest time deposit data")
	}

	timeDepositAccounts, err := deps.fetch.FetchTimeDepositAccounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch time deposit accounts: %w", err)
	}

	bar := progressbar.Default(int64(len(timeDepositAccounts)), "fetching time deposits")
	fetchFailed := atomic.Int32{}
	upsertFailed := atomic.Int32{}

	sem := make(chan struct{}, cfg.ingestConcurrency)
	var wg sync.WaitGroup

timeDepositLoop:
	for _, accountID := range timeDepositAccounts {
		currentAccountID := accountID

		select {
		case <-ctx.Done():
			break timeDepositLoop
		case sem <- struct{}{}:
			wg.Go(func() {
				defer func() {
					<-sem
					_ = bar.Add(1)
				}()

				payload, err := deps.fetch.FetchTimeDepositDetailRaw(ctx, currentAccountID)
				if err != nil {
					fetchFailed.Add(1)
					fmt.Fprintf(os.Stderr, "failed to fetch time deposit %s: %v\n", currentAccountID, err)
					return
				}

				if _, err := deps.store.UpsertJSON(ctx, "raw_time_deposits", "/deposito/inquiry/rekening/deposito", w.asOf, payload, []string{"id", "norekening", "no_rekening"}); err != nil {
					upsertFailed.Add(1)
					fmt.Fprintf(os.Stderr, "failed to upsert time deposit %s: %v\n", currentAccountID, err)
				}
			})
		}
	}

	wg.Wait()
	fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	return nil
}

func enabledJobNames(cfg runtimeConfig) []string {
	jobNames := make([]string, 0, 11)
	if cfg.ingestEOD {
		jobNames = append(jobNames, jobIngestEOD)
	}

	if cfg.fetchJournalTrx {
		jobNames = append(jobNames, jobFetchJournalTrx)
	}

	if cfg.fetchBalanceSheet {
		jobNames = append(jobNames, jobFetchBalanceSheet)
	}

	if cfg.fetchVaultMut {
		jobNames = append(jobNames, jobFetchVaultMut)
	}

	if cfg.fetchTellerMut {
		jobNames = append(jobNames, jobFetchTellerMut)
	}

	if cfg.fetchCOAMovements {
		jobNames = append(jobNames, jobFetchCOAMovements)
	}

	if cfg.fetchMasterData {
		jobNames = append(jobNames, jobFetchMasterData)
	}

	if cfg.fetchCIFAll {
		jobNames = append(jobNames, jobFetchCIFAll)
	}

	if cfg.fetchLoanAll {
		jobNames = append(jobNames, jobFetchLoanAll)
	}

	if cfg.fetchSavingsAll {
		jobNames = append(jobNames, jobFetchSavingsAll)
	}

	if cfg.fetchTimeDepositAll {
		jobNames = append(jobNames, jobFetchTimeDeposit)
	}

	return jobNames
}

func reportBranches() []string {
	branches := make([]string, 0, 9)
	for branch := 0; branch <= 8; branch++ {
		branches = append(branches, fmt.Sprintf("%03d", branch))
	}
	return branches
}

func daysInWindow(w dateWindow) int {
	if w.end.Before(w.start) {
		return 0
	}
	return int(w.end.Sub(w.start).Hours()/24) + 1
}

func requireEnv(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return "", fmt.Errorf("environment variable %q is required", key)
	}
	return val, nil
}

func errorExit(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func envBool(key string) bool {
	val := os.Getenv(key)
	if val == "" {
		return false
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}
	return parsed
}

func envInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return parsed
}
