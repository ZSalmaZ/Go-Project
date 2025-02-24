package stores

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	m "project.com/myproject/models"
)

type PostgresOrderStore struct {
	DB *sql.DB
}

func NewPostgresOrderStore(db *sql.DB) *PostgresOrderStore {
	return &PostgresOrderStore{DB: db}
}

// Order Store Methods
func (s *PostgresStore) CreateOrder(ctx context.Context, order m.Order) (m.Order, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	// Step 1: Insert the order
	query := `INSERT INTO orders (customer_id, total_price, status) 
	          VALUES ($1, $2, $3) RETURNING id`
	var orderID int
	err := s.DB.QueryRowContext(ctx, query, order.Customer.ID, order.TotalPrice, order.Status).Scan(&orderID)
	if err != nil {
		log.Println("❌ Error inserting order:", err)
		return m.Order{}, err
	}
	order.ID = orderID

	// Step 2: Handle order items (link books to order and decrease stock)
	for _, item := range order.Items {
		// Check if book exists and has enough stock
		var availableStock int
		err := s.DB.QueryRow("SELECT stock FROM books WHERE id = $1", item.Book.ID).Scan(&availableStock)
		if err != nil {
			log.Println("❌ Error checking book stock:", err)
			return m.Order{}, err
		}

		if availableStock < item.Quantity {
			log.Println("❌ Not enough stock for book:", item.Book.ID)
			return m.Order{}, sql.ErrNoRows
		}

		// Insert into order_items table
		_, err = s.DB.Exec(`INSERT INTO order_items (order_id, book_id, quantity) VALUES ($1, $2, $3)`, orderID, item.Book.ID, item.Quantity)
		if err != nil {
			log.Println("❌ Error inserting order item:", err)
			return m.Order{}, err
		}

		// Decrease stock
		_, err = s.DB.Exec("UPDATE books SET stock = stock - $1 WHERE id = $2", item.Quantity, item.Book.ID)
		if err != nil {
			log.Println("❌ Error decreasing book stock:", err)
			return m.Order{}, err
		}
	}

	log.Println("✅ Order created successfully with stock updated")
	return order, nil
}

func (s *PostgresStore) GetOrder(ctx context.Context, id int) (m.Order, error) {
	// Step 1: Fetch the Order & Customer
	query := `SELECT o.id, o.customer_id, c.name, c.email, c.street, c.city, c.state, c.postal_code, c.country, o.total_price, o.status, o.created_at
	          FROM orders o
	          JOIN customers c ON o.customer_id = c.id
	          WHERE o.id = $1`

	var order m.Order
	err := s.DB.QueryRow(query, id).Scan(
		&order.ID, &order.Customer.ID, &order.Customer.Name, &order.Customer.Email, &order.Customer.Street, &order.Customer.City, &order.Customer.State, &order.Customer.PostalCode, &order.Customer.Country,
		&order.TotalPrice, &order.Status, &order.CreatedAt,
	)
	if err != nil {
		log.Println("❌ Error retrieving order:", err)
		return m.Order{}, err
	}

	// Step 2: Fetch Order Items (Books + Authors)
	itemQuery := `SELECT b.id, b.title, b.published_at, b.price, b.stock, a.id, a.first_name, a.last_name, oi.quantity
	              FROM order_items oi
	              JOIN books b ON oi.book_id = b.id
	              JOIN authors a ON b.author_id = a.id
	              WHERE oi.order_id = $1`
	rows, err := s.DB.Query(itemQuery, id)
	if err != nil {
		log.Println("❌ Error retrieving order items:", err)
		return m.Order{}, err
	}
	defer rows.Close()

	var items []m.OrderItem
	for rows.Next() {
		var item m.OrderItem
		var author m.Author
		err := rows.Scan(&item.Book.ID, &item.Book.Title, &item.Book.PublishedAt, &item.Book.Price, &item.Book.Stock,
			&author.ID, &author.FirstName, &author.LastName, &item.Quantity)
		if err != nil {
			log.Println("❌ Error scanning order item:", err)
			return m.Order{}, err
		}
		item.Book.Author = author
		items = append(items, item)
	}

	order.Items = items
	return order, nil
}

