package main

import (
    "context"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"

    "github.com/joho/godotenv"
    "golang.org/x/oauth2"
)

var (
    oauth2Config *oauth2.Config
    tokenStorage = make(map[string]*oauth2.Token)
)

func init() {
    // Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        log.Println("Error loading .env file:", err)
    } else {
        log.Println(".env file loaded successfully")
    }

    oauth2Config = &oauth2.Config{
        ClientID:     os.Getenv("JIRA_CLIENT_ID"),
        ClientSecret: os.Getenv("JIRA_CLIENT_SECRET"),
        RedirectURL:  "http://localhost:8080/callback",
        Endpoint: oauth2.Endpoint{
            AuthURL:  "https://auth.atlassian.com/authorize",
            TokenURL: "https://auth.atlassian.com/oauth/token",
        },
        Scopes: []string{"read:jira-work", "read:jira-user"},
    }

    log.Println("OAuth2 config initialized")
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
    log.Println("HandleLogin called")
    url := oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
    log.Println("Redirecting to Jira login page:", url)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleCallback(w http.ResponseWriter, r *http.Request) {
    log.Println("HandleCallback called")
    code := r.URL.Query().Get("code")
    log.Println("Authorization code received:", code)

    token, err := oauth2Config.Exchange(context.Background(), code)
    if err != nil {
        log.Println("Error exchanging code for token:", err)
        http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Store token for the user (for simplicity, we are using a map, but consider using a database)
    tokenStorage["user-id"] = token

    log.Println("Token stored for user-id")
    fmt.Fprintf(w, "Jira connected successfully!")
}

func FetchData(w http.ResponseWriter, r *http.Request) {
    log.Println("FetchData called")
    token := tokenStorage["user-id"]
    if token == nil {
        log.Println("No token found for user-id")
        http.Error(w, "No token found", http.StatusUnauthorized)
        return
    }

    jiraDomain := "hiteshwagh9383.atlassian.net" // Replace with the Jira domain or fetch dynamically
    apiURL := fmt.Sprintf("https://%s/rest/api/2/project", jiraDomain)
    log.Println("Fetching data from Jira API:", apiURL)

    client := oauth2Config.Client(context.Background(), token)
    resp, err := client.Get(apiURL)
    if err != nil {
        log.Println("Error fetching data from Jira API:", err)
        http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println("Error reading Jira API response:", err)
        http.Error(w, "Failed to read response: "+err.Error(), http.StatusInternalServerError)
        return
    }

    log.Println("Data fetched successfully from Jira API")
    w.Header().Set("Content-Type", "application/json")
    w.Write(body)
}

func main() {
    http.HandleFunc("/login", HandleLogin)
    http.HandleFunc("/callback", HandleCallback)
    http.HandleFunc("/fetch-data", FetchData)

    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
