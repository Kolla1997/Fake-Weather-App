package core

import (
	"FakeWeatherApp/internal/store"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	apiURL = "https://api.tomorrow.io/v4/weather/forecast"
	apiKey = "cplb3zlYdrVfAIYKJJr0uw7SvsxMpmXU"
)

var location string

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

func timeToDate(utcTimeStr string) string {
	utcTime, err := time.Parse(time.RFC3339, utcTimeStr)
	if err != nil {
		panic(err)
	}
	cstLocation, err := time.LoadLocation("America/Chicago")
	if err != nil {
		panic(err)
	}
	cstTime := utcTime.In(cstLocation)
	cstString := cstTime.Format("2006-01-02 15:04:05")
	return cstString
}

func FetchWeatherData(state string) store.Weather {
	if state == "" {
		fmt.Println("Empty states passed")
		return store.Weather{}
	}

	// var stateCodes store.Coordinates
	stateCodes := store.StatesData.AllStates[state]
	fmt.Println(stateCodes)
	location = fmt.Sprint(stateCodes.Latitude) + "," + fmt.Sprint(stateCodes.Longitude)
	URL := fmt.Sprintf("%s?location=%s&apikey=%s", apiURL, location, apiKey)
	fmt.Println(URL)
	resp, err := http.Get(URL)
	if err != nil {
		fmt.Println("Error while requesting API: ", err)
		return store.Weather{}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading body:", err)
		return store.Weather{}
	}
	var apiResp WeatherResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Println("Failed to parse JSON:", err)
		return store.Weather{}
	}

	first := apiResp.Timelines.Minutely[0]
	weatherData := store.Weather{
		State:    state,
		Date:     timeToDate(first.Time),
		Temp:     first.Values.Temperature,
		Humidity: first.Values.Humidity,
		Source:   "tomorrow.io",
	}
	store.MU.Lock()
	// store.UpdatedStates = append(store.UpdatedStates, weatherData)
	store.UpdatedStates[state] = weatherData
	store.MU.Unlock()
	return weatherData
}
