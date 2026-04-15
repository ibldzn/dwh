package db

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/ibldzn/dwh-v2/internal/str"
)

const balanceSheetTableName = "balance_sheet_reports"

var balanceSheetAmountColumns = map[string]struct{}{
	"beginning_balance":  {},
	"debit_transaction":  {},
	"credit_transaction": {},
	"last_balance":       {},
}

type preparedCSVData struct {
	columns []string
	rows    [][]string
}

// UpsertEODCSV ingests a CSV content into an eod_* table inferred from fileName.
// It tolerates changing headers by adding new columns on the fly.
func (s *Store) UpsertEODCSV(ctx context.Context, fileName, eodDate, content string) (int, error) {
	if strings.TrimSpace(fileName) == "" {
		return 0, errors.New("file name is required")
	}
	table := eodTableName(fileName)
	return s.UpsertCSV(ctx, table, fileName, eodDate, content)
}

// UpsertCSV ingests a CSV content into the provided table name.
// It tolerates changing headers by adding new columns on the fly.
// sourceFile and asOfDate are optional metadata; pass empty strings to skip.
func (s *Store) UpsertCSV(ctx context.Context, tableName, sourceFile, asOfDate, content string) (int, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("db store is nil")
	}
	if strings.TrimSpace(tableName) == "" {
		return 0, errors.New("table name is required")
	}
	if strings.TrimSpace(content) == "" {
		return 0, nil
	}

	prepared, err := prepareCSVData(content)
	if err != nil {
		return 0, err
	}
	if len(prepared.columns) == 0 {
		return 0, nil
	}

	return s.upsertPreparedCSV(ctx, tableName, sourceFile, asOfDate, prepared)
}

// UpsertBalanceSheetCSV ingests balance sheet CSV content for a single branch/date snapshot.
func (s *Store) UpsertBalanceSheetCSV(ctx context.Context, sourceFile, asOfDate, branch, content string) (int, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("db store is nil")
	}
	if strings.TrimSpace(branch) == "" {
		return 0, errors.New("branch is required")
	}
	if strings.TrimSpace(content) == "" {
		return 0, nil
	}

	prepared, err := prepareBalanceSheetCSVData(content, branch)
	if err != nil {
		return 0, err
	}
	if len(prepared.columns) == 0 {
		return 0, nil
	}

	return s.upsertPreparedCSV(ctx, balanceSheetTableName, sourceFile, asOfDate, prepared)
}

