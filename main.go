package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/avearmin/gorage-sale/internal/database"
	"github.com/avearmin/gorage-sale/internal/oauth2"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB               *database.Queries
	JwtSecret        string
	StateStore       *oauth2.StateStore
	ClientID         string
	ClientSecret     string
	OAuthRedirectURL string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Port has not been specified.")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT Secret has not been specified.")
	}

	clientID := os.Getenv("CLIENT_ID")
	if clientID == "" {
		log.Fatal("Client ID has not been specified.")
	}

	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatal("Client Secret has not been specified.")
	}

	oauthRedirectURL := ""

	config := apiConfig{
		JwtSecret:        jwtSecret,
		ClientID:         clientID,
		ClientSecret:     clientSecret,
		OAuthRedirectURL: oauthRedirectURL,
	}

	dbConn := os.Getenv("DB_CONN_STRING")
	if dbConn == "" {
		log.Println("DB connection string has not been specified. Running with limited endpoints.")
	} else {
		db, err := sql.Open("postgres", dbConn)
		if err != nil {
			log.Fatal(err)
		}
		dbQueries := database.New(db)
		config.DB = dbQueries
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/status", handleStatus)

	if config.DB != nil {
		mux.HandleFunc("/v1/users", config.handleUsers)
		mux.HandleFunc("/v1/items", config.handleItems)
	}

	corsMux := middlewareCors(mux)

	srv := http.Server{
		Addr:         ":" + port,
		Handler:      corsMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
