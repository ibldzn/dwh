package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

const maxIdentifierLength = 64

var jsonReservedColumns = map[string]struct{}{
	"_row_key":       {},
	"_row_hash":      {},
	"source":         {},
	"as_of_date":     {},
	"_overflow_json": {},
	"ingested_at":    {},
	"_parent_key":    {},
	"_array_index":   {},
}

var jsonSchemaFull sync.Map

type jsonChildRow struct {
	index  int
	values map[string]any
}

type jsonChildData struct {
	table            string
	columns          []string
	rows             []jsonChildRow
	existingColumns  map[string]struct{}
	persistedColumns []string
	hasOverflow      bool
}

const concurrentDDLRetryExtra = 6

// UpsertJSON ingests a raw JSON payload into a dynamically evolving table schema.
// payload can be an object or an array of objects.
func (s *Store) UpsertJSON(ctx context.Context, tableName, source, asOfDate string, payload any, keyHints []string) (int, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("db store is nil")
	}

	tableName = normalizeTableName(tableName)
	if tableName == "" {
		return 0, errors.New("table name is required")
	}

	items, err := normalizeJSONPayload(payload)
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, nil
	}

	count := 0
	for _, item := range items {
		if item == nil {
			continue
		}

		obj := normalizeJSONObject(item)
		flat, arrays := flattenJSON(obj)

		rowHash := hashJSONValue(obj)
		rowKey := extractRowKey(flat, keyHints)
		if strings.TrimSpace(rowKey) == "" {
			rowKey = rowHash
		}
		if len(rowKey) > 255 {
			rowKey = rowHash
		}

		parentColumns := sortedKeys(flat)
		parentExisting, err := ensureJSONTable(ctx, s.db, tableName, parentColumns)
		if err != nil {
			return count, err
		}

		parentPersisted := selectPersistedColumns(parentColumns, parentExisting)
		parentOverflowJSON, err := buildOverflowJSON(parentColumns, flat, parentExisting, hasColumn(parentExisting, "_overflow_json"))
		if err != nil {
			return count, fmt.Errorf("build parent overflow for %s failed: %w", tableName, err)
		}

		children := buildChildData(tableName, arrays)
		for i := range children {
			childExisting, err := ensureJSONChildTable(ctx, s.db, children[i].table, children[i].columns)
			if err != nil {
				return count, err
			}
			children[i].existingColumns = childExisting
			children[i].persistedColumns = selectPersistedColumns(children[i].columns, childExisting)
			children[i].hasOverflow = hasColumn(childExisting, "_overflow_json")
		}

		maxAttempts := deadlockRetryMax + concurrentDDLRetryExtra
		var lastErr error
		for attempt := 0; attempt <= maxAttempts; attempt++ {
			if attempt > 0 {
				if err := sleepWithBackoff(ctx, attempt); err != nil {
					return count, err
				}
			}

			if err := s.upsertJSONOnce(
				ctx,
				tableName,
				source,
				asOfDate,
				rowKey,
				rowHash,
				flat,
				parentPersisted,
				hasColumn(parentExisting, "_overflow_json"),
				parentOverflowJSON,
				children,
			); err != nil {
				lastErr = err
				if isConcurrentDDLError(err) {
					if err := waitForSchemaUnlock(ctx, s.db, tableName, children); err != nil {
						return count, err
					}
					continue
				}
				if isRetryableMySQLError(err) {
					continue
				}
				return count, err
			}

			lastErr = nil
			break
		}

		if lastErr != nil {
			return count, lastErr
		}

		count++
	}

	return count, nil
}

