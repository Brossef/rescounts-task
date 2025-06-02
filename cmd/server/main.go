package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/gorilla/mux" //registering the driver
)

var db *sql.DB

func main() {

	dsn := "host=" + os.Getenv("DB_HOST") +
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

	r.HandleFunc("/signup", signupHandler).Methods("POST")
	r.HandleFunc("/login", loginHandler).Methods("POST")

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
//   func createCustomer(email string) (*stripe.Customer, error) {
// 	params := &stripe.CustomerParams{
// 	  Email: stripe.String(email),
// 	}
// 	return customer.New(params)
//   }
