package fetcher

import "context"

func (f *Fetcher) FetchSavingsCashDepositTransactionReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Savings Cash Deposit Transaction Reports csv",
		branch,
		"", // savings product, empty for ALL
		startDate,
		endDate,
	)
}

// https://172.20.57.7/fincloud-taspen-web/system/laporanUmum/data/lap?nm=Saving%20Cash%20Withdrawal%20Transaction%20Report%20csv&type=csv&p=[%22%22,%22%22,%222025-12-1%22,%222025-12-31%22]
func (f *Fetcher) FetchSavingsCashWithdrawalTransactionReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Saving Cash Withdrawal Transaction Report csv",
		branch,
		"", // savings product, empty for ALL
		startDate,
		endDate,
	)
}

// https://172.20.57.7/fincloud-taspen-web/system/laporanUmum/data/lap?nm=Fund%20Transfer%20Report%20csv&type=csv&p=[%22%22,%22%22,%222025-12-1%22,%222025-12-31%22]
func (f *Fetcher) FetchSavingsFundTransferReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"Fund Transfer Report csv",
		branch,
		"", // savings product, empty for ALL
		startDate,
		endDate,
	)
}

func (f *Fetcher) FetchSavingsStandingOrderReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	if branch == "" {
		branch = "ALL"
	}

	return f.client.DownloadReport(
		ctx,
		"Standing Order Report csv",
		branch,
		"", // savings product, empty for all products
		startDate,
		endDate,
		"", // account number, empty for ALL
		"", // status, empty for ALL
	)
}
