package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
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
func FetchLocationsForArtist(artistID int) ([]string, error) {
    response, err := http.Get(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/locations/%d", artistID))
    if err != nil {
        return nil, err
    }
    defer response.Body.Close()

    var locationData struct {
        Locations []string `json:"locations"`
    }
    err = json.NewDecoder(response.Body).Decode(&locationData)
    if err != nil {
        return nil, err
    }

    var locationsWithCoords []string
    for _, location := range locationData.Locations {
        lat, lng, err := GetCoordinates(location)
        if err == nil {
            // Ajoutez l'adresse formatée avec ses coordonnées
            locationsWithCoords = append(locationsWithCoords, fmt.Sprintf("%s (%.5f, %.5f)", location, lat, lng))
        } else {
            // Ajoutez uniquement l'adresse si les coordonnées ne peuvent pas être récupérées
            locationsWithCoords = append(locationsWithCoords, location)
        }
    }
    return locationsWithCoords, nil
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
	// Vérifier si un ID d'artiste est spécifié dans les paramètres de la requête
	idParam := r.URL.Query().Get("id") // Paramètre "id"

	if idParam == "" {
		// Pas de paramètre ID, afficher la liste des artistes
		displayArtistsList(w, r)
	} else {
		// Paramètre ID présent, afficher les détails de l'artiste
		displayArtistDetails(w, idParam)
	}
}

func containsDate(dates []string, targetDate string) bool {
	for _, date := range dates {
		if date >= targetDate {
			return true
		}
	}
	return false
}

// Affiche la liste des artistes
func displayArtistsList(w http.ResponseWriter, r *http.Request) {
	artists, err := FetchArtists()
	if err != nil {
		log.Printf("Error fetching artists: %v", err)
		http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
		return
	}

	// Récupérer les paramètres de recherche et de filtrage
	query := strings.ToLower(r.URL.Query().Get("q"))
	Dates := r.URL.Query().Get("dates")
	memberCount, _ := strconv.Atoi(r.URL.Query().Get("memberCount"))

	var filtered []Artist
	for _, artist := range artists {
		// Filtrage par recherche
		if query != "" {
			if !strings.Contains(strings.ToLower(artist.Name), query) &&
				!strings.Contains(strings.ToLower(strings.Join(artist.Members, " ")), query) {
				continue
			}
		}

		// Filtrage par date
		matchesDate := Dates == "" || containsDate(artist.Dates, Dates)

		// Filtrage par nombre de membres
		matchesMembers := memberCount == 0 || len(artist.Members) == memberCount

		// Ajouter à la liste filtrée si tous les critères sont respectés
		if matchesDate && matchesMembers {
			filtered = append(filtered, artist)
		}
	}

	// Rendre le template avec les artistes filtrés
	tmpl, err := template.ParseFiles("web/html/artists.html")
	if err != nil {
		log.Printf("Error loading template: %v", err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

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

// Affiche les détails d'un artiste spécifique
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

/*func ArtistsHandler(w http.ResponseWriter, r *http.Request) {
    artists, err := FetchArtists()
    if err != nil {
        log.Printf("Error fetching artists: %v", err)
        http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
        return
    }

    // Récupérer les paramètres de recherche et de filtrage
    query := strings.ToLower(r.URL.Query().Get("q"))
    Dates := r.URL.Query().Get("dates")
    memberCount, _ := strconv.Atoi(r.URL.Query().Get("memberCount"))

    var filtered []Artist
    for _, artist := range artists {
        // Filtrage par recherche
        if query != "" {
            if !strings.Contains(strings.ToLower(artist.Name), query) &&
                !strings.Contains(strings.ToLower(strings.Join(artist.Members, " ")), query) {
                continue
            }
        }

        // Filtrage par date
        matchesDate := Dates == "" || containsDate(artist.Dates, Dates)

        // Filtrage par nombre de membres
        matchesMembers := memberCount == 0 || len(artist.Members) == memberCount

        // Ajouter à la liste filtrée si tous les critères sont respectés
        if matchesDate && matchesMembers {
            filtered = append(filtered, artist)
        }
    }

    // Rendre le template avec les artistes filtrés
    tmpl, err := template.ParseFiles("web/html/artists.html")
    if err != nil {
        log.Printf("Error loading template: %v", err)
        http.Error(w, "Unable to load template", http.StatusInternalServerError)
        return
    }

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

func containsDate(dates []string, targetDate string) bool {
    for _, date := range dates {
        if date >= targetDate {
            return true
        }
    }
    return false
}

func ArtistDetailsHandler(w http.ResponseWriter, r *http.Request) {
    // Récupérer l'ID de l'artiste à partir de l'URL
    idStr := strings.TrimPrefix(r.URL.Path, "/artists/")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid artist ID", http.StatusBadRequest)
        return
    }

    // Récupérer les détails de l'artiste
    artists, err := FetchArtists()
    if err != nil {
        http.Error(w, "Unable to fetch artists", http.StatusInternalServerError)
        return
    }

    // Chercher l'artiste correspondant à l'ID
    for _, artist := range artists {
        if artist.ID == id {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(artist)
            return
        }
    }

    http.Error(w, "Artist not found", http.StatusNotFound)
}*/
/*func FetchLocationsForArtist(artistID int) ([]string, error) {
	response, err := http.Get(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/locations/%d", artistID))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var locationData struct {
		Locations []string `json:"locations"`
	}
	err = json.NewDecoder(response.Body).Decode(&locationData)
	if err != nil {
		return nil, err
	}
	return locationData.Locations, nil
}*/
type GeocodingResponse struct {
    Results []struct {
        Geometry struct {
            Location struct {
                Lat float64 `json:"lat"`
                Lng float64 `json:"lng"`
            } `json:"location"`
        } `json:"geometry"`
    } `json:"results"`
}

func GetCoordinates(address string) (float64, float64, error) {
    apiKey := "CLE_API"
    url := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?address=%s&key=%s", address, apiKey)

    resp, err := http.Get(url)
    if err != nil {
        return 0, 0, err
    }
    defer resp.Body.Close()

    var geoResponse GeocodingResponse
    if err := json.NewDecoder(resp.Body).Decode(&geoResponse); err != nil {
        return 0, 0, err
    }

    if len(geoResponse.Results) == 0 {
        return 0, 0, fmt.Errorf("no results found for address: %s", address)
    }

    return geoResponse.Results[0].Geometry.Location.Lat, geoResponse.Results[0].Geometry.Location.Lng, nil
}