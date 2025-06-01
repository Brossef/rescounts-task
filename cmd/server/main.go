package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/gorilla/mux" //registering the driver
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello world!"))
	}).Methods("GET")

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))

	dsn := "host=db port=5432 user=rescounts_user password=rescounts_pass dbname=rescounts_db sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error with sql db :(")
	}
	defer db.Close()
}
