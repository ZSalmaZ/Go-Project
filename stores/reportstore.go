// // reportstore.go
// package stores

// import (
// 	"context"
// 	"database/sql"
// 	"log"
// 	"time"

// 	m "project.com/myproject/models"
// )

// type PostgresReportStore struct {
// 	DB *sql.DB
// }

// func NewPostgresReportStore(db *sql.DB) *PostgresReportStore {
// 	return &PostgresReportStore{DB: db}
// }

// // GenerateSalesReport aggregates total revenue, order count, and top selling books for the given date range.
// func (rs *PostgresStore) GenerateSalesReport(ctx context.Context, startDate, endDate time.Time) (m.SalesReport, error) {
// 	// Query orders to get total revenue and order count.
// 	queryOrders := `
// 		SELECT total_price
// 		FROM orders
// 		WHERE created_at >= $1 AND created_at <= $2
// 	`
// 	rows, err := rs.DB.QueryContext(ctx, queryOrders, startDate, endDate)
// 	if err != nil {
// 		return m.SalesReport{}, err
// 	}
// 	defer rows.Close()

// 	var totalRevenue float64
// 	var totalOrders int
// 	for rows.Next() {
// 		var price float64
// 		if err := rows.Scan(&price); err != nil {
// 			log.Println("Error scanning order total price:", err)
// 			return m.SalesReport{}, err
// 		}
// 		totalRevenue += price
// 		totalOrders++
// 	}

// 	// Query to get top selling books by summing the quantity sold.
// 	// 	queryTopBooks := `
// 	// 		SELECT b.id, b.title, SUM(oi.quantity) AS total_quantity
// 	// 		FROM order_items oi
// 	// 		JOIN orders o ON oi.order_id = o.id
// 	// 		JOIN books b ON oi.book_id = b.id
// 	// 		WHERE o.created_at >= $1 AND o.created_at <= $2
// 	// 		GROUP BY b.id, b.title
// 	// 		ORDER BY total_quantity DESC
// 	// 	`

// 	// 	rowsTop, err := rs.DB.QueryContext(ctx, queryTopBooks, startDate, endDate)
// 	// 	if err != nil {
// 	// 		return m.SalesReport{}, err
// 	// 	}
// 	// 	defer rowsTop.Close()

// 	// 	var topSellingBooks []m.BookSales
// 	// 	for rowsTop.Next() {
// 	// 		var bookID int
// 	// 		var title string
// 	// 		var quantity int
// 	// 		if err := rowsTop.Scan(&bookID, &title, &quantity); err != nil {
// 	// 			log.Println("Error scanning top selling book:", err)
// 	// 			return m.SalesReport{}, err
// 	// 		}
// 	// 		// For brevity, we only include the book's ID and Title here.
// 	// 		book := m.Book{
// 	// 			ID:    bookID,
// 	// 			Title: title,
// 	// 		}
// 	// 		topSellingBooks = append(topSellingBooks, m.BookSales{
// 	// 			Book:     book,
// 	// 			Quantity: quantity,
// 	// 		})
// 	// 	}

// 	// 	report := m.SalesReport{
// 	// 		Timestamp:       time.Now(),
// 	// 		TotalRevenue:    totalRevenue,
// 	// 		TotalOrders:     totalOrders,
// 	// 		TopSellingBooks: topSellingBooks,
// 	// 	}
// 	// 	return report, nil
// 	// }

// 	// // GetSalesReport is a simple wrapper (you could later store reports in a table if needed).
// 	// func (rs *PostgresStore) GetSalesReport(ctx context.Context, startDate, endDate time.Time) (m.SalesReport, error) {
// 	// 	return rs.GenerateSalesReport(ctx, startDate, endDate)
// 	// }

// 	// Query to get top selling books by summing the quantity sold.
// 	// Updated to also select published_at, price, and stock.
// 	queryTopBooks := `
// 	SELECT b.id, b.title, b.published_at, b.price, b.stock, SUM(oi.quantity) AS total_quantity
// 	FROM order_items oi
// 	JOIN orders o ON oi.order_id = o.id
// 	JOIN books b ON oi.book_id = b.id
// 	WHERE o.created_at >= $1 AND o.created_at <= $2
// 	GROUP BY b.id, b.title, b.published_at, b.price, b.stock
// 	ORDER BY total_quantity DESC
// `
// 	rowsTop, err := rs.DB.QueryContext(ctx, queryTopBooks, startDate, endDate)
// 	if err != nil {
// 		return m.SalesReport{}, err
// 	}
// 	defer rowsTop.Close()

