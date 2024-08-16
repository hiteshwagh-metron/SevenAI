package oauth

import (
    "jira-go-connector/config"
    "golang.org/x/oauth2"
    "log"
)

var OAuth2Config *oauth2.Config

func InitOAuthConfig() {
    log.Println("InitOAuthConfig called")
    OAuth2Config = &oauth2.Config{
        ClientID:     config.ClientID,
        ClientSecret: config.ClientSecret,
        RedirectURL:  config.RedirectURI,
        Scopes:       []string{"read:jira-user", "read:jira-work"},
        Endpoint: oauth2.Endpoint{
            AuthURL:  config.AuthURL,
            TokenURL: config.TokenURL,
        },
    }
}
