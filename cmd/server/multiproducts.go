package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
)

type buyItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type buyRequest struct {
	Items           []buyItem `json:"items"`
	PaymentMethodID string    `json:"payment_method_id"`
}

type buyResponse struct {
	Success             bool   `json:"success"`
	StripePaymentIntent string `json:"stripe_payment_intent_id"`
}

func buyProductsHandler(w http.ResponseWriter, r *http.Request) {
	// Decode JSON payload
	var req buyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(req.Items) == 0 || req.PaymentMethodID == "" {
		http.Error(w, "items and payment_method_id are required", http.StatusBadRequest)
		return
	}

	// Extract logged-in user_id from context
	uidVal := r.Context().Value("user_id")
	if uidVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := uidVal.(int)

	// Fetch this user's stripe_customer_id
	var stripeCustomerID sql.NullString
	err := db.QueryRow(
		`SELECT stripe_customer_id FROM users WHERE id = $1;`,
		userID,
	).Scan(&stripeCustomerID)
	if err != nil {
		http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
		return
	}
	if !stripeCustomerID.Valid || stripeCustomerID.String == "" {
		http.Error(w, "No Stripe customer on file. Add a credit card first.", http.StatusBadRequest)
		return
	}
	custID := stripeCustomerID.String

	// Calculate total amount in cents
	var totalAmount int64 = 0
	type lineItem struct {
		ProductID int
		Quantity  int
		Subtotal  int64
	}
	var lineItems []lineItem

	for _, it := range req.Items {
		if it.Quantity <= 0 {
			http.Error(w, "Quantity must be > 0", http.StatusBadRequest)
			return
		}
		// Fetch price_cents for each product
		var priceCents int
		err := db.QueryRow(
			`SELECT price_cents FROM products WHERE id = $1;`,
			it.ProductID,
		).Scan(&priceCents)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found: "+strconv.Itoa(it.ProductID), http.StatusBadRequest)
				return
			}
			http.Error(w, "Failed to fetch product price", http.StatusInternalServerError)
			return
		}
		itemTotal := int64(priceCents) * int64(it.Quantity)
		totalAmount += itemTotal

		lineItems = append(lineItems, lineItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			Subtotal:  itemTotal,
		})
	}

	// Create a Stripe PaymentIntent restricted to "card"
	piParams := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(totalAmount),
		Currency:           stripe.String(string(stripe.CurrencyCAD)),
		Customer:           stripe.String(custID),
		PaymentMethod:      stripe.String(req.PaymentMethodID),
		Confirm:            stripe.Bool(true),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}), // ← restrict to card
	}
	pi, err := paymentintent.New(piParams)
	if err != nil {
		http.Error(w, "Stripe payment failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Record each line item in purchases within a DB transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Server error (begin tx)", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, li := range lineItems {
		_, err := tx.Exec(
			`INSERT INTO purchases
         (user_id, product_id, quantity, total_price_cents, stripe_payment_intent_id)
       VALUES ($1, $2, $3, $4, $5);`,
			userID, li.ProductID, li.Quantity, li.Subtotal, pi.ID,
		)
		if err != nil {
			http.Error(w, "Failed to record purchase", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Server error (commit tx)", http.StatusInternalServerError)
		return
	}

	// Return success JSON
	resp := buyResponse{
		Success:             true,
		StripePaymentIntent: pi.ID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
