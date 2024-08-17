package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var Context context.Context
var AuthClient *auth.Client

type TimeInfo struct {
	Time string `json:"time"`
}

type SignUpInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type UserInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type UIDInfo struct {
	UID string `json:"uid"`
}

type CountInfo struct {
	Count int `json:"count"`
}

func initializeFirebaseAuth() {
	Context = context.Background()
	var err error

	firebaseApp, err := firebase.NewApp(
		Context,
		nil,
		option.WithCredentialsFile("click-utah-d7eaac9e6640.json"),
	)

	if err != nil {
		log.Fatalln(err)
	}

	AuthClient, err = firebaseApp.Auth(Context)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Initialized Firebase Auth")
}

func handleTime(responseWriter http.ResponseWriter, request *http.Request) {
	timeBytes, err := json.Marshal(TimeInfo{
		Time: time.Now().UTC().Add(time.Hour * 9).Format(time.DateTime),
	})

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	_, err = responseWriter.Write(timeBytes)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleSignUp(responseWriter http.ResponseWriter, request *http.Request) {
	var signUpInfo SignUpInfo
	err := json.NewDecoder(request.Body).Decode(&signUpInfo)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(signUpInfo.Email, "utah.edu") {
		http.Error(responseWriter, "Invalid email", http.StatusBadRequest)
		return
	}

	userRecord, err := AuthClient.CreateUser(
		Context,
		(&auth.UserToCreate{}).Email(signUpInfo.Email).Password(signUpInfo.Password),
	)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	userBytes, err := json.Marshal(UserInfo{
		Name:  signUpInfo.Name,
		Count: 0,
	})

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	uid := userRecord.UID
	err = os.WriteFile("./data/"+uid+".json", userBytes, 0600)

	uidBytes, err := json.Marshal(UIDInfo{
		UID: uid,
	})

	responseWriter.Header().Set("Content-Type", "application/json")
	_, err = responseWriter.Write(uidBytes)

	if err != nil {
		innerErr := AuthClient.DeleteUser(Context, uid)

		if innerErr != nil {
			http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleUser(responseWriter http.ResponseWriter, request *http.Request) {
	var uidInfo UIDInfo
	err := json.NewDecoder(request.Body).Decode(&uidInfo)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	dataBytes, err := os.ReadFile("./data/" + uidInfo.UID + ".json")

	if err != nil {
		http.Error(responseWriter, "Invalid uid", http.StatusBadRequest)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	_, err = responseWriter.Write(dataBytes)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleClick(responseWriter http.ResponseWriter, request *http.Request) {
	var uidInfo UIDInfo
	err := json.NewDecoder(request.Body).Decode(&uidInfo)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	userBytes, err := os.ReadFile("./data/" + uidInfo.UID + ".json")

	if err != nil {
		http.Error(responseWriter, "Invalid uid", http.StatusBadRequest)
		return
	}

	var userInfo UserInfo
	err = json.Unmarshal(userBytes, &userInfo)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo.Count++
	userBytes, err = json.Marshal(userInfo)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	err = os.WriteFile("./data/"+uidInfo.UID+".json", userBytes, 0600)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	countBytes, err := json.Marshal(CountInfo{
		Count: userInfo.Count,
	})

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	_, err = responseWriter.Write(countBytes)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	initializeFirebaseAuth()

	http.HandleFunc("/time", handleTime)
	http.HandleFunc("/signup", handleSignUp)
	http.HandleFunc("/user", handleUser)
	http.HandleFunc("/click", handleClick)

	log.Fatalln(http.ListenAndServe(":80", nil))
}
