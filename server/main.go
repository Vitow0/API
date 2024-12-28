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
	http.HandleFunc("/location", handlers.LocationHandler)
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web")))) // take the file css to relie for the templates html
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
