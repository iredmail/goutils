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

func TestLikeDialects(t *testing.T) {
	tests := []struct {
		name        string
		dialect     string
		goquDialect string
		expectedSQL string
	}{
		{name: "mysql", dialect: "mysql", goquDialect: "mysql", expectedSQL: "`email` LIKE BINARY ? ESCAPE '!'"},
		{name: "mariadb", dialect: "mariadb", goquDialect: "mysql", expectedSQL: "`email` LIKE BINARY ? ESCAPE '!'"},
		{name: "postgres", dialect: "postgres", goquDialect: "postgres", expectedSQL: `"email" LIKE $1 ESCAPE '!'`},
		{name: "pgsql", dialect: "pgsql", goquDialect: "postgres", expectedSQL: `"email" LIKE $1 ESCAPE '!'`},
		{name: "sqlite", dialect: "sqlite", goquDialect: "sqlite3", expectedSQL: "`email` LIKE ? ESCAPE '!'"},
		{name: "sqlite3", dialect: "sqlite3", goquDialect: "sqlite3", expectedSQL: "`email` LIKE ? ESCAPE '!'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expression, err := sqlutils.Like(
				tt.dialect,
				goqu.I("email"),
				"@foo_bar.com",
				sqlutils.LikeOptions{Match: sqlutils.LikeMatchSuffix},
			)
			require.NoError(t, err)

			query, args, err := goqu.Dialect(tt.goquDialect).
				From("accounts").
				Where(expression).
				Prepared(true).
				ToSQL()
			require.NoError(t, err)
			assert.Contains(t, query, tt.expectedSQL)
			assert.Equal(t, []interface{}{`%@foo!_bar.com`}, args)
		})
	}
}

func TestLikeMatchModes(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		options  sqlutils.LikeOptions
		expected string
	}{
		{
			name:     "exact escapes literal characters",
			value:    `foo_bar%baz!*`,
			options:  sqlutils.LikeOptions{Match: sqlutils.LikeMatchExact},
			expected: `foo!_bar!%baz!!*`,
		},
		{
			name:     "prefix",
			value:    "foo_bar",
			options:  sqlutils.LikeOptions{Match: sqlutils.LikeMatchPrefix},
			expected: `foo!_bar%`,
		},
		{
			name:     "suffix",
			value:    "foo_bar",
			options:  sqlutils.LikeOptions{Match: sqlutils.LikeMatchSuffix},
			expected: `%foo!_bar`,
		},
		{
			name:     "contains",
			value:    "foo_bar",
			options:  sqlutils.LikeOptions{Match: sqlutils.LikeMatchContains},
			expected: `%foo!_bar%`,
		},
		{
			name:  "star wildcard with literal percent",
			value: `foo_bar*100%*`,
			options: sqlutils.LikeOptions{
				Match:          sqlutils.LikeMatchExact,
				StarAsWildcard: true,
			},
			expected: `foo!_bar%100!%%`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expression, err := sqlutils.Like("postgres", goqu.I("email"), tt.value, tt.options)
			require.NoError(t, err)

			_, args, err := goqu.Dialect("postgres").
				From("accounts").
				Where(expression).
				Prepared(true).
				ToSQL()
			require.NoError(t, err)
			assert.Equal(t, []interface{}{tt.expected}, args)
		})
	}
}

func TestLikeCaseInsensitive(t *testing.T) {
	expression, err := sqlutils.Like(
		"postgres",
		goqu.I("email"),
		"@Foo_Bar.COM",
		sqlutils.LikeOptions{Match: sqlutils.LikeMatchSuffix, CaseInsensitive: true},
	)
	require.NoError(t, err)

	query, args, err := goqu.Dialect("postgres").
		From("accounts").
		Where(expression).
		Prepared(true).
		ToSQL()
	require.NoError(t, err)
	assert.Contains(t, query, `LOWER("email") LIKE $1 ESCAPE '!'`)
	assert.Equal(t, []interface{}{`%@foo!_bar.com`}, args)
}

func TestLikePostgreSQLByteaColumn(t *testing.T) {
	column := goqu.Func("convert_from", goqu.I("email"), "UTF8")
	expression, err := sqlutils.Like(
		"pgsql",
		column,
		"@foo_bar.com",
		sqlutils.LikeOptions{Match: sqlutils.LikeMatchSuffix},
	)
	require.NoError(t, err)

	query, args, err := goqu.Dialect("postgres").
		From("accounts").
		Where(expression).
		Prepared(true).
		ToSQL()
	require.NoError(t, err)
	assert.Contains(t, query, `convert_from("email", $1) LIKE $2 ESCAPE '!'`)
	assert.Equal(t, []interface{}{"UTF8", `%@foo!_bar.com`}, args)
}

func TestLikeErrors(t *testing.T) {
	tests := []struct {
		name          string
		dialect       string
		column        goqu.Expression
		options       sqlutils.LikeOptions
		expectedError string
	}{
		{
			name:          "unsupported dialect",
			dialect:       "sqlserver",
			column:        goqu.I("email"),
			expectedError: `unsupported dialect "sqlserver"`,
		},
		{
			name:          "unsupported match mode",
			dialect:       "mysql",
			column:        goqu.I("email"),
			options:       sqlutils.LikeOptions{Match: sqlutils.LikeMatchMode(255)},
			expectedError: "unsupported match mode 255",
		},
		{
			name:          "missing column",
			dialect:       "mysql",
			expectedError: "column is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expression, err := sqlutils.Like(tt.dialect, tt.column, "value", tt.options)
			require.Error(t, err)
			assert.Nil(t, expression)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestLikeWithSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	gdb := goqu.New("sqlite3", db)
	_, err = gdb.Exec("CREATE TABLE samples (value TEXT NOT NULL)")
	require.NoError(t, err)
	_, err = gdb.Exec(
		"INSERT INTO samples (value) VALUES (?), (?), (?), (?), (?), (?)",
		"user@foo_bar.com",
		"user@fooXbar.com",
		"50%",
		"500",
		"a*b",
		"axxb",
	)
	require.NoError(t, err)

	tests := []struct {
		name     string
		value    string
		options  sqlutils.LikeOptions
		expected []string
	}{
		{
			name:     "literal underscore",
			value:    "@foo_bar.com",
			options:  sqlutils.LikeOptions{Match: sqlutils.LikeMatchSuffix},
			expected: []string{"user@foo_bar.com"},
		},
		{
			name:     "literal percent",
			value:    "50%",
			options:  sqlutils.LikeOptions{Match: sqlutils.LikeMatchContains},
			expected: []string{"50%"},
		},
		{
			name:     "literal star",
			value:    "a*b",
			options:  sqlutils.LikeOptions{Match: sqlutils.LikeMatchExact},
			expected: []string{"a*b"},
		},
		{
			name:  "star wildcard",
			value: "a*b",
			options: sqlutils.LikeOptions{
				Match:          sqlutils.LikeMatchExact,
				StarAsWildcard: true,
			},
			expected: []string{"a*b", "axxb"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expression, err := sqlutils.Like("sqlite3", goqu.I("value"), tt.value, tt.options)
			require.NoError(t, err)

			var values []string
			err = gdb.From("samples").Select("value").Where(expression).ScanVals(&values)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, values)
		})
	}
}