func (s *Store) upsertJSONOnce(
	ctx context.Context,
	tableName, source, asOfDate, rowKey, rowHash string,
	flat map[string]any,
	parentColumns []string,
	hasParentOverflow bool,
	parentOverflowJSON any,
	children []jsonChildData,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	now := time.Now().UTC()

	parentInsertCols := append([]string{"_row_key", "_row_hash", "source", "as_of_date"}, parentColumns...)
	if hasParentOverflow {
		parentInsertCols = append(parentInsertCols, "_overflow_json")
	}
	parentInsertCols = append(parentInsertCols, "ingested_at")

	parentQuery := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s",
		tableName,
		strings.Join(parentInsertCols, ","),
		placeholders(len(parentInsertCols)),
		buildUpdateList(parentInsertCols, map[string]bool{"_row_key": true}),
	)

	parentArgs := make([]any, 0, len(parentInsertCols))
	parentArgs = append(parentArgs, rowKey, rowHash, nullableString(source), nullableString(asOfDate))
	for _, col := range parentColumns {
		parentArgs = append(parentArgs, flat[col])
	}
	if hasParentOverflow {
		parentArgs = append(parentArgs, parentOverflowJSON)
	}
	parentArgs = append(parentArgs, now)

	if _, err := tx.ExecContext(ctx, parentQuery, parentArgs...); err != nil {
		return err
	}

	for _, child := range children {
		deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE _parent_key = ?", child.table)
		if _, err := tx.ExecContext(ctx, deleteQuery, rowKey); err != nil {
			return err
		}

		if len(child.rows) == 0 {
			continue
		}

		childInsertCols := append([]string{"_parent_key", "_array_index", "source", "as_of_date"}, child.persistedColumns...)
		if child.hasOverflow {
			childInsertCols = append(childInsertCols, "_overflow_json")
		}
		childInsertCols = append(childInsertCols, "ingested_at")

		childQuery := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			child.table,
			strings.Join(childInsertCols, ","),
			placeholders(len(childInsertCols)),
		)

		stmt, err := tx.PrepareContext(ctx, childQuery)
		if err != nil {
			return err
		}

		for _, row := range child.rows {
			rowOverflowJSON, err := buildOverflowJSON(child.columns, row.values, child.existingColumns, child.hasOverflow)
			if err != nil {
				_ = stmt.Close()
				return fmt.Errorf("build child overflow for %s failed: %w", child.table, err)
			}

			childArgs := make([]any, 0, len(childInsertCols))
			childArgs = append(childArgs, rowKey, row.index, nullableString(source), nullableString(asOfDate))
			for _, col := range child.persistedColumns {
				childArgs = append(childArgs, row.values[col])
			}
			if child.hasOverflow {
				childArgs = append(childArgs, rowOverflowJSON)
			}
			childArgs = append(childArgs, now)

			if _, err := stmt.ExecContext(ctx, childArgs...); err != nil {
				_ = stmt.Close()
				return err
			}
		}

		if err := stmt.Close(); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func normalizeJSONPayload(payload any) ([]any, error) {
	if payload == nil {
		return nil, nil
	}

	switch v := payload.(type) {
	case []any:
		return v, nil
	case []map[string]any:
		items := make([]any, 0, len(v))
		for _, m := range v {
			items = append(items, m)
		}
		return items, nil
	case map[string]any:
		return []any{v}, nil
	default:
		rv := reflect.ValueOf(payload)
		if rv.IsValid() && rv.Kind() == reflect.Slice {
			items := make([]any, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				items = append(items, rv.Index(i).Interface())
			}
			return items, nil
		}
		return nil, fmt.Errorf("unsupported JSON payload type %T", payload)
	}
}

func normalizeJSONObject(value any) map[string]any {
	switch v := value.(type) {
	case map[string]any:
		return v
	default:
		return map[string]any{"value": v}
	}
}

func normalizeArrayElement(value any) map[string]any {
	switch v := value.(type) {
	case map[string]any:
		return v
	default:
		return map[string]any{"value": v}
	}
}

func buildChildData(parentTable string, arrays map[string][]any) []jsonChildData {
	arrayNames := make([]string, 0, len(arrays))
	for name := range arrays {
		arrayNames = append(arrayNames, name)
	}
	sort.Strings(arrayNames)

	children := make([]jsonChildData, 0, len(arrayNames))
	for _, arrayName := range arrayNames {
		items := arrays[arrayName]
		rows := make([]jsonChildRow, 0, len(items))
		columnSet := make(map[string]struct{})

		for idx, item := range items {
			flat, _ := flattenJSONObject(normalizeArrayElement(item), false)
			if len(flat) == 0 {
				flat = map[string]any{"value": nil}
			}
			for col := range flat {
				columnSet[col] = struct{}{}
			}
			rows = append(rows, jsonChildRow{
				index:  idx,
				values: flat,
			})
		}

		columns := make([]string, 0, len(columnSet))
		for col := range columnSet {
			columns = append(columns, col)
		}
		sort.Strings(columns)

		children = append(children, jsonChildData{
			table:   normalizeTableName(parentTable + "__" + arrayName),
			columns: columns,
			rows:    rows,
		})
	}

	return children
}

