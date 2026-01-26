package fetcher

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ibldzn/dwh-v2/internal/fincloud"
	"github.com/ibldzn/dwh-v2/internal/models"
	"github.com/ibldzn/dwh-v2/internal/str"
)

func (f *Fetcher) FetchCIFDetail(ctx context.Context, cifNo string) (models.CIF, error) {
	req, err := f.client.NewRequestWithSessionID(ctx, "GET", "/cif/inquiry/cif/cif", nil)
	if err != nil {
		return models.CIF{}, err
	}

	q := req.URL.Query()
	q.Set("nocif", cifNo)
	req.URL.RawQuery = q.Encode()

	resp, err := f.client.Do(req)
	if err != nil {
		return models.CIF{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return models.CIF{}, fmt.Errorf("fetch CIF detail failed: %s", resp.Status)
	}

	var data struct {
		Data struct {
			CIF models.CIF `json:"result"`
		} `json:"data"`
		Status string `json:"status"`
	}

	if err := fincloud.DecodeJSON(resp, &data); err != nil {
		return models.CIF{}, err
	}

	if data.Status != "ok" {
		return models.CIF{}, fmt.Errorf("fetch CIF detail failed: status %s", data.Status)
	}

	return data.Data.CIF, nil
}

func (f *Fetcher) FetchCIFList(ctx context.Context) ([]string, error) {
	data, err := f.client.DownloadReport(
		ctx,
		"CIF Opening Report",
		"",
		"1900-01-01",
		time.Now().Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}

	r := strings.NewReader(data)
	buffered := bufio.NewReader(r)
	sample, err := buffered.Peek(4096)
	if err != nil && !errors.Is(err, bufio.ErrBufferFull) && !errors.Is(err, io.EOF) {
		return nil, err
	}

	csvReader := csv.NewReader(buffered)
	csvReader.FieldsPerRecord = -1 // variable number of fields
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.Comma = str.DetectDelimiter(sample)

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	cifNoIdx := -1
	for i, header := range records[0] {
		header = strings.TrimSpace(strings.TrimLeft(header, "\uFEFF")) // remove BOM if present
		if strings.EqualFold(header, "\"CIF No\"") {
			cifNoIdx = i
			break
		}
	}

	if cifNoIdx == -1 {
		return nil, errors.New("cannot find 'CIF No' column in CIF Opening Report")
	}

	var cifNos []string
	for _, record := range records[1:] {
		if cifNoIdx >= len(record) {
			continue
		}

		cifNo := strings.Trim(record[cifNoIdx], "\" ")
		if cifNo != "" {
			cifNos = append(cifNos, cifNo)
		}
	}

	return cifNos, nil
}
