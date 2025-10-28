package apis

import (
	"FakeWeatherApp/internal/core"
	"FakeWeatherApp/internal/store"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type multCities struct {
	Cities []string `json:"cities"`
}

var (
	workers int = 3
	wg      sync.WaitGroup
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to fake weather application!")

}

func appHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Status:", http.StateActive)
}

func getWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(store.UpdatedStates)

}

func stateTitled(state string) string {
	caser := cases.Title(language.English)
	state = caser.String(state)
	fmt.Printf("Fetching weather data for state: %s\n", state)
	return state
}

func postWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var state string

	pathParts := strings.Split(strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/data")), "/")
	if len(pathParts) < 2 || pathParts[2] == "" {
		http.Error(w, "Bad Request: state not provided in path", http.StatusBadRequest)
		return
	}
	// 3. Extract the state from the second segment.
	state = stateTitled(pathParts[2])
	response := core.FetchWeatherData(state)
	json.NewEncoder(w).Encode(response)

}

func postCitiesWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var respCities multCities
	if err := json.NewDecoder(r.Body).Decode(&respCities); err != nil {
		fmt.Println(err)
		http.Error(w, "Error while parsing request", http.StatusBadRequest)
		return
	}
	jobs := make(chan string, len(respCities.Cities))
	results := make(chan store.Weather)
	fmt.Println(respCities.Cities)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				fmt.Println("-->", job)
				results <- core.FetchWeatherData(stateTitled(job))
			}
		}()
	}

	for _, city := range respCities.Cities {
		jobs <- city
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(results)
	}()
	out := make([]store.Weather, 0, len(respCities.Cities))
	for result := range results {
		out = append(out, result)
	}
	if err := json.NewEncoder(w).Encode(out); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}

}

func weatherHadhler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getWeather(w, r)
	case http.MethodPost:
		postCitiesWeather(w, r)

	}
}
