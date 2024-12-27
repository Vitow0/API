package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	apiURL    = "https://google-map-places.p.rapidapi.com/maps/api/geocode/json"
	apiKey    = "7a2cfcfda4msh2f03e4de2794082p1b4d77jsnac469a73d4b2" 
	apiHost   = "google-map-places.p.rapidapi.com"
)

// Struct pour décoder la réponse JSON
type GeocodeResponse struct {
	Results []struct {
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
	Status string `json:"status"`
}

func FetchGeocode(address string) (*GeocodeResponse, error) {
	// Construire l'URL avec les paramètres
	fullURL := fmt.Sprintf("%s?address=%s", apiURL, address)

	// Créer une requête HTTP
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// Ajouter les en-têtes nécessaires
	req.Header.Add("x-rapidapi-key", apiKey)
	req.Header.Add("x-rapidapi-host", apiHost)

	// Envoyer la requête
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Lire et décoder la réponse
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var geocodeResponse GeocodeResponse
	err = json.Unmarshal(body, &geocodeResponse)
	if err != nil {
		return nil, err
	}

	return &geocodeResponse, nil
}

func main() {
	// Exemple d'utilisation
	address := "1600 Amphitheatre Parkway, Mountain View, CA"
	response, err := FetchGeocode(address)
	if err != nil {
		fmt.Println("Erreur lors de la récupération des données :", err)
		return
	}

	for _, result := range response.Results {
		fmt.Printf("Adresse : %s\n", result.FormattedAddress)
		fmt.Printf("Latitude : %f, Longitude : %f\n", result.Geometry.Location.Lat, result.Geometry.Location.Lng)
	}
}
