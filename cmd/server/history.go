package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// purchaseHistoryItem represents one record in the user’s history.
type purchaseHistoryItem struct {
	PurchaseID      int       `json:"purchase_id"`
	ProductID       int       `json:"product_id"`
	ProductName     string    `json:"product_name"`
	Quantity        int       `json:"quantity"`
	TotalPriceCents int64     `json:"total_price_cents"`
	PurchasedAt     time.Time `json:"purchased_at"`
}

func getUserHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user_id from context
	uidVal := r.Context().Value("user_id")
	if uidVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := uidVal.(int)

	// Query purchases joined with products
	rows, err := db.Query(`
    SELECT
      pu.id,
      pu.product_id,
      pr.name,
      pu.quantity,
      pu.total_price_cents,
      pu.purchased_at
    FROM purchases pu
    JOIN products pr ON pu.product_id = pr.id
    WHERE pu.user_id = $1
    ORDER BY pu.purchased_at DESC;
  `, userID)
	if err != nil {
		http.Error(w, "Failed to query purchase history", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	history := make([]purchaseHistoryItem, 0)

	for rows.Next() {
		var item purchaseHistoryItem
		if err := rows.Scan(
			&item.PurchaseID,
			&item.ProductID,
			&item.ProductName,
			&item.Quantity,
			&item.TotalPriceCents,
			&item.PurchasedAt,
		); err != nil {
			http.Error(w, "Error scanning history", http.StatusInternalServerError)
			return
		}
		history = append(history, item)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating history", http.StatusInternalServerError)
		return
	}

	// Return as JSON array
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