func (s *Store) upsertPreparedCSV(
	ctx context.Context,
	tableName, sourceFile, asOfDate string,
	prepared preparedCSVData,
) (int, error) {
	if err := ensureCSVTable(ctx, s.db, tableName, prepared.columns); err != nil {
		return 0, err
	}

	insertCols := append([]string{"_row_hash", "source_file", "row_index", "as_of_date"}, prepared.columns...)
	insertCols = append(insertCols, "ingested_at")
	placeholders := placeholders(len(insertCols))
	updateCols := make([]string, 0, len(insertCols))
	for _, col := range insertCols {
		if col == "_row_hash" {
			continue
		}
		updateCols = append(updateCols, fmt.Sprintf("%s = VALUES(%s)", col, col))
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s",
		tableName,
		strings.Join(insertCols, ","),
		placeholders,
		strings.Join(updateCols, ","),
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for i, row := range prepared.rows {
		rowHash := hashCSVRow(sourceFile, asOfDate, row)
		args := make([]any, 0, len(insertCols))
		args = append(args, rowHash, sourceFile, i+1, asOfDate)
		for _, value := range row {
			if value == "" {
				args = append(args, nil)
			} else {
				args = append(args, value)
			}
		}
		args = append(args, time.Now().UTC())

		if _, err := stmt.ExecContext(ctx, args...); err != nil {
			return count, fmt.Errorf("upsert %s row %d: %w", sourceFile, i+1, err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return count, err
	}

	return count, nil
}

func prepareCSVData(content string) (preparedCSVData, error) {
	records, headers, err := parseCSV(content)
	if err != nil {
		return preparedCSVData{}, err
	}
	if len(headers) == 0 {
		return preparedCSVData{}, nil
	}

	columns := sanitizeHeaders(headers)
	if len(columns) == 0 {
		return preparedCSVData{}, nil
	}

	rows := make([][]string, 0, len(records))
	for _, record := range records {
		rows = append(rows, trimCSVRecord(record, len(columns)))
	}

	return preparedCSVData{
		columns: columns,
		rows:    rows,
	}, nil
}

func prepareBalanceSheetCSVData(content, branch string) (preparedCSVData, error) {
	records, headers, err := parseCSV(content)
	if err != nil {
		return preparedCSVData{}, err
	}
	if len(headers) == 0 {
		return preparedCSVData{}, nil
	}

	sanitizedHeaders := sanitizeHeaders(headers)
	keptIndexes := make([]int, 0, len(sanitizedHeaders))
	columns := make([]string, 0, len(sanitizedHeaders))
	columns = append(columns, "branch")
	for idx := range sanitizedHeaders {
		sanitizedHeaders[idx] = normalizeBalanceSheetColumnName(sanitizedHeaders[idx])
	}

	for idx, column := range sanitizedHeaders {
		if column == "branch" {
			continue
		}
		keptIndexes = append(keptIndexes, idx)
		columns = append(columns, column)
	}

	rows := make([][]string, 0, len(records))
	for _, record := range records {
		row := make([]string, 0, len(columns))
		row = append(row, strings.TrimSpace(branch))
		for _, idx := range keptIndexes {
			value := ""
			if idx < len(record) {
				value = strings.TrimSpace(record[idx])
			}
			if _, ok := balanceSheetAmountColumns[sanitizedHeaders[idx]]; ok {
				value = normalizeBalanceSheetAmount(value)
			}
			row = append(row, value)
		}
		rows = append(rows, row)
	}

	return preparedCSVData{
		columns: columns,
		rows:    rows,
	}, nil
}

func parseCSV(content string) ([][]string, []string, error) {
	r := strings.NewReader(content)
	buffered := bufio.NewReader(r)
	sample, err := buffered.Peek(4096)
	if err != nil && !errors.Is(err, bufio.ErrBufferFull) && !errors.Is(err, io.EOF) {
		return nil, nil, err
	}

	csvReader := csv.NewReader(buffered)
	csvReader.FieldsPerRecord = -1
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.Comma = str.DetectDelimiter(sample)

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(records) == 0 {
		return nil, nil, nil
	}

	headers := make([]string, len(records[0]))
	for i, header := range records[0] {
		headers[i] = strings.TrimSpace(strings.TrimLeft(header, "\uFEFF"))
	}

	return records[1:], headers, nil
}

func trimCSVRecord(record []string, size int) []string {
	row := make([]string, size)
	for idx := range size {
		if idx < len(record) {
			row[idx] = strings.TrimSpace(record[idx])
		}
	}
	return row
}

func normalizeBalanceSheetAmount(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	negative := false
	if len(value) >= 2 && strings.HasPrefix(value, "<") && strings.HasSuffix(value, ">") {
		negative = true
		value = strings.TrimSpace(value[1 : len(value)-1])
	}

	value = strings.ReplaceAll(value, ",", "")
	if value == "" {
		return ""
	}

	if negative {
		value = strings.TrimPrefix(value, "-")
		return "-" + value
	}

	return value
}

func normalizeBalanceSheetColumnName(column string) string {
	if column == "co_a_no" {
		return "coa_no"
	}
	return column
}

func eodTableName(fileName string) string {
	base := fileName
	if idx := strings.LastIndex(base, "."); idx >= 0 {
		base = base[:idx]
	}
	base = snakeCase(base)
	if base == "" {
		base = "file"
	}
	return "eod_" + base
}

// csvTableName builds a prefixed table name from a file name.
func csvTableName(prefix, fileName string) string {
	base := fileName
	if idx := strings.LastIndex(base, "."); idx >= 0 {
		base = base[:idx]
	}
	base = snakeCase(base)
	if base == "" {
		base = "file"
	}
	if prefix == "" {
		return base
	}
	return prefix + "_" + base
}

func sanitizeHeaders(headers []string) []string {
	seen := map[string]int{}
	cols := make([]string, 0, len(headers))
	for i, header := range headers {
		col := snakeCase(header)
		if col == "" {
			col = fmt.Sprintf("col_%d", i+1)
		}
		if col[0] >= '0' && col[0] <= '9' {
			col = "col_" + col
		}
		if count := seen[col]; count > 0 {
			col = fmt.Sprintf("%s_%d", col, count+1)
		}
		seen[col]++
		cols = append(cols, col)
	}
	return cols
}

func snakeCase(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(input))
	underscore := false
	for _, r := range input {
		if r >= 'A' && r <= 'Z' {
			if b.Len() > 0 && !underscore {
				b.WriteByte('_')
			}
			b.WriteByte(byte(r + ('a' - 'A')))
			underscore = false
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteByte(byte(r))
			underscore = false
			continue
		}
		if !underscore {
			b.WriteByte('_')
			underscore = true
		}
	}

	out := strings.Trim(b.String(), "_")
	for strings.Contains(out, "__") {
		out = strings.ReplaceAll(out, "__", "_")
	}
	return out
}

func ensureCSVTable(ctx context.Context, db *sql.DB, table string, columns []string) error {
	if err := createCSVTable(ctx, db, table, columns); err != nil {
		return err
	}

	existing, err := listColumns(ctx, db, table)
	if err != nil {
		return err
	}

	missing := make([]string, 0)
	for _, col := range append([]string{"as_of_date"}, columns...) {
		if _, ok := existing[col]; !ok {
			missing = append(missing, col)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	sort.Strings(missing)
	lockName := schemaLockName(table)
	if err := acquireSchemaLock(ctx, db, lockName); err != nil {
		return err
	}
	defer func() {
		_ = releaseSchemaLock(ctx, db, lockName)
	}()

	for _, col := range missing {
		colType := "TEXT"
		if col == "as_of_date" {
			colType = "DATE"
		}
		stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col, colType)
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			if isDuplicateColumnError(err) {
				continue
			}
			return err
		}
	}
	return nil
}

func isDuplicateColumnError(err error) bool {
	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) {
		return myErr.Number == 1060
	}
	return false
}

func schemaLockName(table string) string {
	return "schema_lock:" + table
}

var errSchemaLockTimeout = errors.New("schema lock timeout")

func acquireSchemaLock(ctx context.Context, db *sql.DB, lockName string) error {
	timeout := schemaLockTimeoutSeconds()
	var ok int
	if err := db.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", lockName, timeout).Scan(&ok); err != nil {
		return err
	}
	if ok != 1 {
		return fmt.Errorf("%w %s", errSchemaLockTimeout, lockName)
	}
	return nil
}

