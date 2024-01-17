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

const (
	host     = "host_name"
	port     = 5432
	user     = "user_name"
	password = "password"
	dbname   = "db_name"
)

var db *sql.DB

// Baby model
type Baby struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Meaning  string `json:"meaning"`
	Language string `json:"language"`
}

func main() {
	// PostgreSQL connection
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Check success of db connection
	err = checkDBConnection()
	if err != nil {
		log.Fatal("Db connection failed", err)
	}

	fmt.Println("Db conection is successful")

	defer db.Close()

	// create router
	router := mux.NewRouter()

	// Endpoints
	router.HandleFunc("/babys/get-all", getAllBabys).Methods("GET")
	router.HandleFunc("/babys/get/{id}", getBaby).Methods("GET")
	router.HandleFunc("/babys/add", createBaby).Methods("POST")
	router.HandleFunc("/babys/update/{id}", updateBaby).Methods("PUT")
	router.HandleFunc("/babys/delete/{id}", deleteBaby).Methods("DELETE")

	// start server
	fmt.Println("Server : Running on port 8080")
	http.ListenAndServe(":8080", router)
}

// get all baby names
func getAllBabys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.Query("SELECT * FROM baby_names")
	if err != nil {
		http.Error(w, "Cannot get data from db", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var babys []Baby
	for rows.Next() {
		var baby Baby
		err := rows.Scan(&baby.ID, &baby.Name, &baby.Meaning, &baby.Language)
		if err != nil {
			http.Error(w, "Cannot get data from db", http.StatusInternalServerError)
			return
		}
		babys = append(babys, baby)
	}

	json.NewEncoder(w).Encode(babys)
}

// Get one baby name
func getBaby(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var baby Baby
	err = db.QueryRow("SELECT * FROM baby_names WHERE id = $1", id).Scan(&baby.ID, &baby.Name, &baby.Meaning, &baby.Language)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	json.NewEncoder(w).Encode(baby)
}

// Create baby name
func createBaby(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var baby Baby
	if err := json.NewDecoder(r.Body).Decode(&baby); err != nil {
		http.Error(w, "Cannot decode JSON", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO baby_names (name, meaning, language) VALUES ($1, $2, $3)", baby.Name, baby.Meaning, baby.Language)
	fmt.Println(err)
	if err != nil {
		http.Error(w, "Cannot get data from db", http.StatusInternalServerError)
		return
	}

	// Başarılı bir şekilde eklendiyse, eklenen bebek bilgisini JSON formatında yanıtla
	w.WriteHeader(http.StatusCreated) // 201 Created yanıt kodu
	json.NewEncoder(w).Encode(baby)
}

// Update baby names
func updateBaby(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedBaby Baby
	_ = json.NewDecoder(r.Body).Decode(&updatedBaby)

	_, err = db.Exec("UPDATE baby_names SET name=$1, meaning=$2, language=$3 WHERE id=$4", updatedBaby.Name, updatedBaby.Meaning, updatedBaby.Language, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	json.NewEncoder(w).Encode(updatedBaby)
}

// Remove baby names
func deleteBaby(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM baby_names WHERE id=$1", id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	json.NewEncoder(w).Encode(fmt.Sprintf("Baby ID %d is deleted", id))
}

// check DB Connection
func checkDBConnection() error {
	err := db.Ping()
	if err != nil {
		return fmt.Errorf("Db connection failed %v", err)
	}
	return nil
}
