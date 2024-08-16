// package main

// import (
// 	"context"
// 	"database/sql"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// 	"os"
	
// 	"github.com/gorilla/mux"
// 	"github.com/joho/godotenv"
// 	_ "github.com/lib/pq"
// 	"golang.org/x/oauth2"
// )

// var (
// 	db             *sql.DB
// 	oauth2Config   *oauth2.Config
// )

// func init() {
// 	// Load environment variables
// 	if err := godotenv.Load(); err != nil {
// 		log.Fatal("Error loading .env file:", err)
// 	}

// 	// Initialize OAuth2 config
// 	oauth2Config = &oauth2.Config{
// 		ClientID:     os.Getenv("CLIENT_ID"),
// 		ClientSecret: os.Getenv("CLIENT_SECRET"),
// 		RedirectURL:  os.Getenv("REDIRECT_URI"),
// 		Scopes:       []string{"read:jira-user", "read:jira-work"},
// 		Endpoint: oauth2.Endpoint{
// 			AuthURL:  os.Getenv("AUTH_URL"),
// 			TokenURL: os.Getenv("TOKEN_URL"),
// 		},
// 	}

// 	// Connect to PostgreSQL
// 	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
// 		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
// 	var err error
// 	db, err = sql.Open("postgres", connStr)
// 	if err != nil {
// 		log.Fatal("Error connecting to the database:", err)
// 	}
// }

// func HandleLogin(w http.ResponseWriter, r *http.Request) {
// 	url := oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
// 	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
// }

// func HandleCallback(w http.ResponseWriter, r *http.Request) {
//     log.Println("HandleCallback")
//     code := r.URL.Query().Get("code")
    
//     log.Println("code", code)

//     // Exchange the authorization code for an access token
//     token, err := oauth2Config.Exchange(context.Background(), code)
//     if err != nil {
//         http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
//         return
//     }
//     log.Println("Access token received:", token.AccessToken)
//     // Retrieve user information (accessible resources)
//     client := oauth2Config.Client(context.Background(), token)
//     resp, err := client.Get("https://api.atlassian.com/oauth/token/accessible-resources")
//     if err != nil {
//         http.Error(w, "Failed to get accessible resources: "+err.Error(), http.StatusInternalServerError)
//         return
//     }
//     defer resp.Body.Close()

//     var resources []map[string]interface{}
//     if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
//         http.Error(w, "Failed to decode resources: "+err.Error(), http.StatusInternalServerError)
//         return
//     }

//     if len(resources) == 0 {
//         http.Error(w, "No accessible resources found", http.StatusInternalServerError)
//         return
//     }

//     userID := resources[0]["id"].(string)    // Use the resource ID as the userID or adjust as needed
//     cloudID := resources[0]["id"].(string)   // Assume the cloud ID is in the first resource's "id" field

//     // Store the token in the database
//     err = storeToken(userID, token)
//     if err != nil {
//         http.Error(w, "Failed to store token: "+err.Error(), http.StatusInternalServerError)
//         return
//     }

//     // Fetch data using the cloudID
//     apiUrl := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/project", cloudID)
//     resp, err = client.Get(apiUrl)
//     if err != nil {
//         http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
//         return
//     }
//     defer resp.Body.Close()

//     body, err := ioutil.ReadAll(resp.Body)
//     if err != nil {
//         http.Error(w, "Failed to read response: "+err.Error(), http.StatusInternalServerError)
//         return
//     }

//     // Log the actual JSON response as a string
//     log.Println("Response body:", string(body))
//     log.Println("Jira connected and data fetched successfully!")

//     // Return the fetched data as a response
//     w.Header().Set("Content-Type", "application/json")
//     w.Write(body)
// }


// func storeToken(userID string, token *oauth2.Token) error {
// 	var existingAccessToken string
//     log.Println("token got:", token.AccessToken)
// 	// Check if there is an existing entry for the user
// 	err := db.QueryRow("SELECT access_token FROM users WHERE user_id = $1", userID).Scan(&existingAccessToken)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			// No entry found, insert the new token
// 			_, err = db.Exec(`
// 				INSERT INTO users (user_id, access_token, refresh_token, token_expiry)
// 				VALUES ($1, $2, $3, $4)
// 			`, userID, token.AccessToken, token.RefreshToken, token.Expiry)
// 			if err != nil {
// 				return fmt.Errorf("failed to insert new token: %v", err)
// 			}
// 			log.Println("New token inserted for user:", userID)
// 		} else {
// 			// Another error occurred while querying
// 			return fmt.Errorf("failed to check existing token: %v", err)
// 		}
// 	} else {
// 		// Entry exists, check if the token has changed
// 		if existingAccessToken != token.AccessToken {
// 			_, err = db.Exec(`
// 				UPDATE users
// 				SET access_token = $2, refresh_token = $3, token_expiry = $4
// 				WHERE user_id = $1
// 			`, userID, token.AccessToken, token.RefreshToken, token.Expiry)
// 			if err != nil {
// 				return fmt.Errorf("failed to update token: %v", err)
// 			}
// 			log.Println("Access token updated successfully for user:", userID)
// 		} else {
// 			log.Println("Access token is the same, no update needed for user:", userID)
// 		}
// 	}

// 	return nil
// }



// func main() {
// 	r := mux.NewRouter()
// 	r.HandleFunc("/login", HandleLogin)
// 	r.HandleFunc("/callback", HandleCallback)
    
// 	log.Println("Starting server on :8080")
// 	log.Fatal(http.ListenAndServe(":8080", r))

// }
    //cron
        // // Create a new cron scheduler
        // c := cron.New()

        // // Schedule fetchData to run every 5 hours
        // _, err := c.AddFunc("0 */5 * * *", func() {
        //     fetchData()
        // })
        // if err != nil {
        //     log.Fatalf("Error scheduling cron job: %v", err)
        // }
    
        // // Start the cron scheduler
        // c.Start()
    
        // // Keep the application running
        // select {} // Blocks forever
package main

import (
    "jira-go-connector/api"
    "jira-go-connector/config"
    "jira-go-connector/db"
    "jira-go-connector/oauth"
    "log"
    "net/http"

    "github.com/gorilla/mux"
)

func main() {
    // Load configuration
    config.Load()

    // Initialize database connection
    db.Init()

    // Initialize OAuth2 configuration
    oauth.InitOAuthConfig()

    // Initialize the router and routes
    r := mux.NewRouter()
    r.HandleFunc("/login", api.HandleLogin)
    r.HandleFunc("/callback", api.HandleCallback)


    // Start the server
    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
