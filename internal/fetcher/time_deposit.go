package fetcher

import (
	"context"
	"fmt"

	"github.com/ibldzn/dwh-v2/internal/fincloud"
	"github.com/ibldzn/dwh-v2/internal/models"
)

func (f *Fetcher) FetchTimeDepositDetail(ctx context.Context, account string) (models.TimeDeposit, error) {
	req, err := f.client.NewRequestWithSessionID(ctx, "GET", "/deposito/inquiry/rekening/deposito", nil)
	if err != nil {
		return models.TimeDeposit{}, err
	}

	q := req.URL.Query()
	q.Set("id", account)
	req.URL.RawQuery = q.Encode()

	resp, err := f.client.Do(req)
	if err != nil {
		return models.TimeDeposit{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return models.TimeDeposit{}, fmt.Errorf("fetch time deposit detail failed: %s", resp.Status)
	}

	var data struct {
		Data struct {
			TimeDeposit models.TimeDeposit `json:"result"`
		} `json:"data"`
		Status string `json:"status"`
	}

	if err := fincloud.DecodeJSON(resp, &data); err != nil {
		return models.TimeDeposit{}, err
	}

	if data.Status != "ok" {
		return models.TimeDeposit{}, fmt.Errorf("fetch time deposit detail failed: status %s", data.Status)
	}

	return data.Data.TimeDeposit, nil
}

func (f *Fetcher) FetchTimeDepositAccounts(ctx context.Context) ([]string, error) {
	req, err := f.client.NewRequestWithSessionID(ctx, "GET", "/deposito/inquiry/rekening/cari", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("cabang", "ALL")
	q.Set("pagenumber", "0")
	q.Set("pagesize", "9999999999999")
	q.Set("rowcount", "0")
	q.Set("status", "")
	req.URL.RawQuery = q.Encode()

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("fetch time deposit accounts failed: %s", resp.Status)
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
		return nil, fmt.Errorf("fetch time deposit accounts failed: status %s", data.Status)
	}

	accounts := make([]string, 0, len(data.Data.Result))
	for _, r := range data.Data.Result {
		accounts = append(accounts, r.ID)
	}

	return accounts, nil
}

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
