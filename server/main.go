package main

import (
	"net/http"
	"groupie/Handlers"
)

func main() {
	http.HandleFunc("/artists", handlers.ArtistsHandler) // go to artists handlers (http://localhost:8080/artists)
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web")))) // take the file css to relie for the templates html
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
