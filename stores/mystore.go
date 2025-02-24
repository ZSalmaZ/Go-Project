package stores

import (
	"database/sql"
	"sync"
)

type PostgresStore struct {
	DB            *sql.DB
	Mu            sync.Mutex
	BookStore     *PostgresBookStore
	AuthorStore   *PostgresAuthorStore
	CustomerStore *PostgresCustomerStore
	OrderStore    *PostgresOrderStore
	ReportStore   *PostgresReportStore
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{
		DB:            db,
		BookStore:     &PostgresBookStore{DB: db},
		AuthorStore:   &PostgresAuthorStore{DB: db},
		CustomerStore: &PostgresCustomerStore{DB: db},
		OrderStore:    &PostgresOrderStore{DB: db},
		ReportStore:   &PostgresReportStore{DB: db},
	}
}
