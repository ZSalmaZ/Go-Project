package models

import (
	"time"
)

type Book struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Author      Author    `json:"author"`
	Genres      []string  `json:"genres"`
	PublishedAt time.Time `json:"published_at"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
}

type Author struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
}

type Customer struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type Order struct {
	ID         int         `json:"id"`
	Customer   Customer    `json:"customer"`
	TotalPrice float64     `json:"total_price"`
	Status     string      `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	Items      []OrderItem `json:"items"`
}

type OrderItem struct {
	Book     Book `json:"book"`
	Quantity int  `json:"quantity"`
}

type SalesReport struct {
	Timestamp       time.Time   `json:"timestamp"`
	TotalRevenue    float64     `json:"total_revenue"`
	TotalOrders     int         `json:"total_orders"`
	TopSellingBooks []BookSales `json:"top_selling_books"`
}

type BookSales struct {
	Book     Book `json:"book"`
	Quantity int  `json:"quantity_sold"`
}

// type SearchCriteriaBooks struct {
// 	Title           string `json:"title"`
// 	AuthorName      string `json:"author_last_name"`
// 	AuthorFirstName string `json:"author_first_name"`
// }

type SearchCriteriaBooks struct {
	Title           string  `json:"title"`
	AuthorName      string  `json:"author_last_name"`
	AuthorFirstName string  `json:"author_first_name"`
	MinPrice        float64 `json:"min_price"`
	MaxPrice        float64 `json:"max_price"`
}

type SearchCriteriaAuthors struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type SearchCriteriaCustomers struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type SearchCriteriaOrders struct {
	CustomerName string `json:"customer_name"`
	Status       string `json:"status"`
}
