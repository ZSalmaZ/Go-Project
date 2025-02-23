package stores

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/lib/pq"
	m "project.com/myproject/models"
)

type PostgresBookStore struct {
	DB *sql.DB
}

func NewPostgresBookStore(db *sql.DB) *PostgresBookStore {
	return &PostgresBookStore{DB: db}
}

// ✅ Create a Book and Ensure Genres Are Handled Correctly
func (s *PostgresStore) CreateBook(book m.Book) (m.Book, error) {
	// Step 1: Insert the book into `books` table
	query := `INSERT INTO books (title, author_id, published_at, price, stock) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var bookID int
	err := s.DB.QueryRow(query, book.Title, book.Author.ID, book.PublishedAt, book.Price, book.Stock).Scan(&bookID)
	if err != nil {
		log.Println("❌ Error inserting book:", err)
		return m.Book{}, err
	}

	// Step 2: Ensure each genre exists and link it in `book_genres`
	for _, genre := range book.Genres {
		var genreID int

		// Check if the genre already exists
		err := s.DB.QueryRow(`SELECT id FROM genres WHERE name = $1`, genre).Scan(&genreID)
		if err == sql.ErrNoRows {
			// Genre does not exist, create it
			err = s.DB.QueryRow(`INSERT INTO genres (name) VALUES ($1) RETURNING id`, genre).Scan(&genreID)
			if err != nil {
				log.Println("❌ Error inserting genre:", err)
				return m.Book{}, err
			}
		} else if err != nil {
			log.Println("❌ Error checking genre:", err)
			return m.Book{}, err
		}

		// Link the book to the genre
		_, err = s.DB.Exec(`INSERT INTO book_genres (book_id, genre_id) VALUES ($1, $2)`, bookID, genreID)
		if err != nil {
			log.Println("❌ Error linking book to genre:", err)
			return m.Book{}, err
		}
	}

	book.ID = bookID
	return book, nil
}

// ✅ Fetch a Single Book with Linked Genres
func (s *PostgresStore) GetBook(id int) (m.Book, error) {
	query := `SELECT b.id, b.title, b.author_id, b.published_at, b.price, b.stock, 
	                 COALESCE(array_agg(g.name) FILTER (WHERE g.name IS NOT NULL), '{}') AS genres
	          FROM books b
	          LEFT JOIN book_genres bg ON b.id = bg.book_id
	          LEFT JOIN genres g ON bg.genre_id = g.id
	          WHERE b.id = $1
	          GROUP BY b.id`

	var book m.Book
	var authorID int
	var genres pq.StringArray

	err := s.DB.QueryRow(query, id).Scan(&book.ID, &book.Title, &authorID, &book.PublishedAt, &book.Price, &book.Stock, &genres)
	if err != nil {
		log.Println("❌ Error retrieving book:", err)
		return m.Book{}, err
	}

	book.Genres = genres
	book.Author, _ = s.GetAuthor(authorID)

	return book, nil
}

// ✅ Fetch All Books with Their Genres
func (s *PostgresStore) GetAllBooks() ([]m.Book, error) {
	query := `SELECT b.id, b.title, b.author_id, b.published_at, b.price, b.stock, 
	                 COALESCE(array_agg(g.name) FILTER (WHERE g.name IS NOT NULL), '{}') AS genres
	          FROM books b
	          LEFT JOIN book_genres bg ON b.id = bg.book_id
	          LEFT JOIN genres g ON bg.genre_id = g.id
	          GROUP BY b.id`

	rows, err := s.DB.Query(query)
	if err != nil {
		log.Println("❌ Error retrieving books:", err)
		return nil, err
	}
	defer rows.Close()

	var books []m.Book
	for rows.Next() {
		var book m.Book
		var authorID int
		var genres pq.StringArray

		err := rows.Scan(&book.ID, &book.Title, &authorID, &book.PublishedAt, &book.Price, &book.Stock, &genres)
		if err != nil {
			log.Println("❌ Error scanning book:", err)
			return nil, err
		}
		book.Genres = genres
		book.Author, _ = s.GetAuthor(authorID)
		books = append(books, book)
	}

	return books, nil
}

// ✅ Update a Book (Handles Both Book Info & Genres)
func (s *PostgresStore) UpdateBook(id int, book m.Book) error {
	// Step 1: Ensure the book exists before updating
	var exists bool
	err := s.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM books WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		log.Println("❌ Error checking book existence:", err)
		return err
	}
	if !exists {
		log.Println("❌ Book not found for update")
		return sql.ErrNoRows
	}

	// Step 2: Update book details
	query := `UPDATE books SET title = $1, author_id = $2, published_at = $3, price = $4, stock = $5 WHERE id = $6`
	_, err = s.DB.Exec(query, book.Title, book.Author.ID, book.PublishedAt, book.Price, book.Stock, id)
	if err != nil {
		log.Println("❌ Error updating book:", err)
		return err
	}

	// Step 3: Update genres only if new genres are provided
	if len(book.Genres) > 0 {
		// Delete old genre links
		_, err = s.DB.Exec("DELETE FROM book_genres WHERE book_id = $1", id)
		if err != nil {
			log.Println("❌ Error clearing book genres:", err)
			return err
		}

		// Step 4: Ensure each genre exists and re-link it
		for _, genre := range book.Genres {
			var genreID int
			err := s.DB.QueryRow(`SELECT id FROM genres WHERE name = $1`, genre).Scan(&genreID)
			if err == sql.ErrNoRows {
				// Genre does not exist, create it
				err = s.DB.QueryRow(`INSERT INTO genres (name) VALUES ($1) RETURNING id`, genre).Scan(&genreID)
				if err != nil {
					log.Println("❌ Error inserting genre:", err)
					return err
				}
			} else if err != nil {
				log.Println("❌ Error checking genre:", err)
				return err
			}

			// Link the book to the genre
			_, err = s.DB.Exec(`INSERT INTO book_genres (book_id, genre_id) VALUES ($1, $2)`, id, genreID)
			if err != nil {
				log.Println("❌ Error linking book to genre:", err)
				return err
			}
		}
	}

	log.Println("✅ Book updated successfully")
	return nil
}

// ✅ Delete a Book (Decrease Stock or Delete Completely)
func (s *PostgresStore) DeleteBook(id int) error {
	var stock int
	err := s.DB.QueryRow("SELECT stock FROM books WHERE id = $1", id).Scan(&stock)
	if err != nil {
		log.Println("❌ Error retrieving book stock:", err)
		return err
	}

	if stock > 1 {
		_, err = s.DB.Exec("UPDATE books SET stock = stock - 1 WHERE id = $1", id)
	} else {
		_, err = s.DB.Exec("DELETE FROM book_genres WHERE book_id = $1", id)
		if err == nil {
			_, err = s.DB.Exec("DELETE FROM books WHERE id = $1", id)
		}
	}

	return err
}

func (s *PostgresStore) SearchBooks(criteria m.SearchCriteriaBooks) ([]m.Book, error) {
	// Base query
	query := `SELECT b.id, b.title, b.published_at, b.price, b.stock, 
	                 a.id, a.first_name, a.last_name,
	                 COALESCE(array_agg(g.name) FILTER (WHERE g.name IS NOT NULL), '{}') AS genres
	          FROM books b
	          JOIN authors a ON b.author_id = a.id
	          LEFT JOIN book_genres bg ON b.id = bg.book_id
	          LEFT JOIN genres g ON bg.genre_id = g.id
	          WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	// Add filters dynamically
	if criteria.Title != "" {
		query += ` AND b.title ILIKE $` + strconv.Itoa(argCount)
		args = append(args, "%"+criteria.Title+"%")
		argCount++
	}

	if criteria.AuthorFirstName != "" {
		query += ` AND a.first_name ILIKE $` + strconv.Itoa(argCount)
		args = append(args, "%"+criteria.AuthorFirstName+"%")
		argCount++
	}

	if criteria.AuthorName != "" {
		query += ` AND a.last_name ILIKE $` + strconv.Itoa(argCount)
		args = append(args, "%"+criteria.AuthorName+"%")
		argCount++
	}

	query += " GROUP BY b.id, a.id"

	// Execute query
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		log.Println("❌ Error searching books:", err)
		return nil, err
	}
	defer rows.Close()

	// Collect results
	var books []m.Book
	for rows.Next() {
		var book m.Book
		var author m.Author
		var genres pq.StringArray
		err := rows.Scan(&book.ID, &book.Title, &book.PublishedAt, &book.Price, &book.Stock,
			&author.ID, &author.FirstName, &author.LastName, &genres)
		if err != nil {
			log.Println("❌ Error scanning book:", err)
			return nil, err
		}
		book.Genres = genres
		book.Author = author
		books = append(books, book)
	}

	// Check if no books were found
	if len(books) == 0 {
		log.Println("🔍 No matching books found")
		return nil, sql.ErrNoRows
	}

	log.Println("✅ Books search successful")
	return books, nil
}
