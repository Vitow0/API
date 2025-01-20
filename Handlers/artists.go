package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// Get the artist struct
type Artist struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Image        string   `json:"image"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Dates        []string `json:"dates"`
	Locations    string   `json:"locations"`
	Relations    []string `json:"members"`
}

// Function to fetch data from the API
func FetchArtists() ([]Artist, error) {
	// Get the API for the artists
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var artists []Artist
	if err := json.NewDecoder(response.Body).Decode(&artists); err != nil {
		return nil, err
	}
	// Call the function FetchArtistDates to fetch the dates for artists
	artistDates, err := FetchArtistDates() 
	if err != nil {
		return nil, err
	}
	for i := range artists {
		// Call the function FetchArtistDates to fetch the location for artists
		locations, err := FetchLocationsForArtist(artists[i].ID)
		if err == nil {
			artists[i].Locations = strings.Join(locations, ", ")
		}
		artists[i].Dates = artistDates[artists[i].ID]
	}
	return artists, nil
}

// Function to fetch artist dates from the API
func FetchArtistDates() (map[int][]string, error) {
	// Get the API for the dates artists
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/dates")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	// Get data response
	var dateResponse struct {
		// set Index struct for artists date
		Index []struct {
			ID    int      `json:"id"`
			Dates []string `json:"dates"`
		} `json:"index"`
	}
	if err := json.NewDecoder(response.Body).Decode(&dateResponse); err != nil {
		return nil, err
	}
	artistDates := make(map[int][]string)
	for _, entry := range dateResponse.Index {
		artistDates[entry.ID] = entry.Dates
	}
	return artistDates, nil
}

// Function to fetch locations for a given artist
func FetchLocationsForArtist(artistID int) ([]string, error) {
	// Get the API for the locations
	response, err := http.Get(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/locations/%d", artistID))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	// Get location data for the location of the artists
	var locationData struct {
		Locations []string `json:"locations"`
	}
	if err := json.NewDecoder(response.Body).Decode(&locationData); err != nil {
		return nil, err
	}
	return locationData.Locations, nil
}

// Function to fetch latitude and longitude for a given address on a map
func GetCoordinates(address string) (float64, float64, error) {
	apiKey := "34a441c385754c569b0b89e63fc51b85"			// API Key for the map
	baseURL := "https://api.opencagedata.com/geocode/v1/json" // URL for the map

	// Set parameters to display the map API
	query := url.Values{}
	query.Set("q", address)
	query.Set("key", apiKey)
	query.Set("limit", "1")

	resp, err := http.Get(fmt.Sprintf("%s?%s", baseURL, query.Encode()))
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	// Set the struct location
	var geoResponse struct {
		Results []struct {
			Geometry struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"geometry"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&geoResponse); err != nil {
		return 0, 0, err
	}
	if len(geoResponse.Results) == 0 {
		return 0, 0, fmt.Errorf("no results found for address: %s", address)
	}
	return geoResponse.Results[0].Geometry.Lat, geoResponse.Results[0].Geometry.Lng, nil
}

// Function to displays artist details in json format
func displayArtistDetails(w http.ResponseWriter, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid artist ID", http.StatusBadRequest)
		return
	}
	artists, err := FetchArtists()
	if err != nil {
		http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
		return
	}
	// Read all artists informations (only the ID here)
	for _, artist := range artists {
		if artist.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(

				// Get the struct for all the details of the artists
				struct {
				ID           int      `json:"id"`
				Name         string   `json:"name"`
				Image        string   `json:"image"`
				CreationDate int      `json:"creationDate"`
				FirstAlbum   string   `json:"firstAlbum"`
				Dates        []string `json:"dates"`
				Locations    string   `json:"locations"`
				Relations    []string `json:"members"`
				BackURL      string   `json:"back_url"`
			}{
				ID:           artist.ID,
				Name:         artist.Name,
				Image:        artist.Image,
				CreationDate: artist.CreationDate,
				FirstAlbum:   artist.FirstAlbum,
				Dates:        artist.Dates,
				Locations:    artist.Locations,
				Relations:    artist.Relations,
				BackURL:      "http://localhost:8080/artists",
			})
			return
		}
	}
	http.Error(w, "Artist not found", http.StatusNotFound)
}
// Function to handle the templates artists
func ArtistsHandler(w http.ResponseWriter, r *http.Request) {
	// Handler for the location of the artists
	place := r.URL.Query().Get("place")
	if place != "" {
		lat, lng, err := GetCoordinates(place)
		if err != nil {
			http.Error(w, "Unable to geocode location", http.StatusInternalServerError)
			return
		}
		tmpl, err := template.ParseFiles("web/html/locations.html")
		if err != nil {
			http.Error(w, "Unable to load template", http.StatusInternalServerError)
			return
		}
		// Get the data struct for the map
		data := struct {
			Place string
			Lat   float64
			Lng   float64
		}{
			Place: place,
			Lat:   lat,
			Lng:   lng,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Unable to render template", http.StatusInternalServerError)
		}
		return
	}
	// Display the summary details of the function FetchArtists
	artists, err := FetchArtists()
	if err != nil {
		log.Printf("Error fetching artists: %v", err)
		http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
		return
	}
	// Variable for filtered and search artists
	query := strings.ToLower(r.URL.Query().Get("q"))
	dates := r.URL.Query().Get("dates")
	memberCount, _ := strconv.Atoi(r.URL.Query().Get("memberCount"))
	idParam := r.URL.Query().Get("id")
	// Call the function to display the Artists Details
	if idParam != "" {
		displayArtistDetails(w, idParam)
		return
	}
	// Variable to filtered the artists
	var filtered []Artist
	// Read all artists informations
	for _, artist := range artists {
		// Search the artists
		if query != "" && !strings.Contains(strings.ToLower(artist.Name), query) &&
			!strings.Contains(strings.ToLower(strings.Join(artist.Relations, " ")), query) {
			continue
		}
		// Filtered the artists
		matchesDate := dates == "" || containsDate(artist.Dates, dates)
		matchesMembers := memberCount == 0 || len(artist.Relations) == memberCount

		if matchesDate && matchesMembers {
			filtered = append(filtered, artist)
		}
	}
	// Check if the templates is working
	tmpl, err := template.New("artists.html").Funcs(template.FuncMap{
		"split": strings.Split,
	}).ParseFiles("web/html/artists.html")
	if err != nil {
		log.Printf("Error loading template: %v", err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	// Get the details summary artist struct
	type ArtistSummary struct {
		ID        int      `json:"id"`
		Name      string   `json:"name"`
		Image     string   `json:"image"`
		Dates     []string `json:"dates"`
		Locations string   `json:"locations"`
		Relations []string `json:"members"`
	}
	// Check only these options for the filtered artists
	var artistSummaries []ArtistSummary
	for _, artist := range filtered {
		artistSummaries = append(artistSummaries, ArtistSummary{
			ID:        artist.ID,
			Name:      artist.Name,
			Image:     artist.Image,
			Dates:     artist.Dates,
			Locations: artist.Locations,
			Relations: artist.Relations,
		})
	}
	// Defined the struct for the details of the summary artists
	data := struct {
		Artists []ArtistSummary
	}{
		Artists: artistSummaries,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}


// Function checks if a target date is in a list of dates
func containsDate(dates []string, targetDate string) bool {
	layout := "02-01-2006"
	// Accept the date without this sign "*"
	targetDate = strings.TrimPrefix(targetDate, "*")
	target, err := time.Parse(layout, targetDate)
	if err != nil {
		return false
	}
	// Read all the dates from artists
	for _, date := range dates {
		// Accept the date with this sign "*"
		cleanDate := strings.TrimPrefix(date, "*")
		parsedDate, err := time.Parse(layout, cleanDate)
		if err == nil && parsedDate.Equal(target) {
			return true
		}
	}
	return false
}
