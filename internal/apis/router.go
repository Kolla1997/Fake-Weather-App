package apis

import (
	"net/http"
)

func FakeWaterRouter() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", appHealthHandler)
	http.HandleFunc("/getWeather", weatherHadhler)
	http.HandleFunc("/getWeather/", postWeather)
}
