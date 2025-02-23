package stores

import (
	"database/sql"
	"log"
	"strconv"

	m "project.com/myproject/models"
)

type PostgresCustomerStore struct {
	DB *sql.DB
}

func NewPostgresCustomerStore(db *sql.DB) *PostgresCustomerStore {
	return &PostgresCustomerStore{DB: db}
}

// ✅ Create a Customer (Fixed Address Fields)
func (s *PostgresStore) CreateCustomer(customer m.Customer) (m.Customer, error) {
	query := `INSERT INTO customers (name, email, street, city, state, postal_code, country) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var id int
	err := s.DB.QueryRow(query, customer.Name, customer.Email, customer.Street, customer.City, customer.State, customer.PostalCode, customer.Country).Scan(&id)
	if err != nil {
		log.Println("❌ Error inserting customer:", err)
		return m.Customer{}, err
	}
	customer.ID = id
	return customer, nil
}

// ✅ Fetch a Single Customer with Correct Fields
func (s *PostgresStore) GetCustomer(id int) (m.Customer, error) {
	query := `SELECT id, name, email, street, city, state, postal_code, country FROM customers WHERE id = $1`
	var customer m.Customer
	err := s.DB.QueryRow(query, id).Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Street, &customer.City, &customer.State, &customer.PostalCode, &customer.Country)
	if err != nil {
		log.Println("❌ Error retrieving customer:", err)
		return m.Customer{}, err
	}
	return customer, nil
}

// ✅ Update a Customer (Fixed Address Fields)
func (s *PostgresStore) UpdateCustomer(id int, customer m.Customer) error {
	query := `UPDATE customers SET name = $1, email = $2, street = $3, city = $4, state = $5, postal_code = $6, country = $7 WHERE id = $8`
	_, err := s.DB.Exec(query, customer.Name, customer.Email, customer.Street, customer.City, customer.State, customer.PostalCode, customer.Country, id)
	return err
}

// ✅ Delete a Customer
func (s *PostgresStore) DeleteCustomer(id int) error {
	query := "DELETE FROM customers WHERE id = $1"
	_, err := s.DB.Exec(query, id)
	return err
}

// ✅ Fetch All Customers (Fixed Address Fields)
func (s *PostgresStore) GetAllCustomers() ([]m.Customer, error) {
	query := `SELECT id, name, email, street, city, state, postal_code, country FROM customers`
	rows, err := s.DB.Query(query)
	if err != nil {
		log.Println("❌ Error retrieving customers:", err)
		return nil, err
	}
	defer rows.Close()

	var customers []m.Customer
	for rows.Next() {
		var customer m.Customer
		err := rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Street, &customer.City, &customer.State, &customer.PostalCode, &customer.Country)
		if err != nil {
			log.Println("❌ Error scanning customer:", err)
			return nil, err
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		log.Println("❌ Error in rows iteration:", err)
		return nil, err
	}

	return customers, nil
}

func (s *PostgresStore) SearchCustomers(criteria m.SearchCriteriaCustomers) ([]m.Customer, error) {
	// Base query
	query := `SELECT id, name, email, street, city, state, postal_code, country FROM customers WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	// Add filters dynamically
	if criteria.Name != "" {
		query += ` AND name ILIKE $` + strconv.Itoa(argCount)
		args = append(args, "%"+criteria.Name+"%")
		argCount++
	}

	if criteria.Email != "" {
		query += ` AND email ILIKE $` + strconv.Itoa(argCount)
		args = append(args, "%"+criteria.Email+"%")
		argCount++
	}

	// Execute query
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		log.Println("❌ Error searching customers:", err)
		return nil, err
	}
	defer rows.Close()

	// Collect results
	var customers []m.Customer
	for rows.Next() {
		var customer m.Customer
		err := rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Street, &customer.City, &customer.State, &customer.PostalCode, &customer.Country)
		if err != nil {
			log.Println("❌ Error scanning customer:", err)
			return nil, err
		}
		customers = append(customers, customer)
	}

	// Check if no customers were found
	if len(customers) == 0 {
		log.Println("🔍 No matching customers found")
		return nil, sql.ErrNoRows
	}

	log.Println("✅ Customers search successful")
	return customers, nil
}
