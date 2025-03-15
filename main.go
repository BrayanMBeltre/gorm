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

type NetflixShow struct {
	ShowID      string    `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Director    *string   `json:"director"`
	CastMembers *string   `json:"cast_members"`
	Country     *string   `json:"country"`
	DateAdded   time.Time `json:"date_added"`
	ReleaseYear int       `json:"release_year"`
	Rating      *string   `json:"rating"`
	Duration    *string   `json:"duration"`
	ListedIn    *string   `json:"listed_in"`
	Description *string   `json:"description"`
}

type Meta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
}

type PaginatedResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

var db *gorm.DB

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Error getting database instance: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
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
	per_page := parseIntQuery(r, "per_page", 10)
	page := parseIntQuery(r, "page", 1)

	var shows []NetflixShow
	if err := db.Offset((page - 1) * per_page).Find(&shows).Error; err != nil {
		http.Error(w, "Error fetching shows", http.StatusInternalServerError)
		log.Printf("Error fetching shows: %v",
			err)
		return
	}

	var total int64
	if err := db.Model(&NetflixShow{}).Count(&total).Error; err != nil {
		http.Error(w, "Error counting shows", http.StatusInternalServerError)
		log.Printf("Error counting shows: %v", err)
		return
	}

	response := PaginatedResponse[NetflixShow]{
		Data: shows,
		Meta: Meta{
			CurrentPage: page,
			PerPage:     per_page,
			Total:       total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
