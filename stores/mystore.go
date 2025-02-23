package stores

import "database/sql"

type PostgresStore struct {
	DB            *sql.DB
	BookStore     *PostgresBookStore
	AuthorStore   *PostgresAuthorStore
	CustomerStore *PostgresCustomerStore
	OrderStore    *PostgresOrderStore
	//ReportStore   *s.PostgresReportStore
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{
		DB:            db,
		BookStore:     &PostgresBookStore{DB: db},
		AuthorStore:   &PostgresAuthorStore{DB: db},
		CustomerStore: &PostgresCustomerStore{DB: db},
		OrderStore:    &PostgresOrderStore{DB: db},
		//	ReportStore:   &s.PostgresReportStore{DB: db},
	}
}
