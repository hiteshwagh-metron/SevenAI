package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

var (
    ClientID     string
    ClientSecret string
    RedirectURI  string
    AuthURL      string
    TokenURL     string
    DBHost       string
    DBPort       string
    DBUser       string
    DBPassword   string
    DBName       string
)

func Load() {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file:", err)
    }
	log.Println("Load called")
    ClientID = os.Getenv("CLIENT_ID")
    ClientSecret = os.Getenv("CLIENT_SECRET")
    RedirectURI = os.Getenv("REDIRECT_URI")
    AuthURL = os.Getenv("AUTH_URL")
    TokenURL = os.Getenv("TOKEN_URL")
	
    DBHost = os.Getenv("DB_HOST")
    DBPort = os.Getenv("DB_PORT")
    DBUser = os.Getenv("DB_USER")
    DBPassword = os.Getenv("DB_PASSWORD")
    DBName = os.Getenv("DB_NAME")
}
