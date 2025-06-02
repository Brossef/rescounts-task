package main

import (
	"database/sql"
	"encoding/json"
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
	StripePaymentMthID string `json:"stripe_pm_id"`
	Brand              string `json:"brand"`
	Last4              string `json:"last4"`
	ExpMonth           int    `json:"exp_month"`
	ExpYear            int    `json:"exp_year"`
}

// addCreditCardHandler attaches a Stripe PaymentMethod to the user’s Stripe Customer,
// then saves metadata in our `credit_cards` table.
func addCreditCardHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Decode request JSON
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

	// 2. Extract logged-in user_id from context
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

	// 3. Fetch the user's email and existing stripe_customer_id
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

	// 4. If stripe_customer_id is empty, create a Stripe Customer and update users
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

	// 5. Attach the PaymentMethod to that Stripe Customer
	attachParams := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}
	pm, err := paymentmethod.Attach(
		req.PaymentMethodID,
		attachParams,
	)
	if err != nil {
		// Stripe might return an error if PM is invalid or already attached
		http.Error(w, "Failed to attach payment method: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 6. Extract card details from the returned PaymentMethod object
	//    (We know it’s a card PaymentMethod because we passed pm_…)
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

	// 7. Insert into credit_cards table
	var newID int
	err = db.QueryRow(
		`INSERT INTO credit_cards 
       (user_id, stripe_pm_id, brand, last4, exp_month, exp_year)
     VALUES ($1, $2, $3, $4, $5, $6)
     RETURNING id;`,
		userID, pm.ID, brand, last4, expMonth, expYear,
	).Scan(&newID)
	if err != nil {
		http.Error(w, "Failed to save credit card", http.StatusInternalServerError)
		return
	}

	// 8. Return the new credit card record as JSON
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
	// 1. Extract logged-in user_id from context
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

	// 2. Parse {card_id} from URL
	vars := mux.Vars(r)
	cardIDStr := vars["card_id"]
	cardID, err := strconv.Atoi(cardIDStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	// 3. Look up that credit_card row to ensure it belongs to this user, and get the Stripe PM ID
	var stripePMID string
	row := db.QueryRow(
		`SELECT stripe_pm_id 
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

	// 4. Detach the PaymentMethod in Stripe
	_, err = paymentmethod.Detach(stripePMID, nil)
	if err != nil {
		// It’s possible this PM was already detached; you could choose to ignore certain errors,
		// but for now return an error if Stripe says so.
		http.Error(w, "Failed to detach payment method: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 5. Delete the row from credit_cards
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
		// This should not happen, since we already SELECTed it above. But just in case:
		http.Error(w, "Credit card not found", http.StatusNotFound)
		return
	}

	// 6. Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
