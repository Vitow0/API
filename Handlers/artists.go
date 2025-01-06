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
	CreationDate int      `json:"creationDate"` 
    FirstAlbum   string   `json:"firstAlbum"`
	Dates     []string `json:"dates"`
	Locations string   `json:"locations"`
	Members   []string `json:"members"`
}

// Function to get the data from the API
func FetchArtists() ([]Artist, error) {

	//get the API artists
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/artists") 
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	// return error if there is an error for decoding the artists
	var artists []Artist
	err = json.NewDecoder(response.Body).Decode(&artists)
	if err != nil {
		return nil, err
	}

	// call the function dates artists
	artistDates, err := FetchArtistDates() 
	if err != nil {
		return nil, err
	}

	// set the loations concert from artists
	for i := range artists {
		locations, err := FetchLocationsForArtist(artists[i].ID)
		if err == nil {
			artists[i].Locations = strings.Join(locations, ", ")
		}
		artists[i].Dates = artistDates[artists[i].ID]
	}
	return artists, nil
}

// function to display the dates from artists
func FetchArtistDates() (map[int][]string, error) {
	// get the API dates for artists
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/dates")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// get the struct Dates
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
	// save the dates for the filter
	artistDates := make(map[int][]string)
	for _, entry := range dateResponse.Index {
		artistDates[entry.ID] = entry.Dates
	}
	return artistDates, nil
}

//function to display the details for artists
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
	// search for all artists
    for _, artist := range artists {
		// if he found the same ID as the artists
        if artist.ID == id {
			// Then he show all the information from the file .json defined
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(
				// defined the struct for the artists_details
				struct {
                ID           int      `json:"id"`
                Name         string   `json:"name"`
                Image        string   `json:"image"`
				CreationDate int      `json:"creationDate"`
                FirstAlbum   string   `json:"firstAlbum"`
                Dates        []string `json:"dates"`
                Locations    string   `json:"locations"`
                Members      []string `json:"members"`
				BackURL      string   `json:"back_url"` // defined the URL to go back
            }{
                ID:           artist.ID,
                Name:         artist.Name,
                Image:        artist.Image,
				CreationDate: artist.CreationDate,
                FirstAlbum:   artist.FirstAlbum,
                Dates:        artist.Dates,
                Locations:    artist.Locations,
                Members:      artist.Members,
				BackURL:      "http://localhost:8080/artists", // to go back to the website from artists_details
            })
            return
        }
    }

    http.Error(w, "Artist not found", http.StatusNotFound)
}

// function to geocoding the location from the map API
func GetCoordinates(address string) (float64, float64, error) {
	// get the API key
	apiKey := "34a441c385754c569b0b89e63fc51b85" 
	// get the URL from the origin of the API (OpenCage)
	baseURL := "https://api.opencagedata.com/geocode/v1/json"
	// set the parmaeters of the API like the adresse location, and the limit of use of API
	query := url.Values{}
	query.Set("q", address)
	query.Set("key", apiKey)
	query.Set("limit", "1")
	// create the URL, from the MAP API it should be like this : baseURL=apiKey (query)
	requestURL := fmt.Sprintf("%s?%s", baseURL, query.Encode())
	// if the URL isn't working, then it return an error
	resp, err := http.Get(requestURL)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	// get the geoResponse structs
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
	// if there is no results, then display this message
	if len(geoResponse.Results) == 0 {
		return 0, 0, fmt.Errorf("no results found for address: %s", address)
	}
	// we define the lattitude and longitude to geocode the locations
	lat := geoResponse.Results[0].Geometry.Lat
	lng := geoResponse.Results[0].Geometry.Lng
	return lat, lng, nil
}

// function forthe templates locations
func FetchLocationsForArtist(artistID int) ([]string, error) {
	// get the API locations
	response, err := http.Get(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/locations/%d", artistID))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	// get the location data structs
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
		// get the templates html, if there is no templates, then it diplay this message as an error
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
		// if the structs is not working, then it also display this message
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

    // Filtered the artists
    var filtered []Artist
	// search the artists
    for _, artist := range artists {
        if query != "" && !strings.Contains(strings.ToLower(artist.Name), query) &&
            !strings.Contains(strings.ToLower(strings.Join(artist.Members, " ")), query) {
            continue
        }
		// defined the variables to matche the dates and memebers
        matchesDate := dates == "" || containsDate(artist.Dates, dates)
        matchesMembers := memberCount == 0 || len(artist.Members) == memberCount
		// check if both dates and memeber matches
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
	// create the structs for artists summary details
    type ArtistSummary struct {
		ID        int      `json:"id"`
		Name      string   `json:"name"`
		Image     string   `json:"image"`
		Dates     []string `json:"dates"`
		Locations string   `json:"locations"`
		Members   []string `json:"members"`
	}
	// defined the details artists
	var artistSummaries []ArtistSummary
	// filtered only these parameters from artists
	for _, artist := range filtered {
		artistSummaries = append(artistSummaries, ArtistSummary{
			ID:        artist.ID,
			Name:      artist.Name,
			Image:     artist.Image,
			Dates:     artist.Dates,
			Locations: artist.Locations,
			Members:   artist.Members,
		})
	}
	// get the data artists summary
	data := struct {
		Artists []ArtistSummary
	}{
		Artists: artistSummaries,
	}
	// if the templates is not rendering then it display this message
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
