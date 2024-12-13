package handlers

import (
	"encoding/json"
	"net/http"
	"text/template"
	"log"
)
// struct data for date
type Date struct {
    ID      int      `json:"id"`
    Dates []string `json:"dates"`
	Name    string   `json:"name"`
}
// struct for data response
type DateResponse struct {
    Index []Date `json:"index"`
}
// function to get the data from API
func FetchDates() ([]Date, error) {
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/dates")
    if err != nil {
        return nil, err
    }
    defer response.Body.Close()
    var dateResponse DateResponse
    err = json.NewDecoder(response.Body).Decode(&dateResponse)
    if err != nil {
        return nil, err
    }
    return dateResponse.Index, nil
}
// function to display the templates
func DatesHandler(w http.ResponseWriter, r *http.Request) {
	// display if there is any error
	dates, err := FetchDates()
	if err != nil {
		log.Printf("Error fetching dates: %v", err)
		http.Error(w, "Unable to fetch Dates", http.StatusInternalServerError)
		return
	}
	// get the templates file html
	tmpl, err := template.ParseFiles("web/html/dates.html")
	if err != nil {
		log.Printf("Error loading template: %v", err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	//get the data struct date
	data := struct {
		Dates []Date
	}{
		Dates: dates,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}