func (s *PostgresStore) UpdateOrder(ctx context.Context, id int, order m.Order) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	query := "UPDATE orders SET customer_id = $1, total_price = $2, status = $3 WHERE id = $4"
	_, err := s.DB.Exec(query, order.Customer.ID, order.TotalPrice, order.Status, id)
	return err
}

func (s *PostgresStore) DeleteOrder(ctx context.Context, id int) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	query := "DELETE FROM orders WHERE id = $1"
	_, err := s.DB.Exec(query, id)
	return err
}

func (s *PostgresStore) GetAllOrders(ctx context.Context) ([]m.Order, error) {
	query := `SELECT o.id, o.customer_id, c.name, c.email, c.street, c.city, c.state, c.postal_code, c.country, o.total_price, o.status, o.created_at
	          FROM orders o
	          JOIN customers c ON o.customer_id = c.id`

	rows, err := s.DB.Query(query)
	if err != nil {
		log.Println("❌ Error retrieving orders:", err)
		return nil, err
	}
	defer rows.Close()

	var orders []m.Order
	for rows.Next() {
		var order m.Order
		err := rows.Scan(&order.ID, &order.Customer.ID, &order.Customer.Name, &order.Customer.Email, &order.Customer.Street, &order.Customer.City, &order.Customer.State, &order.Customer.PostalCode, &order.Customer.Country,
			&order.TotalPrice, &order.Status, &order.CreatedAt)
		if err != nil {
			log.Println("❌ Error scanning order:", err)
			return nil, err
		}

		// Fetch Books for this Order
		itemQuery := `SELECT b.id, b.title, b.published_at, b.price, b.stock, a.id, a.first_name, a.last_name, oi.quantity
		              FROM order_items oi
		              JOIN books b ON oi.book_id = b.id
		              JOIN authors a ON b.author_id = a.id
		              WHERE oi.order_id = $1`
		itemRows, err := s.DB.Query(itemQuery, order.ID)
		if err != nil {
			log.Println("❌ Error retrieving order items:", err)
			return nil, err
		}

		var items []m.OrderItem
		for itemRows.Next() {
			var item m.OrderItem
			var author m.Author
			err := itemRows.Scan(&item.Book.ID, &item.Book.Title, &item.Book.PublishedAt, &item.Book.Price, &item.Book.Stock,
				&author.ID, &author.FirstName, &author.LastName, &item.Quantity)
			if err != nil {
				log.Println("❌ Error scanning order item:", err)
				return nil, err
			}
			item.Book.Author = author
			items = append(items, item)
		}

		itemRows.Close()
		order.Items = items
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *PostgresStore) SearchOrders(ctx context.Context, criteria m.SearchCriteriaOrders) ([]m.Order, error) {
	// Base query
	query := `SELECT o.id, o.customer_id, c.name, c.email, c.street, c.city, c.state, c.postal_code, c.country, o.total_price, o.status, o.created_at
	          FROM orders o
	          JOIN customers c ON o.customer_id = c.id
	          WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	// Add filters dynamically
	if criteria.CustomerName != "" {
		query += ` AND c.name ILIKE $` + strconv.Itoa(argCount)
		args = append(args, "%"+criteria.CustomerName+"%")
		argCount++
	}

	if criteria.Status != "" {
		query += ` AND o.status ILIKE $` + strconv.Itoa(argCount)
		args = append(args, "%"+criteria.Status+"%")
		argCount++
	}

	// Execute query
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		log.Println("❌ Error searching orders:", err)
		return nil, err
	}
	defer rows.Close()

	// Collect results
	var orders []m.Order
	for rows.Next() {
		var order m.Order
		err := rows.Scan(&order.ID, &order.Customer.ID, &order.Customer.Name, &order.Customer.Email, &order.Customer.Street, &order.Customer.City, &order.Customer.State, &order.Customer.PostalCode, &order.Customer.Country,
			&order.TotalPrice, &order.Status, &order.CreatedAt)
		if err != nil {
			log.Println("❌ Error scanning order:", err)
			return nil, err
		}
		orders = append(orders, order)
	}

	// Check if no orders were found
	if len(orders) == 0 {
		log.Println("🔍 No matching orders found")
		return nil, sql.ErrNoRows
	}

	log.Println("✅ Orders search successful")
	return orders, nil
}
