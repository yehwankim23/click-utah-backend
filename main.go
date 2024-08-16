package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
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

func handleSignUp(responseWriter http.ResponseWriter, request *http.Request) {
	type RequestBody struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
	}

	var requestBody RequestBody
	err := json.NewDecoder(request.Body).Decode(&requestBody)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	if !strings.HasSuffix(requestBody.Email, "utah.edu") {
		http.Error(responseWriter, "Invalid email", http.StatusBadRequest)
		return
	}

	userRecord, err := AUTH_CLIENT.CreateUser(
		CONTEXT,
		(&auth.UserToCreate{}).
			Email(requestBody.Email).
			Password(requestBody.Password).
			DisplayName(requestBody.DisplayName),
	)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	uid := userRecord.UID

	_, err = FIRESTORE_CLIENT.Collection("users").Doc(uid).Set(CONTEXT, map[string]int{
		"count": 0,
	})

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		err := AUTH_CLIENT.DeleteUser(CONTEXT, uid)

		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}
}

func handleUser(responseWriter http.ResponseWriter, request *http.Request) {
	type RequestBody struct {
		Uid string `json:"uid"`
	}

	var requestBody RequestBody
	err := json.NewDecoder(request.Body).Decode(&requestBody)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	documentSnapshot, err := FIRESTORE_CLIENT.Collection("users").Doc(requestBody.Uid).Get(CONTEXT)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(documentSnapshot.Data())

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
	defer FIRESTORE_CLIENT.Close()

	http.HandleFunc("/ping", handlePing)
	http.HandleFunc("/signup", handleSignUp)
	http.HandleFunc("/user", handleUser)

	log.Fatalln(http.ListenAndServe(":80", nil))
}
