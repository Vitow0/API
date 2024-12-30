package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"
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

	artistDates, err := FetchArtistDates()
	if err != nil {
		return nil, err
	}

	for i := range artists {
		locations, err := FetchLocationsForArtist(artists[i].ID)
		if err == nil {
			artists[i].Locations = strings.Join(locations, ", ")
		}
		artists[i].Dates = artistDates[artists[i].ID]
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

func displayArtistDetails(w http.ResponseWriter, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid artist ID", http.StatusBadRequest)
		return
	}

	artists, err := FetchArtists()
	if err != nil {
		http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
		return
	}

	for _, artist := range artists {
		if artist.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(artist)
			return
		}
	}

	http.Error(w, "Artist not found", http.StatusNotFound)
}

func GetCoordinates(address string) (float64, float64, error) {
	apiKey := "34a441c385754c569b0b89e63fc51b85" 
	baseURL := "https://api.opencagedata.com/geocode/v1/json"

	query := url.Values{}
	query.Set("q", address)
	query.Set("key", apiKey)
	query.Set("limit", "1")

	requestURL := fmt.Sprintf("%s?%s", baseURL, query.Encode())

	resp, err := http.Get(requestURL)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var geoResponse struct {
		Results []struct {
			Geometry struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"geometry"`
		} `json:"results"`
	}

	err = json.NewDecoder(resp.Body).Decode(&geoResponse)
	if err != nil {
		return 0, 0, err
	}

	if len(geoResponse.Results) == 0 {
		return 0, 0, fmt.Errorf("no results found for address: %s", address)
	}

	lat := geoResponse.Results[0].Geometry.Lat
	lng := geoResponse.Results[0].Geometry.Lng
	return lat, lng, nil
}


func FetchLocationsForArtist(artistID int) ([]string, error) {
	// get the API locations
	response, err := http.Get(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/locations/%d", artistID))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var locationData struct {
		Locations []string `json:"locations"`
	}
	// get and trasmits the locations from the file location.json
	err = json.NewDecoder(response.Body).Decode(&locationData)
	if err != nil {
		return nil, err
	}

	return locationData.Locations, nil
}

// function for the API handlers templates
func ArtistsHandler(w http.ResponseWriter, r *http.Request) {

    // Get the artists ID location from URL
    place := r.URL.Query().Get("place")
    if place != "" {
        lat, lng, err := GetCoordinates(place)
        if err != nil {
            http.Error(w, "Unable to geocode location", http.StatusInternalServerError)
            return
        }

        tmpl, err := template.ParseFiles("web/html/locations.html")
        if err != nil {
            http.Error(w, "Unable to load template", http.StatusInternalServerError)
            return
        }

		// get the data location
        data := struct {
            Place string
            Lat   float64
            Lng   float64
        }{
            Place: place,
            Lat:   lat,
            Lng:   lng,
        }

        if err := tmpl.Execute(w, data); err != nil {
            http.Error(w, "Unable to render template", http.StatusInternalServerError)
        }
        return
    }

    // Get the artists
    artists, err := FetchArtists()
    if err != nil {
        log.Printf("Error fetching artists: %v", err)
        http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
        return
    }

    // Get the filter and search URL
    query := strings.ToLower(r.URL.Query().Get("q"))
    dates := r.URL.Query().Get("dates")
    memberCount, _ := strconv.Atoi(r.URL.Query().Get("memberCount"))
    idParam := r.URL.Query().Get("id")

    // If there is an ID details, then call this function
    if idParam != "" {
        displayArtistDetails(w, idParam)
        return
    }

    // Filtre the artists
    var filtered []Artist
    for _, artist := range artists {
        if query != "" && !strings.Contains(strings.ToLower(artist.Name), query) &&
            !strings.Contains(strings.ToLower(strings.Join(artist.Members, " ")), query) {
            continue
        }

        matchesDate := dates == "" || containsDate(artist.Dates, dates)
        matchesMembers := memberCount == 0 || len(artist.Members) == memberCount

        if matchesDate && matchesMembers {
            filtered = append(filtered, artist)
        }
    }

    // Render the artists in the filter
    tmpl, err := template.New("artists.html").Funcs(template.FuncMap{
        "split": strings.Split,
    }).ParseFiles("web/html/artists.html")
    if err != nil {
        log.Printf("Error loading template: %v", err)
        http.Error(w, "Unable to load template", http.StatusInternalServerError)
        return
    }
	// get the data artists filtered
    data := struct {
        Artists []Artist
    }{
        Artists: filtered,
    }

    if err := tmpl.Execute(w, data); err != nil {
        log.Printf("Error rendering template: %v", err)
        http.Error(w, "Unable to render template", http.StatusInternalServerError)
    }
}

// function to range dates artists
func containsDate(dates []string, targetDate string) bool {
    // Define Date
    layout := "02-01-2006" // Render DD-MM-YYYY

    // remove this caractere * if found out
    targetDate = strings.TrimPrefix(targetDate, "*")

    // Analyse thetarget Date
    target, err := time.Parse(layout, targetDate)

    if err != nil {
        return false // Return false if incorrect target dates
    }

    // Check every date in the list
    for _, date := range dates {
        // remove * if found out
        cleanDate := strings.TrimPrefix(date, "*")
        parsedDate, err := time.Parse(layout, cleanDate)
        if err == nil && parsedDate.Equal(target) {
            return true // return true if a date is found
        }
    }
    return false // return false if there is no date
	
}