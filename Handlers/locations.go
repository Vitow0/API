package handlers

import (
	"encoding/json"
	"net/http"
	"text/template"
    "log"
)

// Struct location for the Data
type Location struct {
	ID       int      `json:"id"`
	Locations []string `json:"locations"`
	Name string  `json:"name"`
    Lat  float64 `json:"lat"`
    Lng  float64 `json:"lng"`
}

// function to get the data from API
func FetchLocations() ([]Location, error) {
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/locations")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var locations []Location
	err = json.NewDecoder(response.Body).Decode(&locations)
	if err != nil {
		return nil, err
	}
	return locations, nil
}

// Function to display in the html
func LocationsHandler(w http.ResponseWriter, r *http.Request) {
    locations, err := FetchLocations()
    if err != nil {
        log.Printf("Error fetching locations: %v", err)
        http.Error(w, "Unable to fetch locations", http.StatusInternalServerError)
        return
    }

    tmpl, err := template.ParseFiles("web/html/locations.html")
    if err != nil {
        http.Error(w, "Unable to load template", http.StatusInternalServerError)
        return
    }

    jsonData, err := json.Marshal(locations) // Convert locations to JSON
    if err != nil {
        http.Error(w, "Unable to encode locations", http.StatusInternalServerError)
        return
    }

    data := struct {
        LocationsJSON string
    }{
        LocationsJSON: string(jsonData), // Pass JSON as a string
    }

    if err := tmpl.Execute(w, data); err != nil {
        http.Error(w, "Unable to render template", http.StatusInternalServerError)
    }
}