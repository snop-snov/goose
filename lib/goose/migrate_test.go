package goose

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getSqlite3Driver(t *testing.T) (DBDriver, string) {
	dbFile := filepath.Join(os.TempDir(), "goose.db")
	return DBDriver{
		Name:    "sqlite3",
		Import:  "github.com/mattn/go-sqlite3",
		Dialect: Sqlite3Dialect{},
		OpenStr: dbFile,
	}, dbFile
}

func getMysqlDriver(t *testing.T) DBDriver {
	dsn := os.Getenv("MYSQL_DATABASE_DSN")
	if dsn == "" {
		t.SkipNow()
	}
	return DBDriver{
		Name:    "mysql",
		Import:  "github.com/go-sql-driver/mysql",
		Dialect: MySqlDialect{},
		OpenStr: dsn,
	}
}

func getPostgresDriver(t *testing.T) DBDriver {
	dsn := os.Getenv("POSTGRES_DATABASE_DSN")
	if dsn == "" {
		t.SkipNow()
	}
	return DBDriver{
		Name:    "postgres",
		Import:  "github.com/lib/pq",
		Dialect: PostgresDialect{},
		OpenStr: dsn,
	}
}

func getRedshiftDriver(t *testing.T) DBDriver {
	dsn := os.Getenv("REDSHIFT_DATABASE_DSN")
	if dsn == "" {
		t.SkipNow()
	}
	return DBDriver{
		Name:    "postgres",
		Dialect: RedshiftDialect{},
		OpenStr: dsn,
	}
}

func TestMigrationSorterLen(t *testing.T) {
	ms := migrationSorter{
		{Version: 1},
		{Version: 2},
		{Version: 4},
		{Version: 3},
	}
	l := ms.Len()
	if l != 4 {
		t.Errorf("expected ms.Len() == 4, but got %d\n", l)
	}
}

func TestMigrationSorterSwap(t *testing.T) {
	ms := migrationSorter{
		{Version: 1},
		{Version: 2},
		{Version: 4},
		{Version: 3},
	}
	ms.Swap(1, 2)
	if ms[1].Version != 4 {
		t.Errorf("expected ms[1].Version == 4, but got %d\n", ms[1].Version)
	}
	if ms[2].Version != 2 {
		t.Errorf("expected ms[2].Version == 2, but got %d\n", ms[1].Version)
	}
}

func TestMigrationSorterLess(t *testing.T) {
	ms := migrationSorter{
		{Version: 1},
		{Version: 2},
		{Version: 4},
		{Version: 3},
	}
	v := ms.Less(2, 3)
	if v != false {
		t.Errorf("expected ms.Less(2,3) == false, but got %v\n", v)
	}
	v = ms.Less(3, 2)
	if v != true {
		t.Errorf("expected ms.Less(3,2) == true, but got %v\n", v)
	}
}

func setupMigrationsDir(migrationMap map[string][2]string) (string, func()) {
	td, err := ioutil.TempDir("", "goose-test-")
	if err != nil {
		panic(err)
	}

	dbPath := filepath.Join(td, "db")
	migrationsPath := filepath.Join(dbPath, "migrations")
	os.MkdirAll(migrationsPath, 0700)

	for name, migrations := range migrationMap {
		migStr := sqlMigrationStr(migrations[0], migrations[1])
		if strings.HasSuffix(name, ".go") {
			n := strings.Index(name, "_")
			migStr = goMigrationStr(name[:n], migrations[0], migrations[1])
		}

		if err := ioutil.WriteFile(filepath.Join(migrationsPath, name), []byte(migStr), 0600); err != nil {
			panic(err)
		}
	}

	return migrationsPath, func() { os.RemoveAll(td) }
}

func sqlMigrationStr(up, down string) string {
	return `-- +goose Up
` + up + `

-- +goose Down
` + down + `
`
}