func isSchemaLockTimeoutError(err error) bool {
	return errors.Is(err, errSchemaLockTimeout)
}

func releaseSchemaLock(ctx context.Context, db *sql.DB, lockName string) error {
	var ok sql.NullInt64
	if err := db.QueryRowContext(ctx, "SELECT RELEASE_LOCK(?)", lockName).Scan(&ok); err != nil {
		return err
	}
	return nil
}

func schemaLockTimeoutSeconds() int {
	val := strings.TrimSpace(os.Getenv("CSV_SCHEMA_LOCK_TIMEOUT_SECONDS"))
	if val == "" {
		return 60
	}
	parsed, err := strconv.Atoi(val)
	if err != nil || parsed < 1 {
		return 60
	}
	return parsed
}

func createCSVTable(ctx context.Context, db *sql.DB, table string, columns []string) error {
	cols := make([]string, 0, len(columns))
	for _, col := range columns {
		cols = append(cols, fmt.Sprintf("%s TEXT", col))
	}
	columnDefs := strings.Join(cols, ",")
	if columnDefs != "" {
		columnDefs += ","
	}

	stmt := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s ("+
			"_row_hash CHAR(64) PRIMARY KEY,"+
			"source_file TEXT,"+
			"row_index BIGINT,"+
			"as_of_date DATE,"+
			"%s"+
			"ingested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP"+
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
		table,
		columnDefs,
	)

	_, err := db.ExecContext(ctx, stmt)
	return err
}

func listColumns(ctx context.Context, db *sql.DB, table string) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, "SHOW COLUMNS FROM "+table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make(map[string]struct{})
	for rows.Next() {
		var field, colType, null, key, defVal, extra sql.NullString
		if err := rows.Scan(&field, &colType, &null, &key, &defVal, &extra); err != nil {
			return nil, err
		}
		if field.Valid {
			cols[field.String] = struct{}{}
		}
	}
	return cols, rows.Err()
}

func hashCSVRow(sourceFile, asOfDate string, values []string) string {
	h := sha256.New()
	_, _ = h.Write([]byte(sourceFile))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(asOfDate))
	_, _ = h.Write([]byte{0})
	for _, v := range values {
		_, _ = h.Write([]byte(v))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
