package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ZipCodeRequest struct {
    ZipCode string `json:"cep"`
}

func main() {

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

    http.HandleFunc("/zipcode", handleZipCodeRequest)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleZipCodeRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    _, span := otel.Tracer("service-a").Start(ctx, "handleZipCodeRequest")
    defer span.End()

    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req ZipCodeRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    if len(req.ZipCode) != 8 {
        http.Error(w, "Invalid zipcode", http.StatusUnprocessableEntity)
        return
    }


    resp, err := http.Get("http://service-b:8081/temperature?zipcode=" + req.ZipCode)
    if err != nil {
        http.Error(w, "Failed to get temperature", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(resp.StatusCode)
	_, err = w.Write(body)
	if err != nil {
		log.Println("Failed to write response:", err)
	}
}