package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type netflixShow struct {
	ShowID      string
	Type        string
	Title       string
	Director    *string
	CastMembers *string
	Country     *string
	DateAdded   *time.Time
	ReleaseYear int
	Rating      *string
	Duration    *string
	ListedIn    *string
	Description *string
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func dbConnect() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	return db
}

func parseIntQuery(r *http.Request, name string, defaultValue int) int {
	qp := r.URL.Query().Get(name)
	if qp == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(qp)
	if err != nil {
		log.Fatalf("Error parsing query parameter %s: %v", name, err)

		return defaultValue

	}

	return value
}

func handler(w http.ResponseWriter, r *http.Request) {
	limit := parseIntQuery(r, "limit", 10)
	page := parseIntQuery(r, "page", 1)

	db := dbConnect()

	var shows []netflixShow
	db.Limit(limit).Offset((page - 1) * limit).Find(&shows)

	json, err := json.Marshal(shows)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