func goMigrationStr(name, up, down string) string {
	return `package main

import (
	"database/sql"
)

// Up is executed when this migration is applied
func Up_` + name + `(txn *sql.Tx) {
	` + up + `
}

// Down is executed when this migration is rolled back
func Down_` + name + `(txn *sql.Tx) {
	` + down + `
}
`
}

func TestCollectMigrations(t *testing.T) {
	md, mdCleanup := setupMigrationsDir(map[string][2]string{
		"20010203040506_first.sql":  [2]string{"SELECT 1;", "SELECT 1;"},
		"20010203040507_second.sql": [2]string{"SELECT 2;", "SELECT 2;"},
		"20010203040508_third.sql":  [2]string{"SELECT 3;", "SELECT 3;"},
		"20160318040509_fourth.go":  [2]string{`println("hi")`, `println("bye")`},
	})
	defer mdCleanup()

	migs, err := CollectMigrations(md)
	require.NoError(t, err)

	assert.Len(t, migs, 4)
	assert.Contains(t, migs, &Migration{
		Version:   20010203040506,
		IsApplied: false,
		Source:    filepath.Join(md, "20010203040506_first.sql"),
	})
	assert.Contains(t, migs, &Migration{
		Version:   20010203040507,
		IsApplied: false,
		Source:    filepath.Join(md, "20010203040507_second.sql"),
	})
	assert.Contains(t, migs, &Migration{
		Version:   20010203040508,
		IsApplied: false,
		Source:    filepath.Join(md, "20010203040508_third.sql"),
	})
	assert.Contains(t, migs, &Migration{
		Version:   20160318040509,
		IsApplied: false,
		Source:    filepath.Join(md, "20160318040509_fourth.go"),
	})
}

func testRunMigrationsOnDb(t *testing.T, driver DBDriver) {
	md, mdCleanup := setupMigrationsDir(map[string][2]string{
		"20010203040506_setup.sql": [2]string{"CREATE TABLE test(value VARCHAR(20));", "DROP TABLE test;"},
		"20010203040507_one.sql":   [2]string{"INSERT INTO test(value) VALUES('one');", "DELETE FROM test WHERE value = 'one';"},
		"20010203040508_two.sql":   [2]string{"INSERT INTO test(value) VALUES('two');", "DELETE FROM test WHERE value = 'two';"},
		"20160318040509_fourth.go": [2]string{`println("hi")`, `println("bye")`},
	})
	defer mdCleanup()
	conf := &DBConf{
		Driver:        driver,
		MigrationsDir: md,
	}

	db, err := OpenDBFromDBConf(conf)
	require.NoError(t, err)

	db.Exec("DROP TABLE goose_db_version")
	db.Exec("DROP TABLE test")

	err = RunMigrationsOnDb(conf, conf.MigrationsDir, 20160318040509, db)
	require.NoError(t, err)

	rows, err := db.Query("SELECT value FROM test")
	require.NoError(t, err)
	defer rows.Close()
	var values []string
	for rows.Next() {
		var value string
		err := rows.Scan(&value)
		require.NoError(t, err)
		values = append(values, value)
	}

	assert.Len(t, values, 2)
	assert.Contains(t, values, "one")
	assert.Contains(t, values, "two")
}
func TestRunMigrationsOnDb_sqlite3(t *testing.T) {
	config, dbFile := getSqlite3Driver(t)
	defer os.Remove(dbFile)
	testRunMigrationsOnDb(t, config)
}
func TestRunMigrationsOnDb_mysql(t *testing.T) {
	testRunMigrationsOnDb(t, getMysqlDriver(t))
}
func TestRunMigrationsOnDb_postgres(t *testing.T) {
	testRunMigrationsOnDb(t, getPostgresDriver(t))
}
func TestRunMigrationsOnDb_redshift(t *testing.T) {
	testRunMigrationsOnDb(t, getRedshiftDriver(t))
}

