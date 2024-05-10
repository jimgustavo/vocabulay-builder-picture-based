package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Answer struct {
	Option string `json:"option"`
	URL    string `json:"url"`
}

type Item struct {
	ID         int      `json:"id"`
	Category   string   `json:"category"`
	Question   string   `json:"question"`
	TargetWord string   `json:"targetWord"`
	Answers    []Answer `json:"answers"`
	Correct    int      `json:"correct"`
}

type Items []Item

var db *sql.DB

func main() {
	rand.Seed(time.Now().UnixNano())
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
	r.HandleFunc("/dataset/batch", createDataSetBatch).Methods("POST")
	r.HandleFunc("/dataset/{id}/duplicate", duplicateDataSetByID).Methods("POST")
	r.HandleFunc("/dataset/{id}", updateDataSet).Methods("PUT")
	r.HandleFunc("/dataset/{id}", deleteDataSet).Methods("DELETE")
	r.HandleFunc("/dataset/{id}/scramble", scrambleAnswersByID).Methods("POST")

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
	var items Items
	for rows.Next() {
		var item Item
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
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the dataset in the JSON format
	json.NewEncoder(w).Encode(items)
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
	var item Item
	var answersJSON []byte
	query := "SELECT category, question, targetWord, answers, correct FROM data_set WHERE id = $1"
	err = db.QueryRow(query, id).Scan(&item.Category, &item.Question, &item.TargetWord, &answersJSON, &item.Correct)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Dataset item not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if err := json.Unmarshal(answersJSON, &item.Answers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the dataset item in the JSON format
	json.NewEncoder(w).Encode(item)
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
	var items Items
	for rows.Next() {
		var item Item
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
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d dataset items for category: %s\n", len(items), category)

	// Respond with the dataset in the JSON format
	json.NewEncoder(w).Encode(items)
}

func createDataSet(w http.ResponseWriter, r *http.Request) {
	// Decode the request body into an Item struct
	var item Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Marshal the Answers field to JSON
	answersJSON, err := json.Marshal(item.Answers)
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
		item.Category,
		item.Question,
		item.TargetWord,
		answersJSON,
		item.Correct,
	).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the ID of the newly created dataset item
	response := map[string]int{"id": id}
	json.NewEncoder(w).Encode(response)
}

func createDataSetBatch(w http.ResponseWriter, r *http.Request) {
	// Decode the request body into a slice of Item structs
	var items []Item
	err := json.NewDecoder(r.Body).Decode(&items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Iterate over each Item object in the slice and insert into the database
	for _, item := range items {
		// Marshal the Answers field to JSON
		answersJSON, err := json.Marshal(item.Answers)
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
			item.Category,
			item.Question,
			item.TargetWord,
			answersJSON,
			item.Correct,
		).Scan(&id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Log the ID of the newly created dataset item
		log.Printf("Created dataset item with ID: %d\n", id)
	}

	// Respond with success message
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Batch of dataset items created successfully")
}

func duplicateDataSetByID(w http.ResponseWriter, r *http.Request) {
	// Extract the dataset ID from the request URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid dataset ID", http.StatusBadRequest)
		return
	}

	// Query the dataset item from the database by ID
	var dataset Item
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

	// Insert the duplicated dataset item into the database
	query = `
        INSERT INTO data_set (category, question, targetWord, answers, correct)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
	var newID int
	err = db.QueryRow(
		query,
		dataset.Category,
		dataset.Question,
		dataset.TargetWord,
		answersJSON,
		dataset.Correct,
	).Scan(&newID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the ID of the newly duplicated dataset item
	response := map[string]int{"id": newID}
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

	// Decode the request body into a Item struct
	var dataset Item
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

func scrambleAnswersByID(w http.ResponseWriter, r *http.Request) {
	// Extract the dataset ID from the request URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid dataset ID", http.StatusBadRequest)
		return
	}

	// Query the dataset item from the database by ID
	var answersJSON []byte
	query := "SELECT answers FROM data_set WHERE id = $1"
	err = db.QueryRow(query, id).Scan(&answersJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Dataset item not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Log the retrieved answers JSON
	log.Printf("Retrieved answers JSON for dataset item with ID %d: %s", id, string(answersJSON))

	// Unmarshal the answers JSON into a slice of Answer
	var answers []Answer
	if err := json.Unmarshal(answersJSON, &answers); err != nil {
		// Log the error
		log.Printf("Error unmarshaling answers JSON: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log the unmarshaled answers
	log.Printf("Unmarshaled answers: %+v", answers)

	// Scramble the answers
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(answers), func(i, j int) {
		answers[i], answers[j] = answers[j], answers[i]
	})

	// Marshal the scrambled answers back to JSON
	scrambledAnswersJSON, err := json.Marshal(answers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the dataset item in the database with the scrambled answers
	updateQuery := "UPDATE data_set SET answers = $1 WHERE id = $2"
	_, err = db.Exec(updateQuery, scrambledAnswersJSON, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with a success message
	fmt.Fprintf(w, "Dataset item with ID %d answers scrambled successfully", id)
}

func (item *Item) ScrambleAnswers() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(item.Answers), func(i, j int) {
		item.Answers[i], item.Answers[j] = item.Answers[j], item.Answers[i]
	})
}

func (items Items) ScrambleAllAnswers() {
	for i := range items {
		items[i].ScrambleAnswers()
	}
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

curl -X POST -H "Content-Type: application/json" -d '[
  {
    "category": "flinstones-present-continuous",
    "question": "Wilma and Betty are playing cards",
    "targetWord": "are playing ",
    "answers": [
      {
        "option": "sitting-on-the-beach",
        "url": "https://i.pinimg.com/736x/53/10/19/5310193ee11d9444e3597f481d18d200.jpg"
      },
      {
        "option": "go-shopping",
        "url": "https://static1.srcdn.com/wordpress/wp-content/uploads/2019/02/flintstonesWilma.jpg"
      },
      {
        "option": "washing-the-dishes",
        "url": "https://www.telegraph.co.uk/multimedia/archive/01728/eletap_1728744a.gif"
      },
      {
        "option": "playing-cards",
        "url": "https://i.ytimg.com/vi/R6Q7K90vb0I/hqdefault.jpg"
      }
    ],
    "correct": 1
  },
  {
    "category": "flinstones-present-continuous",
    "question": "Wilma and Betty are sitting on the beach",
    "targetWord": "are sitting",
    "answers": [
      {
        "option": "sitting-on-the-beach",
        "url": "https://i.pinimg.com/736x/53/10/19/5310193ee11d9444e3597f481d18d200.jpg"
      },
      {
        "option": "go-shopping",
        "url": "https://static1.srcdn.com/wordpress/wp-content/uploads/2019/02/flintstonesWilma.jpg"
      },
      {
        "option": "washing-the-dishes",
        "url": "https://www.telegraph.co.uk/multimedia/archive/01728/eletap_1728744a.gif"
      },
      {
        "option": "playing-cards",
        "url": "https://i.ytimg.com/vi/R6Q7K90vb0I/hqdefault.jpg"
      }
    ],
    "correct": 2
  }
]' http://localhost:8080/dataset/batch

curl -X POST http://localhost:8080/dataset/5/scramble

*/
