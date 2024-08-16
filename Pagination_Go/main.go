package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Movie struct {
	Title         string  `json:"title"`
	Overview      string  `json:"overview"`
	OriginalLang  string  `json:"original_language"`
	ReleaseDate   string  `json:"release_date"`
	Popularity    float64 `json:"popularity"`
	VoteCount     int     `json:"vote_count"`
	VoteAverage   float64 `json:"vote_average"`
}

type Response struct {
	Page         int     `json:"page"`
	TotalPages   int     `json:"total_pages"`
	Results      []Movie `json:"results"`
}

func fetchMovies(apiKey string, maxPages int) ([]Movie, error) {
	baseURL := "https://api.themoviedb.org/3/discover/movie"
	page := 1
	var movies []Movie

	for page <= maxPages {
		// Construct the URL with the current page
		url := fmt.Sprintf("%s?api_key=%s&page=%d", baseURL, apiKey, page)
		fmt.Printf("Fetching page %d\n", page)

		// Make the HTTP request
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("error making the request: %v", err)
		}
		defer resp.Body.Close()

		// Check if the request was successful
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// Decode the JSON response
		var response Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return nil, fmt.Errorf("error decoding the response: %v", err)
		}

		// Add the results to the movie list
		movies = append(movies, response.Results...)

		// Update total pages based on the first request
		if page == 1 {
			maxPages = response.TotalPages
			// For testing purposes, limit the pages to 5
			if maxPages > 5 {
				maxPages = 5
			}
		}

		// Increment the page
		page++
		// Sleep for a second to avoid hitting rate limits
		time.Sleep(1 * time.Second)
	}

	return movies, nil
}

func main() {
	apiKey := "256da2d742d5a5979790e6833447e4b4"
	movies, err := fetchMovies(apiKey, 1) // Initially fetch the first page
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print the movie data
	for _, movie := range movies {
		fmt.Printf("Title: %s\n", movie.Title)
		fmt.Printf("Overview: %s\n", movie.Overview)
		fmt.Printf("Original Language: %s\n", movie.OriginalLang)
		fmt.Printf("Release Date: %s\n", movie.ReleaseDate)
		fmt.Printf("Popularity: %.2f\n", movie.Popularity)
		fmt.Printf("Vote Count: %d\n", movie.VoteCount)
		fmt.Printf("Vote Average: %.2f\n", movie.VoteAverage)
		fmt.Println("-----------------------------")
	}
}