func testRunMigrationsOnDb_missingMiddle(t *testing.T, driver DBDriver) {
	md, mdCleanup := setupMigrationsDir(map[string][2]string{
		"20010203040506_setup.sql": [2]string{"CREATE TABLE test(value VARCHAR(20));", "DROP TABLE test;"},
		"20010203040507_one.sql":   [2]string{"INSERT INTO test(value) VALUES('one');", "DELETE FROM test WHERE value = 'one';"},
		"20010203040508_two.sql":   [2]string{"INSERT INTO test(value) VALUES('two');", "DELETE FROM test WHERE value = 'two';"},
		"20160318040509_fourth.go": [2]string{`println("hi")`, `println("bye")`},
	})
	defer mdCleanup()
	conf := &DBConf{
		Driver:        driver,
		MigrationsDir: md,
	}

	db, err := OpenDBFromDBConf(conf)
	require.NoError(t, err)

	db.Exec("DROP TABLE goose_db_version")
	db.Exec("DROP TABLE test")

	// make the middle migration disappear for a moment
	err = os.Rename(filepath.Join(md, "20010203040507_one.sql"), filepath.Join(md, "20010203040507_one.sql_"))
	require.NoError(t, err)

	err = RunMigrationsOnDb(conf, conf.MigrationsDir, 20160318040509, db)
	require.NoError(t, err)

	rows, err := db.Query("SELECT value FROM test")
	require.NoError(t, err)
	defer rows.Close()
	var values []string
	for rows.Next() {
		var value string
		err := rows.Scan(&value)
		require.NoError(t, err)
		values = append(values, value)
	}

	assert.Len(t, values, 1)
	assert.Contains(t, values, "two")

	// now put it back
	err = os.Rename(filepath.Join(md, "20010203040507_one.sql_"), filepath.Join(md, "20010203040507_one.sql"))
	require.NoError(t, err)

	err = RunMigrationsOnDb(conf, conf.MigrationsDir, 20160318040509, db)
	require.NoError(t, err)

	rows, err = db.Query("SELECT value FROM test")
	require.NoError(t, err)
	defer rows.Close()
	values = []string{}
	for rows.Next() {
		var value string
		err := rows.Scan(&value)
		require.NoError(t, err)
		values = append(values, value)
	}

	assert.Len(t, values, 2)
	assert.Contains(t, values, "one")
	assert.Contains(t, values, "two")
}
func TestRunMigrationsOnDb_missingMiddle_sqlite3(t *testing.T) {
	config, dbFile := getSqlite3Driver(t)
	defer os.Remove(dbFile)
	testRunMigrationsOnDb_missingMiddle(t, config)
}
func TestRunMigrationsOnDb_missingMiddle_mysql(t *testing.T) {
	testRunMigrationsOnDb_missingMiddle(t, getMysqlDriver(t))
}
func TestRunMigrationsOnDb_missingMiddle_postgres(t *testing.T) {
	testRunMigrationsOnDb_missingMiddle(t, getPostgresDriver(t))
}
func TestRunMigrationsOnDb_missingMiddle_redshift(t *testing.T) {
	testRunMigrationsOnDb_missingMiddle(t, getRedshiftDriver(t))
}

