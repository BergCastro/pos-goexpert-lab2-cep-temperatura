package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Location struct {
    City string `json:"localidade"`
}

type Temperature struct {
    TempC float64 `json:"temp_c"`
    TempF float64 `json:"temp_f"`
    TempK float64 `json:"temp_k"`
}

func main() {
	err := godotenv.Load()
    if err != nil {
        log.Println("Error loading .env file")
    }

    http.HandleFunc("/", handleRequest)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
    zipcode := r.URL.Query().Get("zipcode")
    if len(zipcode) != 8 {
        http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
        return
    }

    location, err := getLocation(zipcode)
    if err != nil {
        http.Error(w, "can not find zipcode", http.StatusNotFound)
        return
    }

    temperature, err := getTemperature(location.City)
    if err != nil {
        http.Error(w, "failed to get temperature", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(temperature)
}

func getLocation(zipcode string) (*Location, error) {
    url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", zipcode)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var location Location
    err = json.Unmarshal(body, &location)
    if err != nil {
        return nil, err
    }

    if location.City == "" {
        return nil, fmt.Errorf("city not found for zipcode %s", zipcode)
    }

    return &location, nil
}

func getTemperature(city string) (*Temperature, error) {

    apiKey := os.Getenv("WEATHER_API_KEY")
    url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, city)
    resp, err := http.Get(url)
    if err != nil {
	
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var data map[string]interface{}
    err = json.Unmarshal(body, &data)
    if err != nil {
        return nil, err
    }

    current, ok := data["current"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid response from weather API")
    }

    tempC, ok := current["temp_c"].(float64)
    if !ok {
        return nil, fmt.Errorf("invalid temperature data from weather API")
    }

    tempF := tempC*1.8 + 32
    tempK := tempC + 273.15

    temperature := &Temperature{
        TempC: tempC,
        TempF: tempF,
        TempK: tempK,
    }

    return temperature, nil
}