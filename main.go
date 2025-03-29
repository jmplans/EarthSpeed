package main

import (
	"fmt"
	"os"
	"math"
	"net/http"
	"net/url"
	"encoding/json"
	"io/ioutil"
)

type OpenCageResponse struct {
	Results []struct {
		Geometry struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geometry"`
	} `json:"results"`
}

func getCountryLat(country, apiKey string) (float64, error) {
	endpoint := "https://api.opencagedata.com/geocode/v1/json"
	params := url.Values{}
	params.Add("q", country)
	params.Add("key", apiKey)

	resp, err := http.Get(endpoint + "?" + params.Encode())
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var data OpenCageResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}

	if len(data.Results) == 0 {
		return 0, fmt.Errorf("no result found")
	}

	return data.Results[0].Geometry.Lat, nil
}

func getCountrySpeed(country string) float64 {
	speed := 0.0
	apiKey := "" // add your OpenCage API key here

	lat, err := getCountryLat(country, apiKey)
	if err != nil {
		fmt.Println("Error: ", err)
		return speed
	}

	equatorialKM := 40075.017
	equatorialSpeedKMH := equatorialKM / 24
	latInRadians := lat * math.Pi / 180

	speed = equatorialSpeedKMH * math.Cos(latInRadians) 

	return speed
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <country>")
		return
	}
	country := os.Args[1]
	speed := getCountrySpeed(country)

	fmt.Printf("%s's rotational speed: %.2f km/h\n", country, math.Abs(speed))
}
