package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"

	"github.com/gorilla/mux"
	_ "github.com/gorilla/mux" //registering the driver

	"github.com/golang-jwt/jwt/v5"
)

// type Claims struct {
// 	UserID int `json:"user_id"`
// 	jwt.RegisteredClaims
// }

var db *sql.DB

func main() {

	dsn :=
		"host=" + os.Getenv("DB_HOST") +
			" port=" + os.Getenv("DB_PORT") +
			" user=" + os.Getenv("DB_USER") +
			" password=" + os.Getenv("DB_PASSWORD") +
			" dbname=" + os.Getenv("DB_NAME") +
			" sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error with sql db :(", err)
	}
	defer db.Close()
	// dsn := "host=db port=5432 user=rescounts_user password=rescounts_pass dbname=rescounts_db sslmode=disable"

	r := mux.NewRouter()

	r.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello world!"))
	}).Methods("GET")

	r.HandleFunc("/signup", SignUpHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")

	addr := ":8080"
	log.Printf("Listening on %s…\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func SignUpHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		var userID int
		err = db.QueryRow(`
			INSERT INTO users (username, email, password)
			VALUES ($1, $2, $3)
			RETURNING id
		`, input.Username, input.Email, string(hashedPassword)).Scan(&userID)

		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create JWT token
		claims := Claims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		secret := []byte(os.Getenv("JWT_SECRET"))
		signedToken, err := token.SignedString(secret)
		if err != nil {
			http.Error(w, "Failed to sign token", http.StatusInternalServerError)
			return
		}

		// Respond with token
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": signedToken,
		})
	}
}

func init() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

// Create a customer:
func createCustomer(email string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}
	return customer.New(params)
}
