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

// Artist struct
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

// Function to display data from the API
func FetchArtists() ([]Artist, error) {
	// get the API fo the artists
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var artists []Artist
	if err := json.NewDecoder(response.Body).Decode(&artists); err != nil {
		return nil, err
	}
	artistDates, err := FetchArtistDates()
	if err != nil {
		return nil, err
	}
	for i := range artists {
		locations, err := FetchLocationsForArtist(artists[i].ID)
		if err == nil {
			artists[i].Locations = strings.Join(locations, ", ")
		}
		artists[i].Dates = artistDates[artists[i].ID]
	}
	return artists, nil
}

// Function to display artist dates from the API
func FetchArtistDates() (map[int][]string, error) {
	// get the API for the dates artists
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/dates")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var dateResponse struct {
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

// Function to display locations for a given artist
func FetchLocationsForArtist(artistID int) ([]string, error) {
	// get the API for the locations
	response, err := http.Get(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/locations/%d", artistID))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var locationData struct {
		Locations []string `json:"locations"`
	}
	if err := json.NewDecoder(response.Body).Decode(&locationData); err != nil {
		return nil, err
	}
	return locationData.Locations, nil
}

// Function fetches latitude and longitude for a given address in a map
func GetCoordinates(address string) (float64, float64, error) {
	apiKey := "34a441c385754c569b0b89e63fc51b85"			// API Key for the map
	baseURL := "https://api.opencagedata.com/geocode/v1/json" // URL for the map

	// set parameters to display the map API
	query := url.Values{}
	query.Set("q", address)
	query.Set("key", apiKey)
	query.Set("limit", "1")

	resp, err := http.Get(fmt.Sprintf("%s?%s", baseURL, query.Encode()))
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	// set the struct location
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

// Function to displays artist details in JSON format
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
	for _, artist := range artists {
		if artist.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(

				// get the struct for all the details of the artists
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

// Function to handles artist-related requests
func ArtistsHandler(w http.ResponseWriter, r *http.Request) {
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
		// get the struct for the map
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
	artists, err := FetchArtists()
	if err != nil {
		log.Printf("Error fetching artists: %v", err)
		http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
		return
	}
	query := strings.ToLower(r.URL.Query().Get("q"))
	dates := r.URL.Query().Get("dates")
	memberCount, _ := strconv.Atoi(r.URL.Query().Get("memberCount"))
	idParam := r.URL.Query().Get("id")

	if idParam != "" {
		displayArtistDetails(w, idParam)
		return
	}
	var filtered []Artist
	for _, artist := range artists {
		if query != "" && !strings.Contains(strings.ToLower(artist.Name), query) &&
			!strings.Contains(strings.ToLower(strings.Join(artist.Relations, " ")), query) {
			continue
		}
		matchesDate := dates == "" || containsDate(artist.Dates, dates)
		matchesMembers := memberCount == 0 || len(artist.Relations) == memberCount

		if matchesDate && matchesMembers {
			filtered = append(filtered, artist)
		}
	}
	tmpl, err := template.New("artists.html").Funcs(template.FuncMap{
		"split": strings.Split,
	}).ParseFiles("web/html/artists.html")
	if err != nil {
		log.Printf("Error loading template: %v", err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	// get the details structs
	type ArtistSummary struct {
		ID        int      `json:"id"`
		Name      string   `json:"name"`
		Image     string   `json:"image"`
		Dates     []string `json:"dates"`
		Locations string   `json:"locations"`
		Relations []string `json:"members"`
	}
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
	// defined the structs for the details of the artists
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

// Function to checks if a target date is in a list of dates
func containsDate(dates []string, targetDate string) bool {
	layout := "02-01-2006"
	targetDate = strings.TrimPrefix(targetDate, "*")
	target, err := time.Parse(layout, targetDate)
	if err != nil {
		return false
	}
	for _, date := range dates {
		cleanDate := strings.TrimPrefix(date, "*")
		parsedDate, err := time.Parse(layout, cleanDate)
		if err == nil && parsedDate.Equal(target) {
			return true
		}
	}
	return false
}
