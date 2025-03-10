# 📚 Go RESTful API for Bookstore Management

---

## 🔑 Authentication

### Login

```bash
curl -X POST "http://localhost:8080/login" \
-H "Content-Type: application/json" \
-d '{"username":"admin","password":"password"}'
```

---

## 👤 Authors

### Create an author
```bash
curl -X POST http://localhost:8080/authors \
-H "Content-Type: application/json" \
-d "{\"first_name\":\"F. Scott\",\"last_name\":\"Fitzgerald\",\"bio\":\"American novelist.\"}"
```

### Get all authors
```bash
curl -X GET http://localhost:8080/authors
```

### Get an author by ID
```bash
curl -X GET http://localhost:8080/authors/1
```

### Update an author
```bash
curl -X PUT http://localhost:8080/authors/1 \
-H "Content-Type: application/json" \
-d "{\"first_name\":\"F. Scott\",\"last_name\":\"Fitzgerald\",\"bio\":\"Updated bio.\"}"
```

### Delete an author
```bash
curl -X DELETE http://localhost:8080/authors/1
```

### Search author by first name
```bash
curl -X GET "http://localhost:8080/authors?first_name=F.%20Scott"
```

---

## 📚 Books

### Create a book
```bash
curl -X POST http://localhost:8080/books \
-H "Content-Type: application/json" \
-d "{\"title\":\"The Great Gatsby\",\"author\":{\"id\":1},\"genres\":[\"Classic\",\"Fiction\"],\"published_at\":\"1925-04-10T00:00:00Z\",\"price\":10.99,\"stock\":20}"
```

### Get all books
```bash
curl -X GET http://localhost:8080/books
```

### Get a book by ID
```bash
curl -X GET http://localhost:8080/books/1
```

### Update a book
```bash
curl -X PUT http://localhost:8080/books/1 \
-H "Content-Type: application/json" \
-d "{\"title\":\"The Great Gatsby Updated\",\"author\":{\"id\":1},\"genres\":[\"Classic\",\"Fiction\"],\"published_at\":\"1925-04-10T00:00:00Z\",\"price\":12.99,\"stock\":15}"
```

### Delete a book
```bash
curl -X DELETE http://localhost:8080/books/1
```

### Search book by title
```bash
curl -X GET "http://localhost:8080/books?title=Gatsby"
```

---

## 👥 Customers

### Create a customer
```bash
curl -X POST http://localhost:8080/customers \
-H "Content-Type: application/json" \
-d "{\"name\":\"John Doe\",\"email\":\"sososo@example.com\",\"street\":\"123 Main St\",\"city\":\"Anytown\",\"state\":\"CA\",\"postal_code\":\"12345\",\"country\":\"USA\"}"
```

### Get all customers
```bash
curl -X GET http://localhost:8080/customers
```

### Get a customer by ID
```bash
curl -X GET http://localhost:8080/customers/7
```

### Update a customer
```bash
curl -X PUT http://localhost:8080/customers/9 \
-H "Content-Type: application/json" \
-d "{\"name\":\"John Doe\",\"email\":\"john.new@example.com\",\"street\":\"123 Main St\",\"city\":\"Anytown\",\"state\":\"CA\",\"postal_code\":\"12345\",\"country\":\"USA\"}"
```

### Delete a customer
```bash
curl -X DELETE http://localhost:8080/customers/1
```

### Search customer by name
```bash
curl -X GET "http://localhost:8080/customers?name=John"
```

---

## 🛒 Orders

### Create an order
```bash
curl -X POST http://localhost:8080/orders \
-H "Content-Type: application/json" \
-d "{\"customer\":{\"id\":1},\"total_price\":59.97,\"status\":\"pending\",\"items\":[{\"book\":{\"id\":1},\"quantity\":3}]}"
```

### Get all orders
```bash
curl -X GET http://localhost:8080/orders
```

### Get an order by ID
```bash
curl -X GET http://localhost:8080/orders/1
```

### Update an order
```bash
curl -X PUT http://localhost:8080/orders/3 \
-H "Content-Type: application/json" \
-d "{\"customer\":{\"id\":1},\"total_price\":79.97,\"status\":\"confirmed\",\"items\":[{\"book\":{\"id\":1},\"quantity\":4}]}"
```

### Delete an order
```bash
curl -X DELETE http://localhost:8080/orders/1
```

### Search orders by customer and status
```bash
curl -X GET "http://localhost:8080/orders?customer_name=John&status=pending"
```

---

## 📊 Reports

### Generate sales report (by date range)
```bash
curl -X GET "http://localhost:8080/reports?start_date=2025-01-01&end_date=2025-04-30"
```

---

## ⚙️ Rate Limiting

- Implemented in `ratelimiter.go` under `auth` directory.
- Limit: **10 requests per minute per user**.

### Test rate limiting:

```powershell
# Store your authentication token
$token = "your-jwt-token-here"

# Send multiple requests to test rate limiting
for ($i=1; $i -le 15; $i++) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/api/books" -Method GET -Headers @{ "Authorization" = "Bearer $token" }
        Write-Host ("Request " + $i + ": " + $response.StatusCode)
        if ($response.StatusCode -eq 200) {
            Write-Host "Response: " + $response.Content
        }
    } catch {
        $errorStatus = $_.Exception.Response.StatusCode.Value_
        Write-Host ("Request " + $i + ": Failed with status " + $errorStatus)
    }
    Start-Sleep -Seconds 2
}
```

---

## 📈 Prometheus Monitoring & Logging

### Install Prometheus client:

```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

### Sample Logs:

```
2025/03/04 18:04:14 Server running on http://localhost:8080
2025/03/04 18:04:21 [GET] /metrics [::1]:58459 1.1378ms
2025/03/04 18:04:34 [GET] /api/books [::1]:58467 95.1076ms
2025/03/04 18:04:36 [GET] /metrics [::1]:58459 554.2µs
2025/03/04 18:04:37 [GET] /api/books [::1]:58467 3.0549ms
```

---

## 🚨 Note

> **⚠️ All routes should include the login token for authenticated requests.**

Example:

```bash
curl -X GET http://localhost:8080/authors \
-H "Authorization: Bearer <your-token-here>"
```

---

## ✍️ Authors

Made by Fatima Zahra Fadel & Salma Zouhairi.
