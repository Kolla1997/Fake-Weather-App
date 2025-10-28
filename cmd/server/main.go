package main

import (
	"FakeWeatherApp/internal/apis"
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Sping up the Fake Weather Application!")

	apis.FakeWaterRouter()
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error while sping up the app: ", err)
		return
	}
}
