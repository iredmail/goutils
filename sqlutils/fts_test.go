package sqlutils

import (
	"os"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

type testDoc struct {
	ID      int64  `db:"id"`
	Title   string `db:"title"`
	Content string `db:"content"`
}

func setupFTSDB(t *testing.T) (*goqu.Database, func()) {
	dbFile, err := os.CreateTemp("", "goutils_fts_test_*.db")
	assert.Nil(t, err)
	_ = dbFile.Close()

	sqliteDB, err := InitSQLiteDB(dbFile.Name(), nil, 0, 0)
	assert.Nil(t, err)

	gdb := goqu.New("sqlite", sqliteDB)

	_, err = gdb.Exec(`
		CREATE TABLE docs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL
		);

		CREATE VIRTUAL TABLE IF NOT EXISTS fts_docs USING fts5 (
            title,
			content,
			content='docs',
			content_rowid='id',
			tokenize="trigram case_sensitive 0"
		);

		-- 新建时需要插入记录到 fts 表。
		CREATE TRIGGER IF NOT EXISTS fts_docs_auto_insert
			AFTER INSERT ON docs
			BEGIN
			INSERT INTO fts_docs(rowid, title, content)
                         VALUES (new.id, new.title, new.content);
			END;

		-- 更新时需要同步更新 fts 表里的记录。
		CREATE TRIGGER IF NOT EXISTS fts_docs_auto_update
			AFTER UPDATE ON docs
			BEGIN
			INSERT INTO fts_docs(fts_docs, rowid, title, content)
                          VALUES('delete', old.id, old.title, old.content);
			INSERT INTO fts_docs(rowid, title, content)
                         VALUES (new.id, new.title, new.content);
			END;

		-- 移除时同步移除 fts 表里的记录。
		CREATE TRIGGER IF NOT EXISTS fts_docs_auto_delete
			AFTER DELETE ON docs
			BEGIN
			INSERT INTO fts_docs(fts_docs, rowid, title, content)
                          VALUES('delete', old.id, old.title, old.content);
			END;
	`)
	assert.Nil(t, err)

	testData := []testDoc{
		{ID: 1, Title: "Hello World", Content: "This is a test document with normal content about programming."},
		{ID: 2, Title: "Special Characters", Content: `Document content has 'single' and "double" quotes; plus semicolons; in text.`},
		{ID: 3, Title: "Email Address", Content: "Reach us via email at test@example.com for further information details."},
		{ID: 4, Title: "Fourth Topic", Content: "Discussion about database queries and search performance optimization methods."},
	}

	for _, doc := range testData {
		_, e := gdb.
			Insert("docs").
			Cols("title", "content").
			Vals(goqu.Vals{doc.Title, doc.Content}).
			Executor().Exec()

		assert.Nil(t, e)
	}

	cleanup := func() {
		_ = sqliteDB.Close()
		_ = os.Remove(dbFile.Name())
	}

	return gdb, cleanup
}

func TestFTSSearch_NormalKeyword(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	var results []testDoc
	err := FTSSearch(gdb, "Hello", "fts_docs", "docs", &results)
	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Hello World", results[0].Title)
}

func TestFTSSearch_WithSingleQuote(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	var results []testDoc
	err := FTSSearch(gdb, `'single'`, "fts_docs", "docs", &results)
	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Special Characters", results[0].Title)
}

func TestFTSSearch_WithDoubleQuote(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	var results []testDoc
	err := FTSSearch(gdb, `"double"`, "fts_docs", "docs", &results)
	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Special Characters", results[0].Title)
}

func TestFTSSearch_WithSemicolon(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	var results []testDoc
	err := FTSSearch(gdb, `semicolons;`, "fts_docs", "docs", &results)
	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Special Characters", results[0].Title)
}

func TestFTSSearch_EmailWithAtSign(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	var results []testDoc
	err := FTSSearch(gdb, `test@example.com`, "fts_docs", "docs", &results)
	assert.Nil(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, "Email Address", results[0].Title)
}

func TestFTSSearch_ShortEmailAddress(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	_, err := gdb.
		Insert("docs").
		Cols("title", "content").
		Vals(goqu.Vals{"Short Email", "Please reach us at u@x.io for support."}).
		Executor().Exec()
	assert.Nil(t, err)

	var results []testDoc
	err = FTSSearch(gdb, `u@x.io`, "fts_docs", "docs", &results)
	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Short Email", results[0].Title)
}

func TestFTSSearch_SQLInjectionAttempt_OrClause(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	var results []testDoc
	err := FTSSearch(gdb, `nothing' OR '1'='1`, "fts_docs", "docs", &results)
	assert.Nil(t, err)
	assert.Len(t, results, 0)
}

func TestFTSSearch_SQLInjectionAttempt_DropTable(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	var results []testDoc
	err := FTSSearch(gdb, `x"; DROP TABLE docs; --`, "fts_docs", "docs", &results)
	assert.Nil(t, err)
	assert.Len(t, results, 0)

	var count int64
	_, err = gdb.From("docs").Select(goqu.COUNT("*")).ScanVal(&count)
	assert.Nil(t, err)
	assert.Equal(t, int64(4), count, "docs table should still exist with all 4 rows")

	_, err = gdb.From("fts_docs").Select(goqu.COUNT("*")).ScanVal(&count)
	assert.Nil(t, err)
	assert.Equal(t, int64(4), count, "fts_docs table should still exist")
}

func TestFTSSearch_SQLInjectionAttempt_Union(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	var results []testDoc
	err := FTSSearch(gdb, `" UNION SELECT id, title, content FROM docs WHERE '1'='1`, "fts_docs", "docs", &results)
	assert.Nil(t, err)
	assert.Len(t, results, 0)
}

func TestFTSSearch_SQLInjectionAttempt_AllSpecialChars(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	attacks := []string{
		`'; DROP TABLE docs;--`,
		`" OR 1=1 --`,
		`'; EXEC('DROP TABLE docs');--`,
		`test' AND (SELECT COUNT(*) FROM docs) > 0 AND '1'='1`,
		`" OR "a"="a`,
		`x' WHERE rowid > 0; --`,
	}

	for _, attack := range attacks {
		var results []testDoc
		err := FTSSearch(gdb, attack, "fts_docs", "docs", &results)
		assert.Nil(t, err, "attack payload: %q", attack)
		assert.Len(t, results, 0, "expected no results for injection payload: %q", attack)
	}
}

func TestFTSSearch_EmptyKeyword(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	var results []testDoc
	err := FTSSearch(gdb, "", "fts_docs", "docs", &results)
	assert.Nil(t, err)
	assert.Len(t, results, 0)
}

func TestFTSSearch_NoMatch(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	var results []testDoc
	err := FTSSearch(gdb, "nonexistentkeyword12345", "fts_docs", "docs", &results)
	assert.Nil(t, err)
	assert.Len(t, results, 0)
}

func TestFTSSearch_KeywordAsTerm(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	_, e := gdb.
		Insert("docs").
		Cols("title", "content").
		Vals(goqu.Vals{"AND and OR Search", "This document uses AND and OR as normal words inside content."}).
		Executor().Exec()
	assert.Nil(t, e)

	var results []testDoc
	err := FTSSearch(gdb, "AND and OR", "fts_docs", "docs", &results)
	assert.Nil(t, err)
	assert.NotEmpty(t, results, "expected at least 1 result when searching 'AND and OR' as regular terms")
	found := false
	for _, r := range results {
		if r.Title == "AND and OR Search" {
			found = true
			break
		}
	}
	assert.True(t, found, "should match newly inserted doc with AND/OR words")
}

func TestFTSSearch_BooleanKeywordNotOperator(t *testing.T) {
	gdb, cleanup := setupFTSDB(t)
	defer cleanup()

	_, e := gdb.
		Insert("docs").
		Cols("title", "content").
		Vals(goqu.Vals{"GoodDoc", "ok and good with normal words"}).
		Executor().Exec()
	assert.Nil(t, e)
	_, e = gdb.
		Insert("docs").
		Cols("title", "content").
		Vals(goqu.Vals{"BadDoc", "nothing matches here completely different"}).
		Executor().Exec()
	assert.Nil(t, e)

	var results []testDoc
	err := FTSSearch(gdb, "ok and good", "fts_docs", "docs", &results)
	assert.Nil(t, err)
	found := false
	for _, r := range results {
		if r.Title == "GoodDoc" {
			found = true
			break
		}
	}
	assert.True(t, found, "'ok and good' should match doc containing 'ok and good'")
	for _, r := range results {
		assert.NotEqual(t, "BadDoc", r.Title, "should not match unrelated doc")
	}
}

func TestEscapeFTSKeyword(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`simple`, `simple`},
		{`a*b`, `"a*b"`},
		{`foo_bar`, `foo_bar`},
		{`中文测试`, `中文测试`},
		{`with spaces`, `with spaces`},
		{`with "double" quotes`, `with """double""" quotes`},
		{`'single quotes'`, `"'single" "quotes'"`},
		{`semi;colon`, `"semi;colon"`},
		{`email@test.com`, `"email@test.com"`},
		{`u@x.io`, `"u@x.io"`},
		{`hello AND world`, `hello "AND" world`},
		{`hello OR world`, `hello "OR" world`},
		{`hello NOT world`, `hello "NOT" world`},
		{`hello NEAR world`, `hello NEAR world`},
		{`hello and world`, `hello and world`},
		{`hello or world`, `hello or world`},
		{`hello not world`, `hello not world`},
		{`hello near world`, `hello near world`},
		{`NOT hello`, `"NOT" hello`},
		{`AND`, `"AND"`},
		{`OR`, `"OR"`},
		{`NOT`, `"NOT"`},
		{`NEAR`, `NEAR`},
		{`and`, `and`},
		{`or`, `or`},
		{`not`, `not`},
		{`nearby`, `nearby`},
		{`andrew`, `andrew`},
		{`notion`, `notion`},
		{`andrew or notion nearby`, `andrew or notion nearby`},
		{`AND AND AND`, `"AND" "AND" "AND"`},
		{``, ``},
	}

	for _, tt := range tests {
		result := escapeFTSKeyword(tt.input)
		assert.Equal(t, tt.expected, result, "input: %q", tt.input)
	}
}
