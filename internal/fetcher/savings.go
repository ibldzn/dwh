package fetcher

import (
	"context"
	"fmt"

	"github.com/ibldzn/dwh-v2/internal/fincloud"
	"github.com/ibldzn/dwh-v2/internal/models"
)

func (f *Fetcher) FetchSavingsDetail(ctx context.Context, accountNo string) (models.Saving, error) {
	req, err := f.client.NewRequestWithSessionID(ctx, "GET", "/tabungan/inquiry/rekening/tabungan", nil)
	if err != nil {
		return models.Saving{}, err
	}

	q := req.URL.Query()
	q.Set("id", accountNo)
	req.URL.RawQuery = q.Encode()

	resp, err := f.client.Do(req)
	if err != nil {
		return models.Saving{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return models.Saving{}, fmt.Errorf("fetch saving detail failed: %s", resp.Status)
	}

	var data struct {
		Data struct {
			Saving models.Saving `json:"result"`
		} `json:"data"`
		Status string `json:"status"`
	}

	if err := fincloud.DecodeJSON(resp, &data); err != nil {
		return models.Saving{}, err
	}

	if data.Status != "ok" {
		return models.Saving{}, fmt.Errorf("fetch saving detail failed: status %s", data.Status)
	}

	return data.Data.Saving, nil
}

func (f *Fetcher) FetchSavingsAccounts(ctx context.Context) ([]string, error) {
	// /tabungan/inquiry/rekening/cari?cabang=ALL&datamutasi=false&pagenumber=0&pagesize=9999999999999999999&rowcount=0
	req, err := f.client.NewRequestWithSessionID(ctx, "GET", "/tabungan/inquiry/rekening/cari", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("cabang", "ALL")
	q.Set("datamutasi", "false")
	q.Set("pagenumber", "0")
	q.Set("pagesize", "9999999999999")
	q.Set("rowcount", "0")
	req.URL.RawQuery = q.Encode()

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("fetch savings accounts failed: %s", resp.Status)
	}

	var data struct {
		Data struct {
			Result []struct {
				ID string `json:"id"`
			} `json:"result"`
		} `json:"data"`
		Status string `json:"status"`
	}

	if err := fincloud.DecodeJSON(resp, &data); err != nil {
		return nil, err
	}

	if data.Status != "ok" {
		return nil, fmt.Errorf("fetch savings accounts failed: status %s", data.Status)
	}

	accounts := make([]string, 0, len(data.Data.Result))
	for _, acc := range data.Data.Result {
		accounts = append(accounts, acc.ID)
	}

	return accounts, nil
}

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