func flattenJSON(obj map[string]any) (map[string]any, map[string][]any) {
	return flattenJSONObject(obj, true)
}

func flattenJSONObject(obj map[string]any, extractArrays bool) (map[string]any, map[string][]any) {
	type flatEntry struct {
		path []string
		raw  string
		val  any
	}
	type arrayEntry struct {
		path   []string
		raw    string
		values []any
	}

	entries := make([]flatEntry, 0)
	arrays := make([]arrayEntry, 0)

	var walk func(value any, path []string)
	walk = func(value any, path []string) {
		switch v := value.(type) {
		case map[string]any:
			keys := make([]string, 0, len(v))
			for k := range v {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				walk(v[k], append(path, k))
			}
		case []any:
			if extractArrays {
				arrays = append(arrays, arrayEntry{
					path:   append([]string(nil), path...),
					raw:    strings.Join(path, "."),
					values: v,
				})
				return
			}
			entries = append(entries, flatEntry{
				path: append([]string(nil), path...),
				raw:  strings.Join(path, "."),
				val:  v,
			})
		default:
			entries = append(entries, flatEntry{
				path: append([]string(nil), path...),
				raw:  strings.Join(path, "."),
				val:  v,
			})
		}
	}

	walk(obj, nil)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].raw < entries[j].raw
	})
	sort.Slice(arrays, func(i, j int) bool {
		return arrays[i].raw < arrays[j].raw
	})

	flat := make(map[string]any, len(entries))
	flatSeen := make(map[string]int)
	for _, entry := range entries {
		base := normalizePathSegments(entry.path)
		if _, ok := jsonReservedColumns[base]; ok {
			base = "col_" + base
		}
		base = truncateWithHash(base, maxIdentifierLength)

		name := base
		if count := flatSeen[base]; count > 0 {
			name = appendSuffix(base, fmt.Sprintf("_%d", count+1), maxIdentifierLength)
		}
		flatSeen[base]++
		flat[name] = stringifyScalar(entry.val)
	}

	arrayMap := make(map[string][]any, len(arrays))
	arraySeen := make(map[string]int)
	for _, entry := range arrays {
		base := normalizePathSegments(entry.path)
		if _, ok := jsonReservedColumns[base]; ok {
			base = "arr_" + base
		}
		base = truncateWithHash(base, maxIdentifierLength)

		name := base
		if count := arraySeen[base]; count > 0 {
			name = appendSuffix(base, fmt.Sprintf("_%d", count+1), maxIdentifierLength)
		}
		arraySeen[base]++
		arrayMap[name] = entry.values
	}

	return flat, arrayMap
}

func normalizePathSegments(segments []string) string {
	if len(segments) == 0 {
		return "value"
	}

	out := make([]string, 0, len(segments))
	for _, seg := range segments {
		s := snakeCase(seg)
		if s == "" {
			s = "field"
		}
		if s[0] >= '0' && s[0] <= '9' {
			s = "field_" + s
		}
		out = append(out, s)
	}
	return strings.Join(out, "__")
}

func stringifyScalar(value any) any {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(data)
	}
}

func extractRowKey(flat map[string]any, keyHints []string) string {
	for _, hint := range keyHints {
		key := snakeCase(hint)
		if val, ok := flat[key]; ok {
			return fmt.Sprint(val)
		}
	}

	keys := []string{"id", "norekening", "no_rekening", "nocif", "no_cif", "nopk", "no_pk"}
	for _, key := range keys {
		if val, ok := flat[key]; ok {
			parsed := fmt.Sprint(val)
			if strings.TrimSpace(parsed) != "" {
				return parsed
			}
		}
	}

	return ""
}

func hashJSONValue(value any) string {
	h := sha256.New()
	writeHashValue(h, value)
	return hex.EncodeToString(h.Sum(nil))
}

