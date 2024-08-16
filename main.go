package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var CONTEXT context.Context
var FIREBASE_APP *firebase.App
var AUTH_CLIENT *auth.Client
var FIRESTORE_CLIENT *firestore.Client

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

func initializeAuth() {
	var err error
	AUTH_CLIENT, err = FIREBASE_APP.Auth(CONTEXT)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Initialized Authentication")
}

func initializeFirestore() {
	var err error
	FIRESTORE_CLIENT, err = FIREBASE_APP.Firestore(CONTEXT)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Initialized Firestore")
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
	initializeAuth()
	initializeFirestore()

	http.HandleFunc("/ping", handlePing)

	log.Fatalln(http.ListenAndServe(":80", nil))
}
