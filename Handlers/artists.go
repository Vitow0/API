package handlers

import (
	"encoding/json"
	"net/http"
	"text/template"
	"strings"
	"strconv"
)

// Struct Artist for the data
type Artist struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Image      string   `json:"image"`
	Dates  int      	`json:"dates"`
	Locations string   `json:"locations"`
	Members   []string `json:"members"`
}

// Function to get the data from the API
func FetchArtists() ([]Artist, error) {
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var artists []Artist
	err = json.NewDecoder(response.Body).Decode(&artists)
	if err != nil {
		return nil, err
	}
	return artists, nil
}

// function to diplay the data in the templates html
func ArtistsHandler(w http.ResponseWriter, r *http.Request) {
	artists, err := FetchArtists()
	if err != nil {
		http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("web/html/artists.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Artists []Artist
	}{
		Artists: artists,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}
// function to search the content
func Search_bar(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	if query == "" {
		http.Error(w, "Search query missing", http.StatusBadRequest)
		return
	}

	artists, err := FetchArtists()
	if err != nil {
		http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
		return
	}

	var results []Artist
	for _, artist := range artists {
		if strings.Contains(strings.ToLower(artist.Name), query) ||
			strings.Contains(strings.ToLower(strings.Join(artist.Members, " ")), query) {
			results = append(results, artist)
		}
	}

	w.Header().Set("Content-Type", "web/json/search.json")
	json.NewEncoder(w).Encode(results)
}

//function to filter the results
func FiltersHandler(w http.ResponseWriter, r *http.Request) {
	artists, err := FetchArtists()
	if err != nil {
		http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
		return
	}

	Dates, _ := strconv.Atoi(r.URL.Query().Get("dates"))
	memberCount, _ := strconv.Atoi(r.URL.Query().Get("memberCount"))

	var filtered []Artist
	for _, artist := range artists {
		if (Dates == 0 || artist.Dates >= Dates) &&
			(memberCount == 0 || len(artist.Members) == memberCount) {
			filtered = append(filtered, artist)
		}
	}

	tmpl, err := template.ParseFiles("web/html/filters.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Artists []Artist
	}{
		Artists: filtered,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}
