package fetcher

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ibldzn/dwh-v2/internal/fincloud"
)

type Fetcher struct {
	client *fincloud.Client
}

func NewFetcher(client *fincloud.Client) (*Fetcher, error) {
	if client == nil {
		return nil, errors.New("missing fincloud client")
	}

	return &Fetcher{
		client: client,
	}, nil
}

func (f *Fetcher) FetchEODFiles(ctx context.Context, date string) (map[string]string, error) {
	sessionId, ok := fincloud.SessionIDFromContext(ctx)
	if !ok || sessionId == "" {
		return nil, errors.New("missing session ID in context")
	}

	// parse date as time.Time to validate format
	asOf, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}

	// format date back to string in YYYYMMDD format
	date = asOf.Format("20060102")

	req, err := f.client.NewRequestWithSessionID(ctx, http.MethodGet, "/system/downloaderlaporan/pembuatan/loadorDownload", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("file", date)
	q.Set("jenis", "Folder")
	q.Set("pathfolder", "/app/report/daily")
	req.URL.RawQuery = q.Encode()

	var intermediate struct {
		Data struct {
			Result struct {
				List []struct {
					File  string `json:"file"`
					Jenis string `json:"jenis"`
				} `json:"list"`
			} `json:"result"`
		} `json:"data"`
		Status string `json:"status"`
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := fincloud.DecodeJSON(resp, &intermediate); err != nil {
		return nil, err
	}

	downloadFile := func(fileName string) (string, error) {
		req, err := f.client.NewRequestWithSessionID(ctx, http.MethodGet, "/system/downloaderlaporan/download.php", nil)
		if err != nil {
			return "", err
		}

		q := req.URL.Query()
		q.Set("file", fileName)
		q.Set("path", "/app/report/daily/"+date)
		req.URL.RawQuery = q.Encode()

		resp, err := f.client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		content = bytes.TrimPrefix(content, []byte("\uFEFF")) // remove BOM if exists

		return string(content), nil
	}

	files := make(map[string]string)

	hasMergedOutstandings := false
	detailOutstandings := make([]string, 0, 1*1024*1024) // preallocate 1MB
	detailOutstandingHeader := ""

	for _, file := range intermediate.Data.Result.List {
		if file.Jenis != "File" {
			continue
		}

		if file.File == "DetailOutstandingRekeningPinjaman.csv" {
			hasMergedOutstandings = true
		} else if !hasMergedOutstandings && strings.HasPrefix(file.File, "DetailOutstandingRekeningPinjaman_") {
			content, err := downloadFile(file.File)
			if err != nil {
				return nil, err
			}

			// validate each headers are the same
			lines := strings.SplitN(content, "\n", 2)
			if len(lines) == 0 {
				return nil, fmt.Errorf("unexpected empty content in %s", file.File)
			}

			if detailOutstandingHeader == "" {
				detailOutstandingHeader = lines[0]
			}

			if lines[0] != detailOutstandingHeader {
				return nil, fmt.Errorf("mismatched header in %s", file.File)
			}

			// append content without header
			if len(lines) > 1 {
				detailOutstandings = append(detailOutstandings, lines[1])
			}

			continue
		}

		content, err := downloadFile(file.File)
		if err != nil {
			return nil, err
		}

		files[file.File] = content
	}

	if !hasMergedOutstandings {
		mergedContent := detailOutstandingHeader + "\n" + strings.Join(detailOutstandings, "\n")
		files["DetailOutstandingRekeningPinjaman.csv"] = mergedContent
	}

	return files, nil
}

func (f *Fetcher) FetchCIFOpeningReport(ctx context.Context, branch, startDate, endDate string) (string, error) {
	return f.client.DownloadReport(
		ctx,
		"CIF Opening Report",
		branch,
		startDate,
		endDate,
	)
}