func writeHashValue(h hash.Hash, value any) {
	switch v := value.(type) {
	case nil:
		_, _ = h.Write([]byte{0})
	case string:
		_, _ = h.Write([]byte("s"))
		_, _ = h.Write([]byte(v))
	case json.Number:
		_, _ = h.Write([]byte("n"))
		_, _ = h.Write([]byte(v.String()))
	case float64:
		_, _ = h.Write([]byte("f"))
		_, _ = h.Write([]byte(strconv.FormatFloat(v, 'g', -1, 64)))
	case float32:
		_, _ = h.Write([]byte("f"))
		_, _ = h.Write([]byte(strconv.FormatFloat(float64(v), 'g', -1, 32)))
	case bool:
		if v {
			_, _ = h.Write([]byte("t"))
		} else {
			_, _ = h.Write([]byte("f"))
		}
	case map[string]any:
		_, _ = h.Write([]byte("m"))
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			_, _ = h.Write([]byte(k))
			_, _ = h.Write([]byte{0})
			writeHashValue(h, v[k])
			_, _ = h.Write([]byte{0})
		}
	case []any:
		_, _ = h.Write([]byte("a"))
		for _, item := range v {
			writeHashValue(h, item)
			_, _ = h.Write([]byte{0})
		}
	default:
		data, err := json.Marshal(v)
		if err != nil {
			_, _ = h.Write([]byte(fmt.Sprint(v)))
			return
		}
		_, _ = h.Write(data)
	}
}

func ensureJSONTable(ctx context.Context, db *sql.DB, table string, columns []string) (map[string]struct{}, error) {
	if err := createJSONTable(ctx, db, table); err != nil {
		return nil, err
	}

	return ensureJSONColumns(
		ctx,
		db,
		table,
		columns,
		[]string{"as_of_date", "_overflow_json"},
		map[string]string{
			"as_of_date":     "DATE",
			"_overflow_json": "LONGTEXT",
		},
	)
}

func ensureJSONChildTable(ctx context.Context, db *sql.DB, table string, columns []string) (map[string]struct{}, error) {
	if err := createJSONChildTable(ctx, db, table); err != nil {
		return nil, err
	}

	return ensureJSONColumns(
		ctx,
		db,
		table,
		columns,
		[]string{"source", "as_of_date", "_overflow_json"},
		map[string]string{
			"source":         "TEXT",
			"as_of_date":     "DATE",
			"_overflow_json": "LONGTEXT",
		},
	)
}

func ensureJSONColumns(
	ctx context.Context,
	db *sql.DB,
	table string,
	requested []string,
	required []string,
	columnTypes map[string]string,
) (map[string]struct{}, error) {
	existing, err := listColumns(ctx, db, table)
	if err != nil {
		return nil, err
	}

	missing := missingColumns(existing, requested, required)
	if len(missing) == 0 {
		return existing, nil
	}

	if _, full := jsonSchemaFull.Load(table); full {
		return existing, nil
	}

	sort.Strings(missing)
	lockName := schemaLockName(table)
	if err := acquireSchemaLock(ctx, db, lockName); err != nil {
		return nil, err
	}
	defer func() {
		_ = releaseSchemaLock(ctx, db, lockName)
	}()

	// Re-read columns after locking to avoid duplicate ALTER races.
	existing, err = listColumns(ctx, db, table)
	if err != nil {
		return nil, err
	}

	missing = missingColumns(existing, requested, required)
	if len(missing) == 0 {
		return existing, nil
	}

	if err := ensureDynamicRowFormat(ctx, db, table); err != nil {
		return nil, err
	}

	for _, col := range missing {
		colType := "TEXT"
		if explicit, ok := columnTypes[col]; ok {
			colType = explicit
		}

		stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col, colType)
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			if isDuplicateColumnError(err) {
				existing[col] = struct{}{}
				continue
			}
			if isRowSizeTooLargeError(err) {
				jsonSchemaFull.Store(table, struct{}{})
				break
			}
			return nil, fmt.Errorf("add column %s to %s failed: %w", col, table, err)
		}
		existing[col] = struct{}{}
	}

	return existing, nil
}

func ensureDynamicRowFormat(ctx context.Context, db *sql.DB, table string) error {
	var rowFormat sql.NullString
	if err := db.QueryRowContext(
		ctx,
		`SELECT ROW_FORMAT
		 FROM information_schema.tables
		 WHERE table_schema = DATABASE() AND table_name = ?`,
		table,
	).Scan(&rowFormat); err != nil {
		return err
	}

	if rowFormat.Valid && strings.EqualFold(rowFormat.String, "DYNAMIC") {
		return nil
	}

	stmt := fmt.Sprintf("ALTER TABLE %s ROW_FORMAT=DYNAMIC", table)
	if _, err := db.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("set ROW_FORMAT=DYNAMIC for %s failed: %w", table, err)
	}

	return nil
}

