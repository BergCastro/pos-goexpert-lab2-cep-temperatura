package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Location struct {
    City string `json:"localidade"`
}

type Temperature struct {
    City  string  `json:"city"`
    TempC float64 `json:"temp_c"`
    TempF float64 `json:"temp_f"`
    TempK float64 `json:"temp_k"`
}

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Println("Error loading .env file")
    }


	exporter, err := zipkin.New("http://zipkin:9411/api/v2/spans")
    if err != nil {
        log.Fatal(err)
    }
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(100),
		),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )
    defer func() {
        if err := tp.Shutdown(context.Background()); err != nil {
            log.Fatal(err)
        }
    }()
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

    http.HandleFunc("/temperature", handleTemperatureRequest)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleTemperatureRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    tracer := otel.Tracer("service-b")
    ctx, span := tracer.Start(ctx, "handleTemperatureRequest")
    defer span.End()

    zipcode := r.URL.Query().Get("zipcode")
    if len(zipcode) != 8 {
        http.Error(w, "Invalid zipcode", http.StatusUnprocessableEntity)
        return
    }

    location, err := getLocation(ctx, zipcode)
    if err != nil {
        http.Error(w, "Can not find zipcode", http.StatusNotFound)
        return
    }

    temperature, err := getTemperature(ctx, location.City)
    if err != nil {
        http.Error(w, "Failed to get temperature", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(temperature)
}

func getLocation(ctx context.Context, zipcode string) (*Location, error) {
    ctx, span := otel.Tracer("service-b").Start(ctx, "getLocation")
    defer span.End()

    url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", zipcode)
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }

    client := &http.Client{}
    resp, err := client.Do(req)
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

    span.SetAttributes(attribute.String("city", location.City))

    return &location, nil
}

func getTemperature(ctx context.Context, city string) (*Temperature, error) {
    ctx, span := otel.Tracer("service-b").Start(ctx, "getTemperature")
    defer span.End()

    apiKey := os.Getenv("WEATHER_API_KEY")
    url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, city)
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }

    client := &http.Client{}
    resp, err := client.Do(req)
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
        City:  city,
        TempC: tempC,
        TempF: tempF,
        TempK: tempK,
    }

    span.SetAttributes(
        attribute.Float64("temp_c", tempC),
        attribute.Float64("temp_f", tempF),
        attribute.Float64("temp_k", tempK),
    )

    return temperature, nil
}
