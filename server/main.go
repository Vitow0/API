package main

import (
	"encoding/json"
	"net/http"
	"groupie/Handlers"
)

type Web struct {
    ID       int      `json:"id"`
    Name     string   `json:"name"`
    Image    string   `json:"image"`
    Dates    []string `json:"dates"`
    Locations []string `json:"locations"`
}

func Default_Web() ([]Web, error) {
	respond, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		return nil, err
	}
	defer respond.Body.Close()

	var artists []Web
	err = json.NewDecoder(respond.Body).Decode(&artists)
	if err != nil {
		return nil, err
	}
	return artists, nil
}

func main() {
	http.HandleFunc("/artists", handlers.ArtistsHandler)
	http.HandleFunc("/locations", handlers.LocationsHandler)
	http.HandleFunc("/relations", handlers.RelationsHandler)
	http.HandleFunc("/dates", handlers.DatesHandler)
	http.HandleFunc("/filters", handlers.FiltersHandler)
	http.HandleFunc("/search", handlers.SearchHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
