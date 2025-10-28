package apis

import (
	"FakeWeatherApp/internal/core"
	"FakeWeatherApp/internal/store"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type multCities struct {
	Cities []string `json:"cities"`
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to fake weather application!")
}

func appHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "FakeWeatherApp"})
}

// GET /weather -> all cached
// POST /weather -> {"cities":["Texas","Illinois"]}
func weatherCollectionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		snapshot := store.SnapshotUpdatedStates() // safe copy
		json.NewEncoder(w).Encode(snapshot)
		return

	case http.MethodPost:
		var payload multCities
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		if len(payload.Cities) == 0 {
			http.Error(w, `{"error":"cities cannot be empty"}`, http.StatusBadRequest)
			return
		}

		// worker pool
		workers := 3
		jobs := make(chan string)
		results := make(chan store.Weather, len(payload.Cities))

		for i := 0; i < workers; i++ {
			go func() {
				for city := range jobs {
					results <- core.FetchWeatherData(city)
				}
			}()
		}
		go func() {
			for _, c := range payload.Cities {
				jobs <- c
			}
			close(jobs)
		}()

		// collect with a timeout safeguard
		out := make([]store.Weather, 0, len(payload.Cities))
		timeout := time.After(8 * time.Second)
		for i := 0; i < len(payload.Cities); i++ {
			select {
			case r := <-results:
				out = append(out, r)
			case <-timeout:
				http.Error(w, `{"error":"timeout fetching some cities"}`, http.StatusGatewayTimeout)
				return
			}
		}
		json.NewEncoder(w).Encode(out)
		return

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// GET /weather/{state}
func weatherItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	// "/weather/{state}"
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/weather/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, `{"error":"state required"}`, http.StatusBadRequest)
		return
	}
	state := parts[0]
	data := core.FetchWeatherData(state)
	if (data == store.Weather{}) {
		http.Error(w, `{"error":"state not found or provider error"}`, http.StatusBadGateway)
		return
	}
	json.NewEncoder(w).Encode(data)
}
