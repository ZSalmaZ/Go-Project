package stores

import (
	"database/sql"
	"log"
	"strconv"

	m "project.com/myproject/models"
)

type PostgresAuthorStore struct {
	DB *sql.DB
}

func NewPostgresAuthorStore(db *sql.DB) *PostgresAuthorStore {
	return &PostgresAuthorStore{DB: db}
}

// Author Store Methods
func (s *PostgresStore) CreateAuthor(author m.Author) (m.Author, error) {
	query := "INSERT INTO authors (first_name, last_name, bio) VALUES ($1, $2, $3) RETURNING id"
	var id int
	err := s.DB.QueryRow(query, author.FirstName, author.LastName, author.Bio).Scan(&id)
	if err != nil {
		log.Println("Error inserting author:", err)
		return m.Author{}, err
	}
	author.ID = id
	return author, nil
}

func (s *PostgresStore) GetAuthor(id int) (m.Author, error) {
	query := "SELECT id, first_name, last_name, bio FROM authors WHERE id = $1"
	var author m.Author
	err := s.DB.QueryRow(query, id).Scan(&author.ID, &author.FirstName, &author.LastName, &author.Bio)
	if err != nil {
		log.Println("Error retrieving author:", err)
		return m.Author{}, err
	}
	return author, nil
}

func (s *PostgresStore) UpdateAuthor(id int, author m.Author) error {
	query := "UPDATE authors SET first_name = $1, last_name = $2, bio = $3 WHERE id = $4"
	_, err := s.DB.Exec(query, author.FirstName, author.LastName, author.Bio, id)
	return err
}

func (s *PostgresStore) DeleteAuthor(id int) error {
	query := "DELETE FROM authors WHERE id = $1"
	_, err := s.DB.Exec(query, id)
	return err
}

func (s *PostgresStore) GetAllAuthors() ([]m.Author, error) {
	query := "SELECT id, first_name, last_name, bio FROM authors"
	rows, err := s.DB.Query(query)
	if err != nil {
		log.Println("Error retrieving authors:", err)
		return nil, err
	}
	defer rows.Close()

	var authors []m.Author
	for rows.Next() {
		var author m.Author
		err := rows.Scan(&author.ID, &author.FirstName, &author.LastName, &author.Bio)
		if err != nil {
			log.Println("Error scanning author:", err)
			return nil, err
		}
		authors = append(authors, author)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in rows iteration:", err)
		return nil, err
	}

	return authors, nil
}

func (s *PostgresStore) SearchAuthors(criteria m.SearchCriteriaAuthors) ([]m.Author, error) {
	// Base query
	query := `SELECT id, first_name, last_name, bio FROM authors WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	// Add filters dynamically
	if criteria.FirstName != "" {
		query += ` AND first_name ILIKE $` + strconv.Itoa(argCount)
		args = append(args, "%"+criteria.FirstName+"%")
		argCount++
	}

	if criteria.LastName != "" {
		query += ` AND last_name ILIKE $` + strconv.Itoa(argCount)
		args = append(args, "%"+criteria.LastName+"%")
		argCount++
	}

	// Execute query
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		log.Println("❌ Error searching authors:", err)
		return nil, err
	}
	defer rows.Close()

	// Collect results
	var authors []m.Author
	for rows.Next() {
		var author m.Author
		err := rows.Scan(&author.ID, &author.FirstName, &author.LastName, &author.Bio)
		if err != nil {
			log.Println("❌ Error scanning author:", err)
			return nil, err
		}
		authors = append(authors, author)
	}

	// Check if no authors were found
	if len(authors) == 0 {
		log.Println("🔍 No matching authors found")
		return nil, sql.ErrNoRows
	}

	log.Println("✅ Authors search successful")
	return authors, nil
}
