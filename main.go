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

	router := mux.NewRouter()
	router.Use(commonMiddleware)

	// Define function handlers for each endpoint
	router.HandleFunc("/person/{id}", GetPerson).Methods("GET")
	router.HandleFunc("/person", CreatePerson).Methods("POST")
	router.HandleFunc("/person/{id}", DeletePerson).Methods("DELETE")
	router.HandleFunc("/person/{id}", UpdatePerson).Methods("PATCH")

	// Serve app
	log.Fatal(http.ListenAndServe(":8080", router))
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
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
	result := createRawJsonFromString(fmt.Sprintf(`{"sucess":true, "id": %d}`, lastInsertId))
	json.NewEncoder(w).Encode(result)
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
			fmt.Println("Cannot read person - Zero rows found")
		} else {
			panic(err)
		}
	}

	json.NewEncoder(w).Encode(person)
}

func UpdatePerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	personId := params["id"]

	// Parse request body
	decoder := json.NewDecoder(r.Body)
	var person model.Person
	err := decoder.Decode(&person)
	checkErr(err)

	var lastInsertId int
	var personUpdated bool = true
	sqlStatement := `UPDATE person 
		SET firstname = $1,
			lastname = $2,
			age = $3 
		WHERE id=$4 returning id`
	err = database.QueryRow(sqlStatement, person.Firstname, person.Lastname, person.Age, personId).Scan(&lastInsertId)
	if err != nil {
		if err == sql.ErrNoRows {
			personUpdated = false
			fmt.Println("Cannot update person - zero rows found")
		} else {
			panic(err)
		}
	}

	result := createRawJsonFromString(fmt.Sprintf(`{"updated":%t, "id": %d}`, personUpdated, lastInsertId))
	json.NewEncoder(w).Encode(result)
}

func DeletePerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	personId := params["id"]

	var lastInsertId int
	var personDeleted bool = true
	sqlStatement := `DELETE FROM person WHERE id=$1 returning id`
	err := database.QueryRow(sqlStatement, personId).Scan(&lastInsertId)
	if err != nil {
		if err == sql.ErrNoRows {
			personDeleted = false
			fmt.Println("Cannot delete person - zero rows found")
		} else {
			panic(err)
		}
	}

	result := createRawJsonFromString(fmt.Sprintf(`{"deleted":%t, "id": %d}`, personDeleted, lastInsertId))
	json.NewEncoder(w).Encode(result)
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