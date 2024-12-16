package main

import (
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

func main() {
	http.HandleFunc("/artists", handlers.ArtistsHandler) // go to artists handlers (http://localhost:8080/artists)
	http.HandleFunc("/locations", handlers.LocationsHandler) // go to locations handlers (http://localhost:8080/locations)
	http.HandleFunc("/relations", handlers.RelationsHandler) // go to relations handlers (http://localhost:8080/relations)
	http.HandleFunc("/dates", handlers.DatesHandler) // go to dates handlers (http://localhost:8080/dates)
	http.HandleFunc("/filters", handlers.FiltersHandler) // go to filters (http://localhost:8080/filters)
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web")))) // take the file css to relie for the templates html
	http.HandleFunc("/search", handlers.SearchHandler) // go to search (http://localhost:8080/search)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
