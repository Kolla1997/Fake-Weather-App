package core

import (
	"FakeWeatherApp/internal/store"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const apiURL = "https://api.tomorrow.io/v4/weather/forecast"

var httpClient = &http.Client{Timeout: 5 * time.Second}

type WeatherResponse struct {
	Timelines struct {
		Minutely []struct {
			Time   string `json:"time"`
			Values struct {
				Temperature float64 `json:"temperature"`
				Humidity    float64 `json:"humidity"`
			} `json:"values"`
		} `json:"minutely"`
	} `json:"timelines"`
}

func FetchWeatherData(stateInput string) store.Weather {
	state, ok := store.NormalizeState(stateInput)
	if !ok {
		fmt.Println("Unknown state:", stateInput)
		return store.Weather{}
	}
	coords, ok := store.GetCoords(state)
	if !ok {
		fmt.Println("Missing coordinates for:", state)
		return store.Weather{}
	}

	apiKey := os.Getenv("TOMORROW_API_KEY")
	if apiKey == "" {
		fmt.Println("No TOMORROW_API_KEY set")
		return store.Weather{}
	}

	location := fmt.Sprintf("%f,%f", coords.Latitude, coords.Longitude)
	url := fmt.Sprintf("%s?location=%s&apikey=%s", apiURL, location, apiKey)

	resp, err := httpClient.Get(url)
	if err != nil {
		fmt.Println("Request error:", err)
		return store.Weather{}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Provider status:", resp.Status)
		return store.Weather{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read error:", err)
		return store.Weather{}
	}

	var apiResp WeatherResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Println("JSON parse error:", err)
		return store.Weather{}
	}
	if len(apiResp.Timelines.Minutely) == 0 {
		fmt.Println("No minutely data")
		return store.Weather{}
	}

	first := apiResp.Timelines.Minutely[0]
	weatherData := store.Weather{
		State:    state,
		Date:     store.TimeToChicago(first.Time),
		Temp:     first.Values.Temperature,
		Humidity: first.Values.Humidity,
		Source:   "tomorrow.io",
	}

	store.UpsertWeather(weatherData)
	return weatherData
}
