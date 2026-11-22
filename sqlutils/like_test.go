package sqlutils_test

import (
	"database/sql"
	"testing"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	"github.com/iredmail/goutils/sqlutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestEscapeLikeLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty", input: "", expected: ""},
		{name: "plain", input: "plain", expected: "plain"},
		{name: "underscore", input: "foo_bar", expected: "foo!_bar"},
		{name: "percent", input: "50%", expected: "50!%"},
		{name: "escape character", input: "a!b", expected: "a!!b"},
		{name: "combined", input: "a!_%b", expected: "a!!!_!%b"},
		{name: "asterisk is literal in LIKE", input: "a*b", expected: "a*b"},
		{name: "unicode", input: "用户_名", expected: "用户!_名"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sqlutils.EscapeLikeLiteral(tt.input))
		})
	}
}

func TestEscapeLikeLiteralWithSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	tests := []struct {
		name        string
		literal     string
		matching    string
		nonMatching string
	}{
		{
			name:        "underscore does not match another character",
			literal:     "foo_bar.com",
			matching:    "user@foo_bar.com",
			nonMatching: "user@fooXbar.com",
		},
		{
			name:        "percent does not match a character sequence",
			literal:     "foo%bar.com",
			matching:    "user@foo%bar.com",
			nonMatching: "user@fooXbar.com",
		},
		{
			name:        "escape character remains literal",
			literal:     "foo!bar.com",
			matching:    "user@foo!bar.com",
			nonMatching: "user@foobar.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := "%@" + sqlutils.EscapeLikeLiteral(tt.literal)
			for _, matchTest := range []struct {
				value    string
				expected bool
			}{
				{value: tt.matching, expected: true},
				{value: tt.nonMatching, expected: false},
			} {
				var matched bool
				err := db.QueryRowContext(
					t.Context(),
					"SELECT ? LIKE ? ESCAPE '!'",
					matchTest.value,
					pattern,
				).Scan(&matched)
				require.NoError(t, err)
				assert.Equal(t, matchTest.expected, matched, "value: %q, pattern: %q", matchTest.value, pattern)
			}
		})
	}
}

func TestEscapeLikeLiteralGoquDialects(t *testing.T) {
	pattern := "%@" + sqlutils.EscapeLikeLiteral("foo_bar.com")
	tests := []struct {
		dialect     string
		likeClause  string
		expectedSQL string
	}{
		{
			dialect:     "mysql",
			likeClause:  "? LIKE BINARY ? ESCAPE '!'",
			expectedSQL: "`email` LIKE BINARY ? ESCAPE '!'",
		},
		{
			dialect:     "postgres",
			likeClause:  "? LIKE ? ESCAPE '!'",
			expectedSQL: `"email" LIKE $1 ESCAPE '!'`,
		},
		{
			dialect:     "sqlite3",
			likeClause:  "? LIKE ? ESCAPE '!'",
			expectedSQL: "`email` LIKE ? ESCAPE '!'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			query, args, err := goqu.Dialect(tt.dialect).
				From("accounts").
				Where(goqu.L(tt.likeClause, goqu.I("email"), pattern)).
				Prepared(true).
				ToSQL()
			require.NoError(t, err)
			assert.Contains(t, query, tt.expectedSQL)
			assert.Equal(t, []interface{}{pattern}, args)
		})
	}
}
