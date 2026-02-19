package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// FetchJSONResult fetches an endpoint and returns the raw data.result payload.
func (f *Fetcher) FetchJSONResult(ctx context.Context, method, path string, query url.Values) (any, error) {
	req, err := f.client.NewRequestWithSessionID(ctx, method, path, nil)
	if err != nil {
		return nil, err
	}

	if query != nil {
		req.URL.RawQuery = query.Encode()
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s failed: %s", path, resp.Status)
	}

	var data struct {
		Data struct {
			Result any `json:"result"`
		} `json:"data"`
		Error *struct {
			System string `json:"system"`
		} `json:"error,omitempty"`
		Status string `json:"status"`
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	if data.Status != "ok" {
		if data.Error != nil && data.Error.System != "" {
			return nil, fmt.Errorf("fetch %s failed: %s", path, data.Error.System)
		}
		return nil, fmt.Errorf("fetch %s failed: status %s", path, data.Status)
	}

	return data.Data.Result, nil
}

func (f *Fetcher) FetchCIFDetailRaw(ctx context.Context, cifNo string) (any, error) {
	q := url.Values{}
	q.Set("nocif", cifNo)
	return f.FetchJSONResult(ctx, http.MethodGet, "/cif/inquiry/cif/cif", q)
}

func (f *Fetcher) FetchLoanDetailRaw(ctx context.Context, accountNo string) (any, error) {
	q := url.Values{}
	q.Set("id", accountNo)
	return f.FetchJSONResult(ctx, http.MethodGet, "/pinjaman/inquiry/rekening/pinjaman", q)
}

func (f *Fetcher) FetchSavingDetailRaw(ctx context.Context, accountNo string) (any, error) {
	q := url.Values{}
	q.Set("id", accountNo)
	return f.FetchJSONResult(ctx, http.MethodGet, "/tabungan/inquiry/rekening/tabungan", q)
}

func (f *Fetcher) FetchTimeDepositDetailRaw(ctx context.Context, account string) (any, error) {
	q := url.Values{}
	q.Set("id", account)
	return f.FetchJSONResult(ctx, http.MethodGet, "/deposito/inquiry/rekening/deposito", q)
}

func (f *Fetcher) FetchCIFMasterDataRaw(ctx context.Context) (any, error) {
	return f.FetchJSONResult(ctx, http.MethodGet, "/tabungan/inquiry/rekening//listvalues", nil)
}

func (f *Fetcher) FetchTimeDepositMasterDataRaw(ctx context.Context) (any, error) {
	return f.FetchJSONResult(ctx, http.MethodGet, "/deposito/inquiry/rekening//listvalues", nil)
}

func (f *Fetcher) FetchLoanMasterDataRaw(ctx context.Context) (any, error) {
	return f.FetchJSONResult(ctx, http.MethodGet, "/pinjaman/inquiry/rekening//listvalues", nil)
}
