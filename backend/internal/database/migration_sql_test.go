package database

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplitSQLStatementsHonorsPostgresQuotingAndComments(t *testing.T) {
	source := `-- semicolons in comments are ignored;
CREATE TABLE "semi;colon" (value TEXT DEFAULT 'one;two', note TEXT DEFAULT 'it''s fine');
/* nested comment ; /* inner ; */ outer ; */
SELECT 9;
DO $audit_body$
BEGIN
    PERFORM 'inside;function';
    PERFORM "semi;identifier";
END;
$audit_body$;
INSERT INTO example(value) VALUES (E'escaped\';semicolon');
`
	got, err := splitSQLStatements(source)
	if err != nil {
		t.Fatalf("splitSQLStatements: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("got %d statements, want 4: %#v", len(got), got)
	}
	if !strings.Contains(got[0], `"semi;colon"`) || !strings.Contains(got[0], "'one;two'") {
		t.Fatalf("first statement lost quoted content: %q", got[0])
	}
	if !strings.Contains(got[2], "PERFORM 'inside;function';") || !strings.Contains(got[2], "END;") {
		t.Fatalf("dollar-quoted function body was split: %q", got[2])
	}
	if !strings.Contains(got[3], `E'escaped\';semicolon'`) {
		t.Fatalf("escaped string literal was not preserved: %q", got[3])
	}
}

func TestSplitSQLStatementsAcceptsEmptyAndTrailingComments(t *testing.T) {
	for _, source := range []string{"", "  ; ;  ", "-- comment with ; and no newline"} {
		if _, err := splitSQLStatements(source); err != nil {
			t.Fatalf("splitSQLStatements(%q): %v", source, err)
		}
	}
}

func TestSplitSQLStatementsRejectsUnterminatedQuotesAndComments(t *testing.T) {
	for _, source := range []string{"SELECT 'unterminated", `SELECT "unterminated`, "DO $body$ BEGIN;", "/* unterminated comment"} {
		if _, err := splitSQLStatements(source); err == nil {
			t.Errorf("splitSQLStatements(%q) succeeded, want error", source)
		}
	}
}

func TestSQLDollarQuoteDelimiter(t *testing.T) {
	tests := []struct {
		source string
		want   string
		ok     bool
	}{
		{source: "$$body$$", want: "$$", ok: true},
		{source: "$tag_1$body$tag_1$", want: "$tag_1$", ok: true},
		{source: "$1$not-a-tag$", ok: false},
		{source: "$unterminated", ok: false},
	}
	for _, tt := range tests {
		got, ok := sqlDollarQuoteDelimiter(tt.source, 0)
		if got != tt.want || ok != tt.ok {
			t.Errorf("sqlDollarQuoteDelimiter(%q) = (%q, %t), want (%q, %t)", tt.source, got, ok, tt.want, tt.ok)
		}
	}
}

func TestSplitSQLStatementsPreservesStatementOrder(t *testing.T) {
	got, err := splitSQLStatements("SELECT 1; SELECT 2; SELECT 3")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"SELECT 1", "SELECT 2", "SELECT 3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
