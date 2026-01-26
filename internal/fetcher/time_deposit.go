package fetcher

import "context"

func (f *Fetcher) FetchTimeDepositPlacementReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Time Deposit Placement Report csv",
		branch,
		"ALL", // time deposit product, ALL for all products
		startDate,
		endDate,
	)
}

func (f *Fetcher) FetchTimeDepositClosingReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Time Deposit Closing Report csv",
		branch,
		"ALL", // time deposit product, ALL for all products
		startDate,
		endDate,
	)
}

func (f *Fetcher) FetchTimeDepositWillMatureReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Time Deposit Will Mature Report",
		branch,
		"ALL", // time deposit product, ALL for all products
		startDate,
		endDate,
	)
}

func (f *Fetcher) FetchTimeDepositsRolloverAndInterestChangesReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Time Deposits Rollover and Interest Changes Report csv",
		branch,
		"", // time deposit product, empty for ALL
		startDate,
		endDate,
	)
}
