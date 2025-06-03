package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// saleRecord represents one sale (joined with user & product).
type saleRecord struct {
	PurchaseID      int       `json:"purchase_id"`
	ProductID       int       `json:"product_id"`
	ProductName     string    `json:"product_name"`
	UserID          int       `json:"user_id"`
	Username        string    `json:"username"`
	Quantity        int       `json:"quantity"`
	TotalPriceCents int64     `json:"total_price_cents"`
	PurchasedAt     time.Time `json:"purchased_at"`
}

// getSalesHandler allows admins to filter by date range (from/to) and/or username.
func getSalesHandler(w http.ResponseWriter, r *http.Request) {
	// 1) Parse query params
	q := r.URL.Query()
	fromStr := q.Get("from")      // e.g. "2025-01-01"
	toStr := q.Get("to")          // e.g. "2025-06-01"
	username := q.Get("username") // exact match

	// 2) Build dynamic WHERE clauses
	clauses := []string{"1=1"} // start with a no-op clause
	args := []interface{}{}
	argIdx := 1

	if fromStr != "" {
		if _, err := time.Parse("2006-01-02", fromStr); err != nil {
			http.Error(w, "Invalid 'from' date: use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		clauses = append(clauses, "pu.purchased_at >= $"+strconv.Itoa(argIdx))
		args = append(args, fromStr)
		argIdx++
	}
	if toStr != "" {
		if _, err := time.Parse("2006-01-02", toStr); err != nil {
			http.Error(w, "Invalid 'to' date: use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		clauses = append(clauses, "pu.purchased_at <= $"+strconv.Itoa(argIdx))
		args = append(args, toStr+" 23:59:59") // include entire “to” day
		argIdx++
	}
	if username != "" {
		clauses = append(clauses, "u.username = $"+strconv.Itoa(argIdx))
		args = append(args, username)
		argIdx++
	}

	whereSQL := "WHERE " + strings.Join(clauses, " AND ")

	// 3) Final SQL (join purchases, users, products)
	query := `
		SELECT 
			pu.id,
			pu.product_id,
			pr.name,
			u.id,
			u.username,
			pu.quantity,
			pu.total_price_cents,
			pu.purchased_at
		FROM purchases pu
		JOIN products pr ON pu.product_id = pr.id
		JOIN users u ON pu.user_id = u.id
	` + whereSQL + `
		ORDER BY pu.purchased_at DESC;
	`

	// 4) Execute query
	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to query sales: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 5) Scan results into a slice
	sales := make([]saleRecord, 0)
	for rows.Next() {
		var s saleRecord
		if err := rows.Scan(
			&s.PurchaseID,
			&s.ProductID,
			&s.ProductName,
			&s.UserID,
			&s.Username,
			&s.Quantity,
			&s.TotalPriceCents,
			&s.PurchasedAt,
		); err != nil {
			http.Error(w, "Error scanning row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		sales = append(sales, s)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6) Return as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sales)
}