// 	var topSellingBooks []m.BookSales
// 	for rowsTop.Next() {
// 		var bookID int
// 		var title string
// 		var publishedAt time.Time
// 		var price float64
// 		var stock int
// 		var quantity int
// 		if err := rowsTop.Scan(&bookID, &title, &publishedAt, &price, &stock, &quantity); err != nil {
// 			log.Println("Error scanning top selling book:", err)
// 			return m.SalesReport{}, err
// 		}
// 		book := m.Book{
// 			ID:          bookID,
// 			Title:       title,
// 			PublishedAt: publishedAt,
// 			Price:       price,
// 			Stock:       stock,
// 			// Note: Author is not joined here. If you need it, you can either join the authors table as well or call GetBook separately.
// 		}
// 		topSellingBooks = append(topSellingBooks, m.BookSales{
// 			Book:     book,
// 			Quantity: quantity,
// 		})
// 	}

// 	// Create the report with the aggregated data.
// 	report := m.SalesReport{
// 		Timestamp:       time.Now(),
// 		TotalRevenue:    totalRevenue,
// 		TotalOrders:     totalOrders,
// 		TopSellingBooks: topSellingBooks,
// 	}

// 	return report, nil
// }

// reportstore.go
package stores

import (
	"context"
	"database/sql"
	"log"
	"time"

	m "project.com/myproject/models"
)

type PostgresReportStore struct {
	DB *sql.DB
}

func NewPostgresReportStore(db *sql.DB) *PostgresReportStore {
	return &PostgresReportStore{DB: db}
}

// buildSalesReport aggregates total revenue, order count, and top selling books for the given date range.
func (rs *PostgresStore) buildSalesReport(ctx context.Context, startDate, endDate time.Time) (m.SalesReport, error) {
	// Query orders to get total revenue and order count.
	queryOrders := `
		SELECT total_price
		FROM orders
		WHERE created_at >= $1 AND created_at <= $2
	`
	rows, err := rs.DB.QueryContext(ctx, queryOrders, startDate, endDate)
	if err != nil {
		return m.SalesReport{}, err
	}
	defer rows.Close()

	var totalRevenue float64
	var totalOrders int
	for rows.Next() {
		var price float64
		if err := rows.Scan(&price); err != nil {
			log.Println("Error scanning order total price:", err)
			return m.SalesReport{}, err
		}
		totalRevenue += price
		totalOrders++
	}

	// Query to get top selling books by summing the quantity sold.
	// Updated to also select published_at, price, and stock.
	queryTopBooks := `
		SELECT b.id, b.title, b.published_at, b.price, b.stock, SUM(oi.quantity) AS total_quantity
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		JOIN books b ON oi.book_id = b.id
		WHERE o.created_at >= $1 AND o.created_at <= $2
		GROUP BY b.id, b.title, b.published_at, b.price, b.stock
		ORDER BY total_quantity DESC
	`
	rowsTop, err := rs.DB.QueryContext(ctx, queryTopBooks, startDate, endDate)
	if err != nil {
		return m.SalesReport{}, err
	}
	defer rowsTop.Close()

	var topSellingBooks []m.BookSales
	for rowsTop.Next() {
		var bookID int
		var title string
		var publishedAt time.Time
		var price float64
		var stock int
		var quantity int
		if err := rowsTop.Scan(&bookID, &title, &publishedAt, &price, &stock, &quantity); err != nil {
			log.Println("Error scanning top selling book:", err)
			return m.SalesReport{}, err
		}
		book := m.Book{
			ID:          bookID,
			Title:       title,
			PublishedAt: publishedAt,
			Price:       price,
			Stock:       stock,
		}
		topSellingBooks = append(topSellingBooks, m.BookSales{
			Book:     book,
			Quantity: quantity,
		})
	}

	report := m.SalesReport{
		Timestamp:       time.Now(),
		TotalRevenue:    totalRevenue,
		TotalOrders:     totalOrders,
		TopSellingBooks: topSellingBooks,
	}

	return report, nil
}

// GetSalesReport is a simple wrapper that calls buildSalesReport.
func (rs *PostgresStore) GetSalesReport(ctx context.Context, startDate, endDate time.Time) (m.SalesReport, error) {
	return rs.buildSalesReport(ctx, startDate, endDate)
}
