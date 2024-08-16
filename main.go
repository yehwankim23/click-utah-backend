package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var CONTEXT context.Context
var FIREBASE_APP *firebase.App

func initializeFirebase() {
	var err error

	FIREBASE_APP, err = firebase.NewApp(
		CONTEXT,
		nil,
		option.WithCredentialsFile("click-utah-d7eaac9e6640.json"),
	)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Initialized Firebase")
}

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
	CONTEXT = context.Background()
	initializeFirebase()

	http.HandleFunc("/ping", handlePing)

	log.Fatalln(http.ListenAndServe(":80", nil))
}
