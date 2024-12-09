package handlers

import (
	"encoding/json"
	"net/http"
	"text/template"
)

// Struct location for the Data
type Location struct {
	ID       int      `json:"id"`
	Locations []string `json:"locations"`
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
		http.Error(w, "Unable to fetch locations", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("web/html/locations.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Locations []Location
	}{
		Locations: locations,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}
