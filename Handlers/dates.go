package handlers

import (
	"encoding/json"
	"net/http"
	"text/template"
	"log"
)

type Date struct {
    ID    int      `json:"id"`
    Dates []string `json:"dates"` 
}
func FetchDates()([]Date, error) {
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/dates")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var Dates []Date
	err = json.NewDecoder(response.Body).Decode(&Dates)
	if err != nil {
		return nil, err
	}
	return Dates, nil
}

func DatesHandler(w http.ResponseWriter, r *http.Request) {
	dates, err := FetchDates()
	if err != nil {
		log.Printf("Error fetching dates: %v", err)
		http.Error(w, "Unable to fetch Dates", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("web/html/dates.html")
	if err != nil {
		log.Printf("Error loading template: %v", err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

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
