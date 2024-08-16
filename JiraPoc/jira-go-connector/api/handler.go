package api

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
    "jira-go-connector/db"
    "jira-go-connector/oauth"
    "golang.org/x/oauth2"
)


func HandleLogin(w http.ResponseWriter, r *http.Request) {
    url := oauth.OAuth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
    log.Println("Generated OAuth2 Authorization URL:", url)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleCallback(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")

    token, err := oauth.OAuth2Config.Exchange(context.Background(), code)
    if err != nil {
        http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
        return
    }

    client := oauth.OAuth2Config.Client(context.Background(), token)
    resp, err := client.Get("https://api.atlassian.com/oauth/token/accessible-resources")
    if err != nil {
        http.Error(w, "Failed to get accessible resources: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    var resources []map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
        http.Error(w, "Failed to decode resources: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if len(resources) == 0 {
        http.Error(w, "No accessible resources found", http.StatusInternalServerError)
        return
    }

    userID := resources[0]["id"].(string)
    cloudID := resources[0]["id"].(string)

    err = storeToken(userID, token)
    if err != nil {
        http.Error(w, "Failed to store token: "+err.Error(), http.StatusInternalServerError)
        return
    }

    apiUrl := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/project", cloudID)
    resp, err = client.Get(apiUrl)
    if err != nil {
        http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        http.Error(w, "Failed to read response: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    log.Println("Response body:", string(body))
    Fedratedsearch(string(body))
    w.Write(body)
}

func storeToken(userID string, token *oauth2.Token) error {
    var existingAccessToken string
    err := db.DB.QueryRow("SELECT access_token FROM users WHERE user_id = $1", userID).Scan(&existingAccessToken)
    if err != nil {
        if err == sql.ErrNoRows {
            _, err = db.DB.Exec(`
                INSERT INTO users (user_id, access_token, refresh_token, token_expiry)
                VALUES ($1, $2, $3, $4)
            `, userID, token.AccessToken, token.RefreshToken, token.Expiry)
            if err != nil {
                return fmt.Errorf("failed to insert new token: %v", err)
            }
            log.Println("New token inserted for user:", userID)
        } else {
            return fmt.Errorf("failed to check existing token: %v", err)
        }
    } else {
        if existingAccessToken != token.AccessToken {
            _, err = db.DB.Exec(`
                UPDATE users
                SET access_token = $2, refresh_token = $3, token_expiry = $4
                WHERE user_id = $1
            `, userID, token.AccessToken, token.RefreshToken, token.Expiry)
            if err != nil {
                return fmt.Errorf("failed to update token: %v", err)
            }
            log.Println("Access token updated successfully for user:", userID)
        } else {
            log.Println("Access token is the same, no update needed for user:", userID)
        }
    }

    return nil
}

type AvatarUrls struct {
    X48 string `json:"48x48"`
    X24 string `json:"24x24"`
    X16 string `json:"16x16"`
    X32 string `json:"32x32"`
}

type Project struct {
    Expand        string     `json:"expand"`
    Self          string     `json:"self"`
    ID            string     `json:"id"`
    Key           string     `json:"key"`
    Name          string     `json:"name"`
    AvatarUrls    AvatarUrls `json:"avatarUrls"`
    ProjectTypeKey string    `json:"projectTypeKey"`
    Simplified    bool       `json:"simplified"`
    Style         string     `json:"style"`
    IsPrivate     bool       `json:"isPrivate"`
    Properties    map[string]interface{} `json:"properties"`
    EntityID      string     `json:"entityId"`
    UUID          string     `json:"uuid"`
}
func Fedratedsearch(body string) {
    fmt.Println("Fedratedsearch called")
    // Unmarshal JSON string into a slice of Project
    var projects []Project
    searchTerm := "ST" 
    err := json.Unmarshal([]byte(body), &projects)
    if err != nil {
        fmt.Println("Error unmarshalling JSON:", err)
        return
    }

    // Convert search term to lowercase for case-insensitive search
    searchTerm = strings.ToLower(searchTerm)

    // Search for projects matching the search term
    var result []Project
    for _, project := range projects {
        if strings.Contains(strings.ToLower(project.Name), searchTerm) {
            result = append(result, project)
        }
    }

    // Print the search results
    for _, project := range result {
        fmt.Printf("ID: %s, Key: %s, Name: %s\n", project.ID, project.Key, project.Name)
    }
}