func missingColumns(existing map[string]struct{}, requested []string, required []string) []string {
	missing := make([]string, 0)
	seen := make(map[string]struct{})

	for _, col := range requested {
		if _, ok := existing[col]; ok {
			continue
		}
		if _, ok := seen[col]; ok {
			continue
		}
		seen[col] = struct{}{}
		missing = append(missing, col)
	}

	for _, col := range required {
		if _, ok := existing[col]; ok {
			continue
		}
		if _, ok := seen[col]; ok {
			continue
		}
		seen[col] = struct{}{}
		missing = append(missing, col)
	}

	return missing
}

func isRowSizeTooLargeError(err error) bool {
	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) {
		return myErr.Number == 1118
	}
	return false
}

func isConcurrentDDLError(err error) bool {
	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) {
		return myErr.Number == 1684
	}
	return false
}

func waitForSchemaUnlock(ctx context.Context, db *sql.DB, parentTable string, children []jsonChildData) error {
	tableSet := map[string]struct{}{
		parentTable: {},
	}
	for _, child := range children {
		tableSet[child.table] = struct{}{}
	}

	tables := make([]string, 0, len(tableSet))
	for table := range tableSet {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	for _, table := range tables {
		lockName := schemaLockName(table)
		if err := acquireSchemaLock(ctx, db, lockName); err != nil {
			return fmt.Errorf("wait schema lock %s failed: %w", lockName, err)
		}
		if err := releaseSchemaLock(ctx, db, lockName); err != nil {
			return fmt.Errorf("release schema lock %s failed: %w", lockName, err)
		}
	}

	return nil
}

func createJSONTable(ctx context.Context, db *sql.DB, table string) error {
	stmt := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s ("+
			"_row_key VARCHAR(255) PRIMARY KEY,"+
			"_row_hash CHAR(64),"+
			"source TEXT,"+
			"as_of_date DATE,"+
			"_overflow_json LONGTEXT,"+
			"ingested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP"+
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC",
		table,
	)

	_, err := db.ExecContext(ctx, stmt)
	return err
}

func createJSONChildTable(ctx context.Context, db *sql.DB, table string) error {
	stmt := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s ("+
			"_parent_key VARCHAR(255) NOT NULL,"+
			"_array_index BIGINT NOT NULL,"+
			"source TEXT,"+
			"as_of_date DATE,"+
			"_overflow_json LONGTEXT,"+
			"ingested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,"+
			"PRIMARY KEY (_parent_key, _array_index)"+
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC",
		table,
	)

	_, err := db.ExecContext(ctx, stmt)
	return err
}

func selectPersistedColumns(columns []string, existing map[string]struct{}) []string {
	persisted := make([]string, 0, len(columns))
	for _, col := range columns {
		if hasColumn(existing, col) {
			persisted = append(persisted, col)
		}
	}
	return persisted
}

func buildOverflowJSON(columns []string, values map[string]any, existing map[string]struct{}, enabled bool) (any, error) {
	if !enabled {
		return nil, nil
	}

	overflow := make(map[string]any)
	for _, col := range columns {
		if hasColumn(existing, col) {
			continue
		}
		value, ok := values[col]
		if !ok {
			continue
		}
		overflow[col] = value
	}
	if len(overflow) == 0 {
		return nil, nil
	}

	payload, err := json.Marshal(overflow)
	if err != nil {
		return nil, err
	}
	return string(payload), nil
}

func hasColumn(existing map[string]struct{}, col string) bool {
	_, ok := existing[col]
	return ok
}

func normalizeTableName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteByte(byte(r + ('a' - 'A')))
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_':
			b.WriteByte(byte(r))
		default:
			b.WriteByte('_')
		}
	}

	out := strings.Trim(b.String(), "_")
	if out == "" {
		return ""
	}
	if out[0] >= '0' && out[0] <= '9' {
		out = "t_" + out
	}
	return truncateWithHash(out, maxIdentifierLength)
}

func truncateWithHash(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 8 {
		return value[:limit]
	}
	hash := shortHash(value)
	suffix := "_" + hash
	if limit <= len(suffix) {
		return suffix[len(suffix)-limit:]
	}
	return value[:limit-len(suffix)] + suffix
}

func appendSuffix(value, suffix string, limit int) string {
	if len(value)+len(suffix) <= limit {
		return value + suffix
	}
	if limit <= len(suffix) {
		return suffix[len(suffix)-limit:]
	}
	return value[:limit-len(suffix)] + suffix
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:4])
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
