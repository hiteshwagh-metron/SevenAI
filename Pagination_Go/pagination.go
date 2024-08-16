package main
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"context"
	"github.com/jackc/pgx/v4"
)

type Movie struct{
	Title string `json:"title"`
	Overview string `json:"overview"`
	OriginalLang string `json:"original_language"`
	ReleaseDate string `json:"release_date"`
	Popularity float64 `json:"popularity"`
	VoteCount     int     `json:"vote_count"`
	VoteAverage   float64 `json:"vote_average"`
}

type Response struct {
	Page         int     `json:"page"`
	TotalPages   int     `json:"total_pages"`
	Results      []Movie `json:"results"`
}

func fetchMovies(api_key string, startPage int, conn *pgx.Conn) ([]Movie, error){
	baseURL:="https://api.themoviedb.org/3/discover/movie"
	page := startPage
	var movies []Movie
	
	for {
		u, err := url.Parse(baseURL) //This function from the net/url package parses a URL string (baseURL) and returns a url.URL struct.
		if err != nil {
			return nil, fmt.Errorf("error parsing base URL: %v ", err)
		}
		params:= url.Values{}  //  This initializes params as an empty url.Values map. At this point, it’s an empty map, ready to have query parameters added to it.
		params.Add("api_key", api_key) //adds query parameters to the params map
		params.Add("page",fmt.Sprintf("%d",page)) //fmt.Sprintf is a flexible way to format data as a string. You can use it to convert an integer to a string with a specific format.

		u.RawQuery = params.Encode() // https://api.themoviedb.org/3/discover/movie?api_key=your_api_key&page=2
		fmt.Println(u.String())
		fmt.Printf("Fetching page %d \n",page)

		resp, err := http.Get(u.String()) // HTTP GET request to the URL represented by u
		if err != nil {
			return nil, fmt.Errorf("error making the request: %v", err)
		}

		defer resp.Body.Close() // Ensures that the response body is closed when the function returns, releasing any resources associated with it.

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		var response Response
		err= json.NewDecoder(resp.Body).Decode(&response) // json.NewDecoder(resp.Body): Creates a new JSON decoder that reads from the response body. Decode(&response): Decodes the JSON data into the response variable.
		if err != nil {
			return nil, fmt.Errorf("error decoding the response: %v", err)
		}
		if err := updateLastPage(conn, page); err != nil {
			return nil, fmt.Errorf("error updating last page number: %v", err)
		}
		movies = append(movies, response.Results...)
		if page>=response.TotalPages{
			break
		}
		page++
		time.Sleep(1 * time.Second)
	}
	return movies, nil
}
func getLastPage(conn *pgx.Conn) (int, error) {
	var lastPage int
	err := conn.QueryRow(context.Background(), "SELECT last_page FROM fetch_status WHERE id = 1").Scan(&lastPage)
	if err != nil {
		return 0, err
	}
	return lastPage, nil
}

func updateLastPage(conn *pgx.Conn, page int) error {
	_, err := conn.Exec(context.Background(), "UPDATE fetch_status SET last_page = $1 WHERE id = 1", page)
	return err
}
func main(){
	api_key:= "256da2d742d5a5979790e6833447e4b4"

	dbConnString:= "postgres://plsql:plsql@localhost:5432/mydatabase"
	conn, err := pgx.Connect(context.Background(), dbConnString)
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	startPage, err := getLastPage(conn)
	if err != nil {
		fmt.Printf("Error reading last page: %v\n", err)
		return
	}

	movies, err := fetchMovies(api_key, startPage+1, conn)
	if  err != nil{
		fmt.Printf("Error: %v \n ", err)
	}

	for _ , movie:= range movies{
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