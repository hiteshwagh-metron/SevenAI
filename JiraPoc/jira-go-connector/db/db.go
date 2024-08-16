package db

import (
    "database/sql"
    "fmt"
    "log"
    "jira-go-connector/config"
    _ "github.com/lib/pq"
)

var DB *sql.DB

func Init() {
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)
    var err error
    DB, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Error connecting to the database:", err)
    }
}
