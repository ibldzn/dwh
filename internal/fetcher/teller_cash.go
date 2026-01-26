package fetcher

import "context"

func (f *Fetcher) FetchVaultMutationReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Vault Mutation Report csv",
		branch,
		startDate,
		endDate,
	)
}

func (f *Fetcher) FetchTellerMutationReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Teller Mutation Report (Teller's Blotter) csv",
		branch,
		"ALL", // teller id, ALL for all tellers
		startDate,
		endDate,
	)
}
