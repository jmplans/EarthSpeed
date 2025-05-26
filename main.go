package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
)

// Geodetic constants – WGS-84 reference ellipsoid (kilometres)
const (
	a        = 6378.137              // semi-major axis
	b        = 6356.752              // semi-minor axis
	e2       = (a*a - b*b) / (a * a) // first eccentricity squared
	dayHours = 23.934444             //24.0                  // mean solar day
)

type OpenCageResponse struct {
	Results []struct {
		Geometry struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geometry"`
	} `json:"results"`
}

type CountryLocation struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type LocationCache map[string]CountryLocation

func LoadCache(filename string) (LocationCache, error) {
	cache := make(LocationCache)

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return cache, nil
}

func SaveCache(filename string, newEntry map[string]CountryLocation) error {
	cache, err := LoadCache(filename)
	if err != nil {
		return err
	}
	for k, v := range newEntry {
		cache[k] = v
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0o644)
}

// Horizontal distance from Earth’s spin axis at latitude φ (degrees)
func radiusAtLatitude(phiDeg float64) float64 {
	phi := phiDeg * math.Pi / 180
	N := a / math.Sqrt(1-e2*math.Sin(phi)*math.Sin(phi))
	return N * math.Cos(phi) // km
}

// OpenCage lookup (lat/long for a country name)
func fetchCountryLat(country, apiKey string) (float64, error) {
	endpoint := "https://api.opencagedata.com/geocode/v1/json"
	params := url.Values{}
	params.Add("q", country)
	params.Add("key", apiKey)

	resp, err := http.Get(endpoint + "?" + params.Encode())
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data OpenCageResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}
	if len(data.Results) == 0 {
		return 0, fmt.Errorf("no result found for %q", country)
	}

	loc := CountryLocation{
		Lat: data.Results[0].Geometry.Lat,
		Lng: data.Results[0].Geometry.Lng,
	}

	if err := SaveCache("country_cache.json", map[string]CountryLocation{country: loc}); err != nil {
		fmt.Println("warning: cache save failed:", err)
	}
	return loc.Lat, nil
}

// Core routine – rotational speed at the given country’s latitude
func getCountrySpeed(country string) float64 {
	cacheFile := "country_cache.json"
	cache, _ := LoadCache(cacheFile)

	loc, ok := cache[country]
	if !ok {
		fmt.Println("Fetching latitude from OpenCage…")
		apiKey := "" // your OpenCage API key
		lat, err := fetchCountryLat(country, apiKey)
		if err != nil {
			fmt.Println("Error:", err)
			return 0
		}
		loc.Lat = lat
	}

	r := radiusAtLatitude(loc.Lat)
	return math.Abs(2 * math.Pi * r / dayHours) // km h⁻¹
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <country>")
		return
	}

	country := os.Args[1]
	speed := getCountrySpeed(country)
	if speed == 0 {
		return
	}
	fmt.Printf("%s’s rotational speed: %.2f km/h\n", country, speed)
}
