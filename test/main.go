package main

import (
	"fmt"
	"net/http"
	"io"
)

func main() {

	url := "https://google-map-places.p.rapidapi.com/maps/api/geocode/json?address=1600%20Amphitheatre%2BParkway%2C%20Mountain%20View%2C%20CA&language=en&region=en&result_type=administrative_area_level_1&location_type=APPROXIMATE"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("x-rapidapi-key", "7a2cfcfda4msh2f03e4de2794082p1b4d77jsnac469a73d4b2")
	req.Header.Add("x-rapidapi-host", "google-map-places.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))

}