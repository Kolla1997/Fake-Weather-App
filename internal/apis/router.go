package apis

import "net/http"

func RegisterRoutes() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", appHealthHandler)

	// Weather
	http.HandleFunc("/weather", weatherCollectionHandler) // GET list, POST batch
	http.HandleFunc("/weather/", weatherItemHandler)      // GET one: /weather/{state}
}
