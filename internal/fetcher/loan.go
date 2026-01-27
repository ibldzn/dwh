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
