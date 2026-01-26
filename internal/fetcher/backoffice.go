package fetcher

import "context"

// nm=Journal%20Transaction%20csv&type=csv&p=[%22%22,%22%%22,%222025-12-29%22,%222025-12-29%22,%22%22,%22%22]
func (f *Fetcher) FetchJournalTransactionReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Journal Transaction csv",
		branch,
		"%", // transaction type, % for ALL
		startDate,
		endDate,
		"", // journal id, empty for ALL
		"", // reference no, empty for ALL
	)
}

func (f *Fetcher) FetchBalanceSheetReport(ctx context.Context, branch, date string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Balance Sheet Report csv",
		branch,
		date,
	)
}

func (f *Fetcher) FetchProfitAndLossStatement(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Profit and Loss Statement csv",
		branch,
		startDate,
		endDate,
	)
}

func (f *Fetcher) FetchCoAMovementReport(ctx context.Context, accountCode, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"CoA Movement Report csv",
		accountCode,
		startDate,
		endDate,
		branch,
	)
}

func (f *Fetcher) FetchFundDistributionReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Fund Distribution Report csv",
		branch,
		"", // savings product, always empty
		"", // fund distribution type, empty for ALL
		"", // transaction type, empty for ALL
		startDate,
		endDate,
	)
}
