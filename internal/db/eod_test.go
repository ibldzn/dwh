package db

import (
	"context"
	"reflect"
	"testing"
)

func TestPrepareBalanceSheetCSVDataDropsCSVBranchAndNormalizesAmounts(t *testing.T) {
	content := stringsJoinLines(
		`Branch|"CoA No"|"Chart of Account"|"Beginning Balance"|"Debit Transaction"|"Credit Transaction"|"Last Balance"`,
		`"001"|"1001"|"Cash on Hand"|"123,456.78"|"<10.00>"|""|"<123,446.78>"`,
	)

	got, err := prepareBalanceSheetCSVData(content, "003")
	if err != nil {
		t.Fatalf("prepareBalanceSheetCSVData failed: %v", err)
	}

	wantColumns := []string{
		"branch",
		"coa_no",
		"chart_of_account",
		"beginning_balance",
		"debit_transaction",
		"credit_transaction",
		"last_balance",
	}
	if !reflect.DeepEqual(got.columns, wantColumns) {
		t.Fatalf("unexpected columns: got %v, want %v", got.columns, wantColumns)
	}

	wantRows := [][]string{{
		"003",
		"1001",
		"Cash on Hand",
		"123456.78",
		"-10.00",
		"",
		"-123446.78",
	}}
	if !reflect.DeepEqual(got.rows, wantRows) {
		t.Fatalf("unexpected rows: got %v, want %v", got.rows, wantRows)
	}
}

func TestNormalizeBalanceSheetAmount(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "positive", input: "123,456.78", want: "123456.78"},
		{name: "negative angle brackets", input: "<123,456.78>", want: "-123456.78"},
		{name: "trimmed negative", input: "  <10.00>  ", want: "-10.00"},
		{name: "blank", input: "", want: ""},
		{name: "spaces only", input: "   ", want: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeBalanceSheetAmount(tc.input); got != tc.want {
				t.Fatalf("unexpected normalized value: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPrepareBranchScopedCSVDataInjectsBranchColumn(t *testing.T) {
	content := stringsJoinLines(
		`Branch|"Transaction ID"|"Amount"`,
		`"001"|"TX-01"|"1500"`,
	)

	got, err := prepareBranchScopedCSVData(content, "003")
	if err != nil {
		t.Fatalf("prepareBranchScopedCSVData failed: %v", err)
	}

	wantColumns := []string{"_branch", "branch", "transaction_i_d", "amount"}
	if !reflect.DeepEqual(got.columns, wantColumns) {
		t.Fatalf("unexpected columns: got %v, want %v", got.columns, wantColumns)
	}

	wantRows := [][]string{{"003", "001", "TX-01", "1500"}}
	if !reflect.DeepEqual(got.rows, wantRows) {
		t.Fatalf("unexpected rows: got %v, want %v", got.rows, wantRows)
	}
}

func TestPrepareBranchScopedCSVDataRejectsBlankBranch(t *testing.T) {
	content := stringsJoinLines(
		`"Transaction ID"|"Amount"`,
		`"TX-01"|"1500"`,
	)

	if _, err := prepareBranchScopedCSVData(content, "   "); err == nil {
		t.Fatal("expected blank branch to be rejected")
	}
}

func TestHashCSVRowIncludesAsOfDate(t *testing.T) {
	row := []string{"1001", "Cash on Hand", "123456.78"}

	hashA := hashCSVRow("Balance Sheet Report csv", "2026-03-11", row)
	hashB := hashCSVRow("Balance Sheet Report csv", "2026-03-12", row)
	hashC := hashCSVRow("Balance Sheet Report csv", "2026-03-11", row)

	if hashA == hashB {
		t.Fatalf("hash should differ across as_of_date: got %q and %q", hashA, hashB)
	}
	if hashA != hashC {
		t.Fatalf("hash should be stable for same source/date/row: got %q and %q", hashA, hashC)
	}
}

func TestHashCSVRowIncludesBranchValue(t *testing.T) {
	rowA := []string{"000", "1001", "Cash on Hand", "123456.78"}
	rowB := []string{"001", "1001", "Cash on Hand", "123456.78"}

	hashA := hashCSVRow("Balance Sheet Report csv", "2026-03-11", rowA)
	hashB := hashCSVRow("Balance Sheet Report csv", "2026-03-11", rowB)

	if hashA == hashB {
		t.Fatalf("hash should differ across branch values: got %q and %q", hashA, hashB)
	}
}

func TestPrepareBranchScopedCSVDataMakesHashBranchAware(t *testing.T) {
	content := stringsJoinLines(
		`"Transaction ID"|"Amount"`,
		`"TX-01"|"1500"`,
	)

	preparedA, err := prepareBranchScopedCSVData(content, "000")
	if err != nil {
		t.Fatalf("prepareBranchScopedCSVData failed: %v", err)
	}
	preparedB, err := prepareBranchScopedCSVData(content, "001")
	if err != nil {
		t.Fatalf("prepareBranchScopedCSVData failed: %v", err)
	}

	hashA := hashCSVRow("Vault Mutation Report csv", "2026-03-11", preparedA.rows[0])
	hashB := hashCSVRow("Vault Mutation Report csv", "2026-03-11", preparedB.rows[0])
	if hashA == hashB {
		t.Fatalf("hash should differ across injected _branch values: got %q and %q", hashA, hashB)
	}
}

func TestUpsertBranchScopedCSVRejectsBlankBranch(t *testing.T) {
	store := &Store{}
	content := stringsJoinLines(
		`"Transaction ID"|"Amount"`,
		`"TX-01"|"1500"`,
	)

	if _, err := store.UpsertBranchScopedCSV(context.Background(), "vault_mutations", "Vault Mutation Report csv", "2026-03-11", "", content); err == nil {
		t.Fatal("expected blank branch to be rejected")
	}
}

func stringsJoinLines(lines ...string) string {
	if len(lines) == 0 {
		return ""
	}
	joined := lines[0]
	for _, line := range lines[1:] {
		joined += "\n" + line
	}
	return joined
}
