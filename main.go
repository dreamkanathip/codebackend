package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Appointment struct {
	gorm.Model
	DateTime string `json:"dateTime"`
	Doctor   string `json:"doctor"`
	Name     string `json:"name"`
	Problem  string `json:"problem"`
}

var db *gorm.DB

func DB() *gorm.DB {
	return db
}

func initDB() {
	database, err := gorm.Open(sqlite.Open("sa-66.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	database.AutoMigrate(&Appointment{})
	db = database
}

func main() {
	initDB()

	corsOptions := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Replace with the origin of your frontend application
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	r := mux.NewRouter()
	r.HandleFunc("/appointment", createAppointment).Methods("POST")
	r.HandleFunc("/appointments", getAppointments).Methods("GET")

	// Wrap the router with CORS middleware
	handler := corsOptions.Handler(r)
	http.Handle("/", handler)

	serverAddr := ":8080"
	fmt.Printf("Server is listening on %s...\n", serverAddr)

	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createAppointment(w http.ResponseWriter, r *http.Request) {
	var appointment Appointment
	err := json.NewDecoder(r.Body).Decode(&appointment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Bad Request: %v", err)
		return
	}

	// Create a new appointment record
	result := DB().Create(&appointment)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		log.Printf("Failed to create appointment: %v", result.Error)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("Appointment created successfully")
}

func getAppointments(w http.ResponseWriter, r *http.Request) {
	var appointments []Appointment
	result := DB().Find(&appointments)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		log.Printf("Failed to fetch appointments: %v", result.Error)
		return
	}

	// Serialize the appointments to JSON and send as response
	jsonBytes, err := json.Marshal(appointments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Failed to serialize appointments: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)

	log.Printf("Appointments fetched successfully")
}
