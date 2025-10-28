package main

import (
	"FakeWeatherApp/internal/apis"
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Starting the Fake Weather Application!")
	apis.RegisterRoutes()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error while starting the app:", err)
	}
}
