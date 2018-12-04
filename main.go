package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"rest-api/model"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var people []model.Person
// var db *sql.DB

const (
	DB_USER     = "dbuser"
	DB_PASSWORD = "dbuser"
	DB_NAME     = "testing"
)

func main() {



	people = append(people, model.Person{ID: "1", Firstname: "John", Lastname: "Doe", Address: &model.Address{City: "City X", State: "State X"}})
	people = append(people, model.Person{ID: "2", Firstname: "Koko", Lastname: "Doe", Address: &model.Address{City: "City Z", State: "State Y"}})
	people = append(people, model.Person{ID: "3", Firstname: "Francis", Lastname: "Sunday"})

	router := mux.NewRouter()
	router.Use(commonMiddleware)

	// Define function handlers for each endpoint
	router.HandleFunc("/people", GetPeople).Methods("GET")
	router.HandleFunc("/people/{id}", GetPerson).Methods("GET")
	router.HandleFunc("/people/{id}", CreatePerson).Methods("POST")
	router.HandleFunc("/people/{id}", DeletePerson).Methods("DELETE")

	// Serve app
	log.Fatal(http.ListenAndServe(":8080", router))
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// localhost:8080/people
func GetPeople(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(people) // return all people in json format
}

// localhost:8080/people/1
func GetPerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for _, item := range people {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
		}
	}
}

func CreatePerson(w http.ResponseWriter, r *http.Request) {

	// Open database
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
	DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	var lastInsertId int
	err = db.QueryRow("INSERT INTO person(id,firstname,lastname) VALUES($1,$2,$3) returning id;", 1, "james", "bond").Scan(&lastInsertId)
	checkErr(err)
	fmt.Println("last inserted id =", lastInsertId)
		
	params := mux.Vars(r)
	var person model.Person
	_ = json.NewDecoder(r.Body).Decode(&person)
	person.ID = params["id"]
	people = append(people, person)
	json.NewEncoder(w).Encode(people)
}

func DeletePerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for index, item := range people {
		if item.ID == params["id"] {
			people = append(people[:index], people[index+1:]...) // removing item from a slice
			break
		}
	}
	json.NewEncoder(w).Encode(people)
}


func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}