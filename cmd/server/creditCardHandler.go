package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/paymentmethod"
)

// This struct mirrors the incoming JSON payload.
type addCCRequest struct {
	PaymentMethodID string `json:"payment_method_id"`
}

// This struct is what we return after inserting into credit_cards.
type creditCardResponse struct {
	ID                 int    `json:"id"`
	StripePaymentMthID string `json:"stripe_payment_method_id"`
	Brand              string `json:"brand"`
	Last4              string `json:"last4"`
	ExpMonth           int    `json:"exp_month"`
	ExpYear            int    `json:"exp_year"`
}

// addCreditCardHandler attaches a Stripe PaymentMethod to the user’s Stripe Customer,
// then saves metadata in our `credit_cards` table.
func addCreditCardHandler(w http.ResponseWriter, r *http.Request) {
	// Decode request JSON
	var req addCCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.PaymentMethodID == "" {
		http.Error(w, "payment_method_id is required", http.StatusBadRequest)
		return
	}

	// Extract logged-in user_id from context
	ctx := r.Context()
	uidVal := ctx.Value("user_id")
	if uidVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := uidVal.(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch the user's email and existing stripe_customer_id
	var stripeCustID sql.NullString
	var userEmail string
	row := db.QueryRow(
		`SELECT email, stripe_customer_id FROM users WHERE id = $1;`,
		userID,
	)
	if err := row.Scan(&userEmail, &stripeCustID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// If stripe_customer_id is empty, create a Stripe Customer and update users
	customerID := stripeCustID.String
	if !stripeCustID.Valid || stripeCustID.String == "" {
		// Create a new Stripe Customer
		custParams := &stripe.CustomerParams{
			Email: stripe.String(userEmail),
		}
		cust, err := customer.New(custParams)
		if err != nil {
			http.Error(w, "Failed to create Stripe customer", http.StatusInternalServerError)
			return
		}
		customerID = cust.ID

		// Update users.stripe_customer_id
		_, err = db.Exec(
			`UPDATE users SET stripe_customer_id = $1 WHERE id = $2;`,
			customerID, userID,
		)
		if err != nil {
			http.Error(w, "Server error updating stripe_customer_id", http.StatusInternalServerError)
			return
		}
	}

	// Attach the PaymentMethod to that Stripe Customer
	attachParams := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}
	pm, err := paymentmethod.Attach(
		req.PaymentMethodID,
		attachParams,
	)
	if err != nil {
		// if PM is invalid or already attached
		http.Error(w, "Failed to attach payment method: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract card details from the returned PaymentMethod object
	card := pm.Card
	brand := ""
	last4 := ""
	expMonth := 0
	expYear := 0
	if card != nil {
		brand = string(card.Brand)
		last4 = card.Last4
		expMonth = int(card.ExpMonth)
		expYear = int(card.ExpYear)
	}

	// Insert into credit_cards table
	var newID int
	insertQuery := `
  INSERT INTO credit_cards
    (user_id, stripe_payment_method_id, brand, last4, exp_month, exp_year)
  VALUES ($1, $2, $3, $4, $5, $6)
  RETURNING id;
  `
	err = db.QueryRow(
		insertQuery,
		userID, pm.ID, brand, last4, expMonth, expYear,
	).Scan(&newID)
	if err != nil {
		// Log the full error and the attempted query values
		log.Printf(
			"ERROR inserting credit_card row: %v\n   Query: %s\n   Values: userID=%d, pm.ID=%q, brand=%q, last4=%q, expMonth=%d, expYear=%d\n",
			err, insertQuery, userID, pm.ID, brand, last4, expMonth, expYear,
		)
		http.Error(w, "Failed to save credit card", http.StatusInternalServerError)
		return
	}

	// Return the new credit card record as JSON
	resp := creditCardResponse{
		ID:                 newID,
		StripePaymentMthID: pm.ID,
		Brand:              brand,
		Last4:              last4,
		ExpMonth:           expMonth,
		ExpYear:            expYear,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func deleteCreditCardHandler(w http.ResponseWriter, r *http.Request) {
	// Extract logged-in user_id from context
	ctx := r.Context()
	uidVal := ctx.Value("user_id")
	if uidVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := uidVal.(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse {card_id} from URL
	vars := mux.Vars(r)
	cardIDStr := vars["card_id"]
	cardID, err := strconv.Atoi(cardIDStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	// Look up that credit_card row to ensure it belongs to this user, and get the Stripe PM ID
	var stripePMID string
	row := db.QueryRow(
		`SELECT stripe_payment_method_id 
       FROM credit_cards 
      WHERE id = $1 AND user_id = $2;`,
		cardID, userID,
	)
	if err := row.Scan(&stripePMID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Credit card not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Detach the PaymentMethod in Stripe
	_, err = paymentmethod.Detach(stripePMID, nil)
	if err != nil {
		// It’s possible this PM was already detached; you could choose to ignore certain errors,
		// but for now return an error if Stripe says so.
		http.Error(w, "Failed to detach payment method: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Delete the row from credit_cards
	res, err := db.Exec(
		`DELETE FROM credit_cards WHERE id = $1;`,
		cardID,
	)
	if err != nil {
		http.Error(w, "Server error deleting credit card", http.StatusInternalServerError)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		// Just in case
		http.Error(w, "Credit card not found", http.StatusNotFound)
		return
	}

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
