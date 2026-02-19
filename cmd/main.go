package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strconv"
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

func main() {
	if err := godotenv.Load(); err != nil {
		errorExit("failed to load .env file: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	client, err := fincloud.NewClient(fincloud.Config{})
	if err != nil {
		errorExit("failed to create fincloud client")
	}

	session, err := client.Login(ctx, fincloud.Credentials{
		Username:   requireEnv("FINCLOUD_USERNAME"),
		Password:   requireEnv("FINCLOUD_PASSWORD"),
		RoleID:     requireEnv("FINCLOUD_ROLE_ID"),
		LocationID: requireEnv("FINCLOUD_LOCATION_ID"),
	})
	if err != nil {
		errorExit("failed to login", err)
	}

	ctx = fincloud.WithFincloudSessionID(ctx, session.ID)

	fetch, err := fetcher.NewFetcher(client)
	if err != nil {
		errorExit("failed to create fetcher", err)
	}

	storeDB, err := db.Open(requireEnv("MYSQL_DSN"))
	if err != nil {
		errorExit("failed to open database: %v", err)
	}
	defer storeDB.Close()

	if err := db.Migrate(ctx, storeDB); err != nil {
		errorExit("failed to migrate database: %v", err)
	}

	store, err := db.NewStore(storeDB)
	if err != nil {
		errorExit("failed to create db store: %v", err)
	}

	jsonIngest := envBool("JSON_INGEST")

	initialDate := time.Now().UTC().AddDate(0, 0, -7)

	if envBool("INGEST_EOD") {
		yesterday := time.Now().UTC().AddDate(0, 0, -1)

		type eodData struct {
			file string
			date string
			data string
		}

		bar := progressbar.Default(int64(yesterday.Sub(initialDate).Hours()/24)+1, "fetching and ingesting EOD files")
		sem := make(chan struct{}, max(envInt("INGEST_CONCURRENCY", 10), 1))
		eodCh := make(chan eodData)
		eodDone := make(chan struct{})
		upsertFailed := atomic.Int32{}
		fetchFailed := atomic.Int32{}

		go func() {
			defer close(eodDone)
			for eod := range eodCh {
				_, err := store.UpsertEODCSV(ctx, eod.file, eod.date, eod.data)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to ingest EOD %s: %v\n", eod.file, err)
					upsertFailed.Add(1)
					continue
				}
			}
		}()

		var wg sync.WaitGroup

		for d := initialDate; !d.After(yesterday); d = d.AddDate(0, 0, 1) {
			sem <- struct{}{}
			wg.Go(func() {
				defer func() {
					<-sem
					_ = bar.Add(1)
				}()

				dateStr := d.Format("2006-01-02")
				eodFiles, err := fetch.FetchEODFiles(ctx, dateStr)
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
	}

	if envBool("FETCH_JOURNAL_TRX") {
		yesterday := time.Now().UTC().AddDate(0, 0, -1)
		bar := progressbar.Default(int64(yesterday.Sub(initialDate).Hours()/24)+1, "fetching and ingesting journal transaction reports")

		for d := initialDate; !d.After(yesterday); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			journal, err := fetch.FetchJournalTransactionReport(ctx, "", dateStr, dateStr)
			if err != nil {
				errorExit("failed to fetch journal transaction report for %s: %v", dateStr, err)
			}

			_, err = store.UpsertCSV(ctx, "journal_transactions", "Journal Transaction Report csv", dateStr, journal)
			if err != nil {
				errorExit("failed to ingest journal transaction report for %s: %v", dateStr, err)
			}

			_ = bar.Add(1)
		}
	}

	if envBool("FETCH_COA_MOVEMENTS") {
		accounts, err := client.FetchAccountCodes(ctx)
		if err != nil {
			errorExit("failed to fetch account codes", err)
		}

		yesterday := time.Now().UTC().AddDate(0, 0, -1)
		bar := progressbar.Default(int64(len(accounts)*int(yesterday.Sub(initialDate).Hours()/24)+1), "fetching and ingesting COA movements")

		upsertFailed := atomic.Int32{}
		fetchFailed := atomic.Int32{}

		concurrency := max(envInt("INGEST_CONCURRENCY", 10), 1)
		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

	coaLoop:
		for d := initialDate; !d.After(yesterday); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			for accCode := range accounts {
				select {
				case <-ctx.Done():
					break coaLoop
				case sem <- struct{}{}:
					wg.Go(func() {
						defer func() {
							<-sem
							_ = bar.Add(1)
						}()

						coaData, err := fetch.FetchCoAMovementReport(ctx, accCode, "", dateStr, dateStr)
						if err != nil {
							fetchFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to fetch COA movements for %s: %v\n", accCode, err)
							return
						}

						// fmt.Printf("fetched COA movements for %s on %s (%d bytes)\n", accCode, dateStr, len(coaData))
						// fmt.Println(coaData)

						if _, err := store.UpsertCSV(ctx, "coa_movements", fmt.Sprintf("COA Movement Report csv - %s", accCode), dateStr, coaData); err != nil {
							upsertFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to upsert COA movements for %s: %v\n", accCode, err)
							return
						}
					})
				}
			}
		}

		wg.Wait()
		fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	}

	if envBool("FETCH_MASTER_DATA") {

	}

	if envBool("FETCH_CIF_ALL") {
		cifs, err := fetch.FetchCIFList(ctx)
		if err != nil {
			errorExit("failed to fetch CIF list", err)
		}

		bar := progressbar.Default(int64(len(cifs)), "fetching CIFs")
		fetchFailed := atomic.Int32{}
		upsertFailed := atomic.Int32{}

		concurrency := max(envInt("INGEST_CONCURRENCY", 10), 1)
		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

	cifLoop:
		for _, cifNo := range cifs {
			select {
			case <-ctx.Done():
				break cifLoop
			case sem <- struct{}{}:
				wg.Go(func() {
					defer func() {
						<-sem
						_ = bar.Add(1)
					}()

					if jsonIngest {
						payload, err := fetch.FetchCIFDetailRaw(ctx, cifNo)
						if err != nil {
							fetchFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to fetch CIF %s: %v\n", cifNo, err)
							return
						}

						if _, err := store.UpsertJSON(ctx, "raw_cif", "/cif/inquiry/cif/cif", time.Now().UTC().Format("2006-01-02"), payload, []string{"nocif", "no_cif", "id"}); err != nil {
							upsertFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to upsert CIF %s: %v\n", cifNo, err)
							return
						}
					} else {
						cif, err := fetch.FetchCIFDetail(ctx, cifNo)
						if err != nil {
							fetchFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to fetch CIF %s: %v\n", cifNo, err)
							return
						}

						if err := store.UpsertCIF(ctx, cif); err != nil {
							upsertFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to upsert CIF %s: %v\n", cifNo, err)
							return
						}
					}
				})
			}
		}

		wg.Wait()
		fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	}

	if envBool("FETCH_LOAN_ALL") {
		loanAccounts, err := fetch.FetchLoanAccounts(ctx, "Aktif")
		if err != nil {
			errorExit("failed to fetch loan accounts", err)
		}

		bar := progressbar.Default(int64(len(loanAccounts)), "fetching loans")
		fetchFailed := atomic.Int32{}
		upsertFailed := atomic.Int32{}

		concurrency := max(envInt("INGEST_CONCURRENCY", 10), 1)
		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

	loanLoop:
		for _, loanID := range loanAccounts {
			select {
			case <-ctx.Done():
				break loanLoop
			case sem <- struct{}{}:
				wg.Go(func() {
					defer func() {
						<-sem
						_ = bar.Add(1)
					}()

					if jsonIngest {
						payload, err := fetch.FetchLoanDetailRaw(ctx, loanID)
						if err != nil {
							fetchFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to fetch loan %s: %v\n", loanID, err)
							return
						}

						if _, err := store.UpsertJSON(ctx, "raw_loans", "/pinjaman/inquiry/rekening/pinjaman", time.Now().UTC().Format("2006-01-02"), payload, []string{"id", "nopk", "no_pk"}); err != nil {
							upsertFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to upsert loan %s: %v\n", loanID, err)
							return
						}
					} else {
						loan, err := fetch.FetchLoansDetail(ctx, loanID)
						if err != nil {
							fetchFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to fetch loan %s: %v\n", loanID, err)
							return
						}

						if err := store.UpsertLoan(ctx, loan); err != nil {
							upsertFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to upsert loan %s: %v\n", loanID, err)
							return
						}
					}
				})
			}
		}

		wg.Wait()
		fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	}

	if envBool("FETCH_SAVINGS_ALL") {
		savingAccounts, err := fetch.FetchSavingsAccounts(ctx)
		if err != nil {
			errorExit("failed to fetch savings accounts", err)
		}

		bar := progressbar.Default(int64(len(savingAccounts)), "fetching savings")
		fetchFailed := atomic.Int32{}
		upsertFailed := atomic.Int32{}

		concurrency := max(envInt("INGEST_CONCURRENCY", 10), 1)
		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

	savingLoop:
		for _, savingID := range savingAccounts {
			select {
			case <-ctx.Done():
				break savingLoop
			case sem <- struct{}{}:
				wg.Go(func() {
					defer func() {
						<-sem
						_ = bar.Add(1)
					}()

					if jsonIngest {
						payload, err := fetch.FetchSavingDetailRaw(ctx, savingID)
						if err != nil {
							fetchFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to fetch saving %s: %v\n", savingID, err)
							return
						}

						if _, err := store.UpsertJSON(ctx, "raw_savings", "/tabungan/inquiry/rekening/tabungan", time.Now().UTC().Format("2006-01-02"), payload, []string{"norekening", "no_rekening", "id"}); err != nil {
							upsertFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to upsert saving %s: %v\n", savingID, err)
							return
						}
					} else {
						saving, err := fetch.FetchSavingsDetail(ctx, savingID)
						if err != nil {
							fetchFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to fetch saving %s: %v\n", savingID, err)
							return
						}

						if err := store.UpsertSaving(ctx, saving); err != nil {
							upsertFailed.Add(1)
							fmt.Fprintf(os.Stderr, "failed to upsert saving %s: %v\n", savingID, err)
							return
						}
					}
				})
			}
		}

		wg.Wait()
		fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	}

	if envBool("FETCH_TIME_DEPOSIT_ALL") {
		if !jsonIngest {
			errorExit("JSON_INGEST must be true to ingest time deposit data")
		}

		timeDepositAccounts, err := fetch.FetchTimeDepositAccounts(ctx)
		if err != nil {
			errorExit("failed to fetch time deposit accounts", err)
		}

		bar := progressbar.Default(int64(len(timeDepositAccounts)), "fetching time deposits")
		fetchFailed := atomic.Int32{}
		upsertFailed := atomic.Int32{}

		concurrency := max(envInt("INGEST_CONCURRENCY", 10), 1)
		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

	timeDepositLoop:
		for _, accountID := range timeDepositAccounts {
			select {
			case <-ctx.Done():
				break timeDepositLoop
			case sem <- struct{}{}:
				wg.Go(func() {
					defer func() {
						<-sem
						_ = bar.Add(1)
					}()

					payload, err := fetch.FetchTimeDepositDetailRaw(ctx, accountID)
					if err != nil {
						fetchFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to fetch time deposit %s: %v\n", accountID, err)
						return
					}

					if _, err := store.UpsertJSON(ctx, "raw_time_deposits", "/deposito/inquiry/rekening/deposito", time.Now().UTC().Format("2006-01-02"), payload, []string{"id", "norekening", "no_rekening"}); err != nil {
						upsertFailed.Add(1)
						fmt.Fprintf(os.Stderr, "failed to upsert time deposit %s: %v\n", accountID, err)
						return
					}
				})
			}
		}

		wg.Wait()
		fmt.Printf("done (fetch failed: %d, upsert failed: %d)\n", fetchFailed.Load(), upsertFailed.Load())
	}
}

func requireEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		errorExit("environment variable %q is required", key)
	}
	return val
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