func testRunMigrationsOnDb_down(t *testing.T, driver DBDriver) {
	md, mdCleanup := setupMigrationsDir(map[string][2]string{
		"20010203040506_setup.sql": [2]string{"CREATE TABLE test(value VARCHAR(20));", "DROP TABLE test;"},
		"20010203040507_one.sql":   [2]string{"INSERT INTO test(value) VALUES('one');", "DELETE FROM test WHERE value = 'one';"},
		"20010203040508_two.sql":   [2]string{"INSERT INTO test(value) VALUES('two');", "DELETE FROM test WHERE value = 'two';"},
		"20160318040509_fourth.go": [2]string{`println("hi")`, `println("bye")`},
	})
	defer mdCleanup()
	conf := &DBConf{
		Driver:        driver,
		MigrationsDir: md,
	}

	db, err := OpenDBFromDBConf(conf)
	require.NoError(t, err)

	db.Exec("DROP TABLE goose_db_version")
	db.Exec("DROP TABLE test")

	// up
	err = RunMigrationsOnDb(conf, conf.MigrationsDir, 20160318040509, db)
	require.NoError(t, err)

	// down
	err = RunMigrationsOnDb(conf, conf.MigrationsDir, 20010203040507, db)
	require.NoError(t, err)

	rows, err := db.Query("SELECT value FROM test")
	require.NoError(t, err)
	defer rows.Close()
	var values []string
	for rows.Next() {
		var value string
		err := rows.Scan(&value)
		require.NoError(t, err)
		values = append(values, value)
	}

	assert.Len(t, values, 1)
	assert.Contains(t, values, "one")
}
func TestRunMigrationsOnDb_down_sqlite3(t *testing.T) {
	config, dbFile := getSqlite3Driver(t)
	defer os.Remove(dbFile)
	testRunMigrationsOnDb_down(t, config)
}
func TestRunMigrationsOnDb_down_mysql(t *testing.T) {
	testRunMigrationsOnDb_down(t, getMysqlDriver(t))
}
func TestRunMigrationsOnDb_down_postgres(t *testing.T) {
	testRunMigrationsOnDb_down(t, getPostgresDriver(t))
}
func TestRunMigrationsOnDb_down_redshift(t *testing.T) {
	testRunMigrationsOnDb_down(t, getRedshiftDriver(t))
}

func testRunMigrationsOnDb_upDownUp(t *testing.T, driver DBDriver) {
	md, mdCleanup := setupMigrationsDir(map[string][2]string{
		"20010203040506_setup.sql": [2]string{"CREATE TABLE test(value VARCHAR(20));", "DROP TABLE test;"},
		"20010203040507_one.sql":   [2]string{"INSERT INTO test(value) VALUES('one');", "DELETE FROM test WHERE value = 'one';"},
		"20010203040508_two.sql":   [2]string{"INSERT INTO test(value) VALUES('two');", "DELETE FROM test WHERE value = 'two';"},
		"20160318040509_fourth.go": [2]string{`println("hi")`, `println("bye")`},
	})
	defer mdCleanup()
	conf := &DBConf{
		Driver:        driver,
		MigrationsDir: md,
	}

	db, err := OpenDBFromDBConf(conf)
	require.NoError(t, err)

	db.Exec("DROP TABLE goose_db_version")
	db.Exec("DROP TABLE test")

	// up
	err = RunMigrationsOnDb(conf, conf.MigrationsDir, 20160318040509, db)
	require.NoError(t, err)

	// down
	err = RunMigrationsOnDb(conf, conf.MigrationsDir, 0, db)
	require.NoError(t, err)

	rows, err := db.Query("SELECT value FROM test")
	require.Error(t, err) // table won't exist

	// up
	err = RunMigrationsOnDb(conf, conf.MigrationsDir, 20160318040509, db)
	require.NoError(t, err)

	rows, err = db.Query("SELECT value FROM test")
	require.NoError(t, err)
	defer rows.Close()
	var values []string
	for rows.Next() {
		var value string
		err := rows.Scan(&value)
		require.NoError(t, err)
		values = append(values, value)
	}

	assert.Len(t, values, 2)
	assert.Contains(t, values, "one")
	assert.Contains(t, values, "two")
}
func TestRunMigrationsOnDb_upDownUp_sqlite3(t *testing.T) {
	config, dbFile := getSqlite3Driver(t)
	defer os.Remove(dbFile)
	testRunMigrationsOnDb_upDownUp(t, config)
}
func TestRunMigrationsOnDb_upDownUp_mysql(t *testing.T) {
	testRunMigrationsOnDb_upDownUp(t, getMysqlDriver(t))
}
func TestRunMigrationsOnDb_upDownUp_postgres(t *testing.T) {
	testRunMigrationsOnDb_upDownUp(t, getPostgresDriver(t))
}
func TestRunMigrationsOnDb_upDownUp_redshift(t *testing.T) {
	testRunMigrationsOnDb_upDownUp(t, getRedshiftDriver(t))
}
