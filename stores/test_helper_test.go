package stores // ✅ NO stores_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testDB *sql.DB
var store *PostgresStore // ✅ No alias, just plain type

func TestMain(m *testing.M) {
	fmt.Println("⚙️ TestMain is running...") // ✅ Debug print

	var err error
	testDB, err = sql.Open("postgres", "postgres://postgres:2SOUsalma2003@localhost:5432/mylibrary?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	store = NewPostgresStore(testDB) // ✅ No alias! Just direct call

	code := m.Run()

	testDB.Close()
	os.Exit(code)
}
