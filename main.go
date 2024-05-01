package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Define a struct to represent a single dataset item
type DataSet struct {
	ID         int               `json:"id"`
	Category   string            `json:"category"`
	Question   string            `json:"question"`
	TargetWord string            `json:"targetWord"`
	Answers    map[string]string `json:"answers"`
	Correct    int               `json:"correct"`
}

var db *sql.DB

func main() {
	// Initialize database connection
	var err error
	db, err = sql.Open("postgres", "postgres://tavito:mamacita@localhost:5432/data_set_pb?sslmode=disable")
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

func getDataSet(w http.ResponseWriter, r *http.Request) {
	// Query all dataset items from the database
	rows, err := db.Query("SELECT id, category, question, targetWord, answers, correct FROM data_set")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Iterate over the rows and build the dataset slice
	var dataset []DataSet
	for rows.Next() {
		var item DataSet
		var answersJSON []byte
		err := rows.Scan(&item.ID, &item.Category, &item.Question, &item.TargetWord, &answersJSON, &item.Correct)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(answersJSON, &item.Answers); err != nil {
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
	var answersJSON []byte
	query := "SELECT category, question, targetWord, answers, correct FROM data_set WHERE id = $1"
	err = db.QueryRow(query, id).Scan(&dataset.Category, &dataset.Question, &dataset.TargetWord, &answersJSON, &dataset.Correct)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Dataset item not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if err := json.Unmarshal(answersJSON, &dataset.Answers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	rows, err := db.Query("SELECT id, question, targetWord, answers, correct FROM data_set WHERE category = $1", category)
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
		var answersJSON []byte
		err := rows.Scan(&item.ID, &item.Question, &item.TargetWord, &answersJSON, &item.Correct)
		if err != nil {
			log.Printf("Error scanning row: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(answersJSON, &item.Answers); err != nil {
			log.Printf("Error unmarshaling answers JSON: %v\n", err)
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

	// Marshal the Answers field to JSON
	answersJSON, err := json.Marshal(dataset.Answers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert the dataset item into the database
	query := `
        INSERT INTO data_set (category, question, targetWord, answers, correct )
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
	var id int
	err = db.QueryRow(
		query,
		dataset.Category,
		dataset.Question,
		dataset.TargetWord,
		answersJSON,
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

	// Marshal the Answers field to JSON
	answersJSON, err := json.Marshal(dataset.Answers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log the received dataset
	log.Printf("Received dataset: %+v", dataset)

	// Update the dataset item in the database
	query := `
        UPDATE data_set
        SET category = $1, question = $2, targetWord = $3, answers = $4, correct = $5
        WHERE id = $6
    `
	_, err = db.Exec(
		query,
		dataset.Category,
		dataset.Question,
		dataset.TargetWord,
		answersJSON,
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

CREATE DATABASE data_set_pb;

DROP DATABASE data_set_pb;     //for deleting a database

\c data_set_pb

pwd

\i /Users/tavito/Documents/go/vocabulary-builder-picture-based/data_set_pb.sql

\dt


////////////////Curl Commands///////////////////
curl -X POST \
  http://localhost:8080/dataset \
  -H 'Content-Type: application/json' \
  -d '{
	"category": "easy-word",
	"question": "This farm yielded very well this year.",
	"targetWord": "yieldedssss",
	"answers": {
		"produced": "https://bloglatam.jacto.com/wp-content/uploads/2022/04/cultivo-de-tomate.jpg",
		"dry": "https://cdn.diariojornada.com.ar/imagenes/2021/12/25/555449_5659.jpg",
		"sell": "https://s3.envato.com/files/250543034/preview.jpg",
		"fire": "https://www.shutterstock.com/image-photo/wildfire-on-wheat-field-stubble-600nw-1916978717.jpg"
	},
	"correct": 0
}'

curl -X PUT \
  http://localhost:8080/dataset/4 \
  -H 'Content-Type: application/json' \
  -d '{
"id": 4,
"category": "hard-word",
"question": "This farm dry very well this year.",
"targetWord": "dry",
"answers": {
"dry": "https://cdn.diariojornada.com.ar/imagenes/2021/12/25/555449_5659.jpg",
"fire": "https://www.shutterstock.com/image-photo/wildfire-on-wheat-field-stubble-600nw-1916978717.jpg",
"produced": "https://bloglatam.jacto.com/wp-content/uploads/2022/04/cultivo-de-tomate.jpg",
"sell": "https://s3.envato.com/files/250543034/preview.jpg"
},
"correct": 0
}'

curl -X DELETE http://localhost:8080/dataset/5


*/
