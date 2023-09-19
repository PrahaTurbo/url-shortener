package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func ExampleApplication_MakeURLHandler() {
	reader := strings.NewReader("https://yandex.ru")
	url := "http://127.0.0.1:8080/"

	r, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(r)
}

func ExampleApplication_GetOriginHandler() {
	url := "http://127.0.0.1:8080/abc-123"

	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(r)
}

func ExampleApplication_JSONHandler() {
	reader := strings.NewReader(`{"url": "https://yandex.ru"}`)
	url := "http://127.0.0.1:8080/api/shorten"

	r, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(r)
}

func ExampleApplication_BatchHandler() {
	type BatchRequest struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	urls := []BatchRequest{
		{
			CorrelationID: "1",
			OriginalURL:   "https://yandex.ru",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://google.com",
		},
	}

	body, err := json.Marshal(urls)
	if err != nil {
		log.Fatal(err)
	}

	reader := strings.NewReader(string(body))
	url := "http://127.0.0.1:8080/api/shorten/batch"

	r, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(r)
}

func ExampleApplication_GetUserURLsHandler() {
	url := "http://127.0.0.1:8080/api/user/urls"

	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(r)
}

func ExampleApplication_DeleteURLsHandler() {
	reader := strings.NewReader(`["abc-123", "xyz-234", "asd-345"]`)
	url := "http://127.0.0.1:8080/api/user/urls"

	r, err := http.NewRequest(http.MethodDelete, url, reader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(r)
}

func ExampleApplication_PingHandler() {
	url := "http://127.0.0.1:8080/ping"

	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(r)
}
