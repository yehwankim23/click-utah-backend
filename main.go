package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func handlePing(responseWriter http.ResponseWriter, request *http.Request) {
	location, err := time.LoadLocation("Asia/Seoul")

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(map[string]string{
		"pong": time.Now().In(location).Format(time.DateTime),
	})

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	_, err = responseWriter.Write(response)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/ping", handlePing)

	log.Fatalln(http.ListenAndServe(":80", nil))
}
