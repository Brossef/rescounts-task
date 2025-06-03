package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/gorilla/mux" // Registering the driver
)

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int    `json:"price_cents"`
}

func listProductsHandler(w http.ResponseWriter, r *http.Request) {
	// Query the DB
	rows, err := db.Query(`
    SELECT id, name, description, price_cents
    FROM products
    ORDER BY id
  `)
	if err != nil {
		http.Error(w, "Failed to query products", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Build a slice of Product
	products := make([]Product, 0)

	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PriceCents); err != nil {
			http.Error(w, "Error scanning product", http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating products", http.StatusInternalServerError)
		return
	}

	// Return as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func createProductHandler(w http.ResponseWriter, r *http.Request) {
	// Decode JSON body into a Product struct
	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		PriceCents  int    `json:"price_cents"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Basic validation
	if payload.Name == "" || payload.PriceCents <= 0 {
		http.Error(w, "name and price_cents are required (price_cents > 0)", http.StatusBadRequest)
		return
	}

	// Insert into products
	var newID int
	err := db.QueryRow(
		`INSERT INTO products (name, description, price_cents)
      VALUES ($1, $2, $3)
      RETURNING id;`,
		payload.Name, payload.Description, payload.PriceCents,
	).Scan(&newID)
	if err != nil {
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	// Return the created product (including new ID)
	newProduct := Product{
		ID:          newID,
		Name:        payload.Name,
		Description: payload.Description,
		PriceCents:  payload.PriceCents,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProduct)
}

func updateProductHandler(w http.ResponseWriter, r *http.Request) {
	// Extract {id} from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	prodID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Decode JSON payload (same fields as create)
	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		PriceCents  int    `json:"price_cents"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Basic validation
	if payload.Name == "" || payload.PriceCents <= 0 {
		http.Error(w, "name and price_cents are required (price_cents > 0)", http.StatusBadRequest)
		return
	}

	// UPDATE Query
	res, err := db.Exec(
		`UPDATE products
       SET name = $1, description = $2, price_cents = $3
     WHERE id = $4;`,
		payload.Name, payload.Description, payload.PriceCents, prodID,
	)
	if err != nil {
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	// Return 200 OK with the updated product
	updated := Product{
		ID:          prodID,
		Name:        payload.Name,
		Description: payload.Description,
		PriceCents:  payload.PriceCents,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	// Extract {id} from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	prodID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Run DELETE Query
	res, err := db.Exec(`DELETE FROM products WHERE id = $1;`, prodID)
	if err != nil {
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
