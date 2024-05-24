package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {

	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}


	code := m.Run()
	os.Exit(code)
}

func TestHandleRequest_ValidZipcode(t *testing.T) {
	
	req, err := http.NewRequest("GET", "/?zipcode=60125001", nil)
	if err != nil {
		t.Fatal(err)
	}


	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRequest)


	handler.ServeHTTP(rr, req)


	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}


	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, expectedContentType)
	}
}

func TestHandleRequest_InvalidZipcode(t *testing.T) {
	
	req, err := http.NewRequest("GET", "/?zipcode=123", nil)
	if err != nil {
		t.Fatal(err)
	}


	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRequest)


	handler.ServeHTTP(rr, req)


	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}

	expectedErrorMessage := "invalid zipcode\n"
	if rr.Body.String() != expectedErrorMessage {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedErrorMessage)
	}
}

func TestHandleRequest_NotFoundZipcode(t *testing.T) {
	
	req, err := http.NewRequest("GET", "/?zipcode=99999999", nil)
	if err != nil {
		t.Fatal(err)
	}


	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRequest)


	handler.ServeHTTP(rr, req)


	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}


	expectedErrorMessage := "can not find zipcode\n"
	if rr.Body.String() != expectedErrorMessage {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedErrorMessage)
	}
}