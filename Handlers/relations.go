package handlers

import (
	"encoding/json"
	"net/http"
	"text/template"
)

// Struct relation for the data struct
type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// Functions to get the data from API
func FetchRelations() ([]Relation, error) {
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/relation")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var relations []Relation
	err = json.NewDecoder(response.Body).Decode(&relations)
	if err != nil {
		return nil, err
	}
	return relations, nil
}

// Function to display from the templates html
func RelationsHandler(w http.ResponseWriter, r *http.Request) {
	relations, err := FetchRelations()
	if err != nil {
		http.Error(w, "Unable to fetch relations", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("web/html/relations.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Relations []Relation
	}{
		Relations: relations,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}