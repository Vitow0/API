package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"text/template"
	"strconv"
	"strings"
)

// Struct Artist 
type Artist struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Image     string   `json:"image"`
	Dates     []string `json:"dates"`
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

// Function to display the data 
func ArtistsHandler(w http.ResponseWriter, r *http.Request) {

	// get data from API
	artists, err := FetchArtists()
	if err != nil {
		log.Printf("Error fetching artists: %v", err)
		http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
		return
	}

	// get the templates
	tmpl, err := template.ParseFiles("web/html/artists.html")
	if err != nil {
		log.Printf("Error loading template: %v", err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	// give the data
	data := struct {
		Artists []Artist
	}{
		Artists: artists,
	}

	// Render the artists templates
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
		return
	}
}

func FiltersHandler(w http.ResponseWriter, r *http.Request) {
	artists, err := FetchArtists()
	if err != nil {
		http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
		return
	}

	
	Dates := r.URL.Query().Get("dates")
	memberCount, _ := strconv.Atoi(r.URL.Query().Get("memberCount"))

	var filtered []Artist
	for _, artist := range artists {

		// Check if the artist matches the date filter
		matchesDate := false
		if Dates == "" {
			matchesDate = true
		} else {
			for _, date := range artist.Dates {
				if date >= Dates { 
					matchesDate = true
					break
				}
			}
		}

		// check if the artist is matched
		matchesMembers := memberCount == 0 || len(artist.Members) == memberCount

		// Add to filtered list if two parameters match
		if matchesDate && matchesMembers {
			filtered = append(filtered, artist)
		}
	}

	// Render the results
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

func SearchHandler(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}