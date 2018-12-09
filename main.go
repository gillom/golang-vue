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
var database *sql.DB

const (
	DB_USER     = "dbuser"
	DB_PASSWORD = "dbuser"
	DB_NAME     = "testing"
)

func initDb() {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
	DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	database = db
	checkErr(err)
}

func main() {

	// Database setup
	initDb()
	defer database.Close()	// executed at end of execution of the current function (this case main() wont ever end so database connection will stay alive)

	people = append(people, model.Person{ID: "1", Firstname: "John", Lastname: "Doe", Address: &model.Address{City: "City X", State: "State X"}})
	people = append(people, model.Person{ID: "2", Firstname: "Koko", Lastname: "Doe", Address: &model.Address{City: "City Z", State: "State Y"}})
	people = append(people, model.Person{ID: "3", Firstname: "Francis", Lastname: "Sunday"})

	router := mux.NewRouter()
	router.Use(commonMiddleware)

	// Define function handlers for each endpoint
	router.HandleFunc("/people", GetPeople).Methods("GET")
	router.HandleFunc("/people/{id}", GetPerson).Methods("GET")
	router.HandleFunc("/people", CreatePerson).Methods("POST")
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

func GetPeople(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(people) // return all people in json format
}

func GetPerson(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	personId := params["id"]
	
	var person model.Person
	sqlStatement := `SELECT * FROM person WHERE id=$1`
	row := database.QueryRow(sqlStatement, personId)
	err := row.Scan(&person.ID, &person.Firstname, &person.Lastname, &person.Age)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Zero rows found")
		} else {
			panic(err)
		}
	}

	json.NewEncoder(w).Encode(person)
}

func CreatePerson(w http.ResponseWriter, r *http.Request) {

	// Parse request body
	decoder := json.NewDecoder(r.Body)
    var person model.Person
    err := decoder.Decode(&person)
	checkErr(err)

	// Insert query
	var lastInsertId int
	err = database.QueryRow(
		`INSERT INTO person(firstname,lastname, age) 
		VALUES($1,$2,$3) returning id;`, 
		person.Firstname, person.Lastname, person.Age).Scan(&lastInsertId)
	checkErr(err)
	fmt.Println("last inserted id =", lastInsertId)

	// Return success response
	result := createRawJsonFromString(`{"sucess":true}`)
	json.NewEncoder(w).Encode(result)
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

// https://stackoverflow.com/questions/40429296/converting-string-to-json-or-struct-in-golang
func createRawJsonFromString(rawJson string) map[string]interface{} {
    in := []byte(rawJson)
    var raw map[string]interface{}
    json.Unmarshal(in, &raw)
	return raw
}