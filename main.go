package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

// Define a struct to represent a single dataset item
type DataSet struct {
	ID         int      `json:"id"`
	Category   string   `json:"category"`
	Question   string   `json:"question"`
	TargetWord string   `json:"targetWord"`
	Picture    string   `json:"picture"`
	Answers    []string `json:"answers"`
	Correct    int      `json:"correct"`
}

var db *sql.DB

func main() {
	// Initialize database connection
	var err error
	db, err = sql.Open("postgres", "postgres://tavito:mamacita@localhost:5432/data_set?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Initialize Gorilla Mux router
	r := mux.NewRouter()

	// Define API endpoints
	r.HandleFunc("/dataset", getDataSet).Methods("GET")
	r.HandleFunc("/dataset/{id}", getDataSetByID).Methods("GET")
	r.HandleFunc("/dataset/category/{category}", getDataSetByCategory).Methods("GET")
	r.HandleFunc("/dataset", createDataSet).Methods("POST")
	r.HandleFunc("/dataset/{id}", updateDataSet).Methods("PUT")
	r.HandleFunc("/dataset/{id}", deleteDataSet).Methods("DELETE")

	// Serve static files from the "static" directory
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Handler functions
func getDataSet(w http.ResponseWriter, r *http.Request) {
	// Query all dataset items from the database
	rows, err := db.Query("SELECT id, category, question, targetWord, picture, answers, correct FROM data_set")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Iterate over the rows and build the dataset slice
	var dataset []DataSet
	for rows.Next() {
		var item DataSet
		err := rows.Scan(&item.ID, &item.Category, &item.Question, &item.TargetWord, &item.Picture, pq.Array(&item.Answers), &item.Correct)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		dataset = append(dataset, item)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the dataset in the JSON format
	json.NewEncoder(w).Encode(dataset)
}

func getDataSetByID(w http.ResponseWriter, r *http.Request) {
	// Extract the dataset ID from the request URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid dataset ID", http.StatusBadRequest)
		return
	}

	// Query the dataset item from the database by ID
	var dataset DataSet
	query := "SELECT category, question, targetWord, picture, answers, correct FROM data_set WHERE id = $1"
	err = db.QueryRow(query, id).Scan(&dataset.Category, &dataset.Question, &dataset.TargetWord, &dataset.Picture, pq.Array(&dataset.Answers), &dataset.Correct)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Dataset item not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Respond with the dataset item in the JSON format
	json.NewEncoder(w).Encode(dataset)
}

func getDataSetByCategory(w http.ResponseWriter, r *http.Request) {
	// Extract the category from the request URL
	vars := mux.Vars(r)
	category := vars["category"]

	log.Printf("Received request to fetch dataset for category: %s\n", category)

	// Query dataset items from the database by category
	rows, err := db.Query("SELECT id, question, targetWord, picture, answers, correct FROM data_set WHERE category = $1", category)
	if err != nil {
		log.Printf("Error querying database: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Iterate over the rows and build the dataset slice
	var dataset []DataSet
	for rows.Next() {
		var item DataSet
		err := rows.Scan(&item.ID, &item.Question, &item.TargetWord, &item.Picture, pq.Array(&item.Answers), &item.Correct)
		if err != nil {
			log.Printf("Error scanning row: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		dataset = append(dataset, item)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d dataset items for category: %s\n", len(dataset), category)

	// Respond with the dataset in the JSON format
	json.NewEncoder(w).Encode(dataset)
}

func createDataSet(w http.ResponseWriter, r *http.Request) {
	// Decode the request body into a DataSet struct
	var dataset DataSet
	err := json.NewDecoder(r.Body).Decode(&dataset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert the dataset item into the database
	query := `
        INSERT INTO data_set (category, question, targetWord, picture, answers, correct )
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `
	var id int
	err = db.QueryRow(
		query,
		dataset.Category,
		dataset.Question,
		dataset.TargetWord,
		dataset.Picture,
		pq.Array(dataset.Answers),
		dataset.Correct,
	).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the ID of the newly created dataset item
	response := map[string]int{"id": id}
	json.NewEncoder(w).Encode(response)
}

func updateDataSet(w http.ResponseWriter, r *http.Request) {
	// Log that the updateDataSet handler function has been called
	log.Println("updateDataSet handler function called")

	// Extract the dataset ID from the request URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid dataset ID", http.StatusBadRequest)
		log.Println("Invalid dataset ID:", err)
		return
	}

	// Decode the request body into a DataSet struct
	var dataset DataSet
	err = json.NewDecoder(r.Body).Decode(&dataset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Error decoding JSON:", err)
		return
	}

	// Log the received dataset
	log.Printf("Received dataset: %+v", dataset)

	// Update the dataset item in the database
	query := `
        UPDATE data_set
        SET question = $1, targetWord = $2, picture = $3, answers = $4, correct = $5
        WHERE id = $6
    `
	_, err = db.Exec(
		query,
		dataset.Question,
		dataset.TargetWord,
		dataset.Picture,
		pq.Array(dataset.Answers),
		dataset.Correct,
		id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error updating dataset in database:", err)
		return
	}

	// Log that the dataset item has been updated successfully
	log.Println("Dataset item updated successfully")

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Dataset item updated successfully")
}

func deleteDataSet(w http.ResponseWriter, r *http.Request) {
	// Extract the dataset ID from the request URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid dataset ID", http.StatusBadRequest)
		return
	}

	// Delete the dataset item from the database
	_, err = db.Exec("DELETE FROM data_set WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Dataset item deleted successfully")
}

/*
///////////Postgres Database//////////
psql

\l

CREATE DATABASE data_set;

DROP DATABASE data_set;     //for deleting a database

\c data_set

pwd

\i /Users/tavito/Documents/go/vocabulary-builder-with-picture/data_set.sql

\dt


////////////////Curl Commands///////////////////

curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "id": 7,
    "category": "hard-word",
    "question": "Huawei have been hoarding parts in anticipation of a ban and have sought other suppliers",
    "targetWord": "hoarding",
    "picture": "https://www.shutterstock.com/shutterstock/photos/1363334219/display_1500/stock-photo-carlos-barbosa-rio-grande-do-sul-brasil-april-interior-of-auto-parts-store-1363334219.jpg",
    "answers": [
      "v. keeping for future",
      "v. wasting",
      "v. needing",
      "v. buying"
    ],
    "correct": 0
  }' \
  http://localhost:8080/dataset


curl -X PUT -H "Content-Type: application/json" -d '{
    "question": "Updated question",
    "targetWord": "updated";2B,
    "picture": "https://example.com/updated.jpg",
    "answers": ["updated answer 1", "updated answer 2", "updated answer 3", "updated answer 4"],
    "correct": 1
}' http://localhost:8080/dataset/6

curl -X PUT \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Huawei have been hoarding parts in anticipation of a ban and have sought other suppliers",
    "targetWord": "hoarding",
    "picture": "https://www.shutterstock.com/shutterstock/photos/1363334219/display_1500/stock-photo-carlos-barbosa-rio-grande-do-sul-brasil-april-interior-of-auto-parts-store-1363334219.jpg",
    "answers": [
      "v. keeping for future",
      "v. wasting",
      "v. needing",
      "v. buying"
    ],
    "correct": 0
  }' \
  http://localhost:8080/dataset/6



curl -X GET \
  'http://localhost:8080/dataset/easy-word'


curl -X DELETE http://localhost:8080/dataset/1


*/
