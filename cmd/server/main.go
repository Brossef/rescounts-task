package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/gorilla/mux" //registering the driver
	"github.com/stripe/stripe-go/v82"
)

var db *sql.DB

func main() {

	dsn := "host=" + os.Getenv("DB_HOST") +
		" port=" + os.Getenv("DB_PORT") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASSWORD") +
		" dbname=" + os.Getenv("DB_NAME") +
		" sslmode=disable"
	// dsn := "host=db port=5432 user=rescounts_user password=rescounts_pass dbname=rescounts_db sslmode=disable"
	log.Printf("DSN: %q\n", dsn)
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("cannot open database:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("cannot connect to database:", err)
	}
	defer db.Close()

	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		log.Fatal("STRIPE_SECRET_KEY is not set")
	}
	stripe.Key = stripeKey

	r := mux.NewRouter()

	r.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello world!"))
	}).Methods("GET")

	r.HandleFunc("/signup", signupHandler).Methods("POST")
	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.Handle("/products", jwtMiddleware(http.HandlerFunc(listProductsHandler))).Methods("GET")

	// Admin-only routes (through JWT -> adminMiddleware):
	r.Handle(
		"/admin/products",
		jwtMiddleware(adminMiddleware(http.HandlerFunc(createProductHandler))),
	).Methods("POST")

	r.Handle(
		"/admin/products/{id}",
		jwtMiddleware(adminMiddleware(http.HandlerFunc(updateProductHandler))),
	).Methods("PUT")

	r.Handle(
		"/admin/products/{id}",
		jwtMiddleware(adminMiddleware(http.HandlerFunc(deleteProductHandler))),
	).Methods("DELETE")

	r.Handle(
		"/users/buy",
		jwtMiddleware(http.HandlerFunc(buyProductsHandler)),
	).Methods("POST")

	r.Handle(
		"/users/history",
		jwtMiddleware(http.HandlerFunc(getUserHistoryHandler)),
	).Methods("GET")

	r.Handle(
		"/users/creditcards",
		jwtMiddleware(http.HandlerFunc(addCreditCardHandler)),
	).Methods("POST")

	r.Handle(
		"/users/creditcards/{card_id}",
		jwtMiddleware(http.HandlerFunc(deleteCreditCardHandler)),
	).Methods("DELETE")

	addr := ":8080"
	log.Printf("Listening on %s…\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

// func init() {
// 	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
//   }

//   // Create a customer:
//   func createStripeCustomer(email string) (*stripe.Customer, error) {
// 	params := &stripe.CustomerParams{
// 	  Email: stripe.String(email),
// 	}
// 	return customer.New(params)
//   }
