package fetcher

import (
	"context"
	"fmt"

	"github.com/ibldzn/dwh-v2/internal/fincloud"
	"github.com/ibldzn/dwh-v2/internal/models"
)

func (f *Fetcher) FetchLoansDetail(ctx context.Context, accountNo string) (models.Loan, error) {
	req, err := f.client.NewRequestWithSessionID(ctx, "GET", "/pinjaman/inquiry/rekening/pinjaman", nil)
	if err != nil {
		return models.Loan{}, err
	}

	q := req.URL.Query()
	q.Set("id", accountNo)
	req.URL.RawQuery = q.Encode()

	resp, err := f.client.Do(req)
	if err != nil {
		return models.Loan{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return models.Loan{}, fmt.Errorf("fetch loan detail failed: %s", resp.Status)
	}

	var data struct {
		Data struct {
			Loan models.Loan `json:"result"`
		} `json:"data"`
		Status string `json:"status"`
	}

	if err := fincloud.DecodeJSON(resp, &data); err != nil {
		return models.Loan{}, err
	}

	if data.Status != "ok" {
		return models.Loan{}, fmt.Errorf("fetch loan detail failed: status %s", data.Status)
	}

	return data.Data.Loan, nil
}

func (f *Fetcher) FetchLoanAccounts(ctx context.Context, status string) ([]string, error) {
	validStatuses := map[string]struct{}{
		"Aktif":  {},
		"Closed": {},
		"HT":     {}, // Charge off
		"WO":     {}, // Write off
	}

	if _, ok := validStatuses[status]; !ok {
		return nil, fmt.Errorf("invalid loan account status: %s", status)
	}

	// /pinjaman/inquiry/rekening/cari?cabang=ALL&jenispinjaman=&pagenumber=0&pagesize=50&rowcount=0&status=Aktif
	req, err := f.client.NewRequestWithSessionID(ctx, "GET", "/pinjaman/inquiry/rekening/cari", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("cabang", "ALL")
	q.Set("jenispinjaman", "")
	q.Set("pagenumber", "0")
	q.Set("pagesize", "9999999999999")
	q.Set("rowcount", "0")
	q.Set("status", status)
	req.URL.RawQuery = q.Encode()

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("fetch loan accounts failed: %s", resp.Status)
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
		return nil, fmt.Errorf("fetch loan accounts failed: status %s", data.Status)
	}

	accounts := make([]string, 0, len(data.Data.Result))
	for _, r := range data.Data.Result {
		accounts = append(accounts, r.ID)
	}

	return accounts, nil
}
