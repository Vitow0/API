package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"text/template"
	"strconv"
	"strings"
)

// Struct Artist 
type Artist struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Image     string   `json:"image"`
	Dates     []string `json:"dates"`
	Locations string   `json:"locations"`
	Members   []string `json:"members"`
}

// Function to get the data from the API
func FetchArtists() ([]Artist, error) {
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var artists []Artist
	err = json.NewDecoder(response.Body).Decode(&artists)
	if err != nil {
		return nil, err
	}
	return artists, nil
}

func FetchArtistDates() (map[int][]string, error) {
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
	err = json.NewDecoder(response.Body).Decode(&dateResponse)
	if err != nil {
		return nil, err
	}
	artistDates := make(map[int][]string)
	for _, entry := range dateResponse.Index {
		artistDates[entry.ID] = entry.Dates
	}
	return artistDates, nil
}


func ArtistsHandler(w http.ResponseWriter, r *http.Request) {
	artists, err := FetchArtists()
	if err != nil {
		log.Printf("Error fetching artists: %v", err)
		http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
		return
	}
	artistDates, err := FetchArtistDates()
	if err != nil {
		log.Printf("Error fetching artist dates: %v", err)
		http.Error(w, "Unable to fetch artist dates", http.StatusInternalServerError)
		return
	}
	for i := range artists {
		artists[i].Dates = artistDates[artists[i].ID]
	}

	tmpl, err := template.ParseFiles("web/html/artists.html")

	if err != nil {
		log.Printf("Error loading template: %v", err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	data := struct {
		Artists []Artist
	}{
		Artists: artists,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}



func FiltersHandler(w http.ResponseWriter, r *http.Request) {
	artists, err := FetchArtists()
	if err != nil {
		http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
		return
	}

	
	Dates := r.URL.Query().Get("dates")
	memberCount, _ := strconv.Atoi(r.URL.Query().Get("memberCount"))

	var filtered []Artist
	for _, artist := range artists {

		// Check if the artist matches the date filter
		matchesDate := false
		if Dates == "" {
			matchesDate = true
		} else {
			for _, date := range artist.Dates {
				if date >= Dates { 
					matchesDate = true
					break
				}
			}
		}

		// check if the artist is matched
		matchesMembers := memberCount == 0 || len(artist.Members) == memberCount

		// Add to filtered list if two parameters match
		if matchesDate && matchesMembers {
			filtered = append(filtered, artist)
		}
	}

	// Render the results
	tmpl, err := template.ParseFiles("web/html/filters.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Artists []Artist
	}{
		Artists: filtered,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
    query := strings.ToLower(r.URL.Query().Get("q"))
    if query == "" {
        tmpl, err := template.ParseFiles("web/html/search.html")
        if err != nil {
            http.Error(w, "Unable to load template", http.StatusInternalServerError)
            return
        }

        data := struct {
            Message string
        }{
            Message: "Please enter a search query.",
        }

        tmpl.Execute(w, data)
        return
    }

    artists, err := FetchArtists()
    if err != nil {
        http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
        return
    }

    var results []Artist
    for _, artist := range artists {
        if strings.Contains(strings.ToLower(artist.Name), query) ||
            strings.Contains(strings.ToLower(strings.Join(artist.Members, " ")), query) {
            results = append(results, artist)
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(results)
}