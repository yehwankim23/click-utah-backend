package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var Context context.Context
var AuthClient *auth.Client

var Mutexes = make(map[string]*sync.Mutex)

type TimeInfo struct {
	Time string `json:"time"`
}

type SignUpInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type FirebaseInfo struct {
	Error FirebaseError `json:"error"`
}

type FirebaseError struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Errors  []FirebaseErrors `json:"errors"`
}

type FirebaseErrors struct {
	Message string `json:"message"`
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
}

type UserInfo struct {
	Name      string `json:"name"`
	Count     int    `json:"count"`
	Timestamp int64  `json:"timestamp"`
}

type UIDInfo struct {
	UID string `json:"uid"`
}

type CountInfo struct {
	Count int `json:"count"`
}

type RenameInfo struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

type NameInfo struct {
	Name string `json:"name"`
}

func getMutex(fileName string) *sync.Mutex {
	mutex, exists := Mutexes[fileName]

	if exists {
		return mutex
	}

	mutex = &sync.Mutex{}
	Mutexes[fileName] = mutex
	return mutex
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

	email := signUpInfo.Email
	password := signUpInfo.Password
	name := signUpInfo.Name

	if email == "" || password == "" || name == "" {
		http.Error(responseWriter, "Missing email, password, and/or name", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(email, "utah.edu") {
		http.Error(responseWriter, "Invalid email domain", http.StatusBadRequest)
		return
	}

	userRecord, err := AuthClient.CreateUser(
		Context,
		(&auth.UserToCreate{}).Email(email).Password(password),
	)

	if err != nil {
		errString := err.Error()
		_, after, found := strings.Cut(errString, "body: ")

		if !found {
			http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
			return
		}

		var firebaseInfo FirebaseInfo
		err := json.Unmarshal([]byte(after), &firebaseInfo)

		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		}

		firebaseError := firebaseInfo.Error
		http.Error(responseWriter, firebaseError.Message, firebaseError.Code)
		return
	}

	userBytes, err := json.Marshal(UserInfo{
		Name:      name,
		Count:     0,
		Timestamp: time.Now().UnixMicro(),
	})

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	uid := userRecord.UID
	fileName := "./data/" + uid + ".json"
	mutex := getMutex(fileName)
	mutex.Lock()
	defer mutex.Unlock()
	err = os.WriteFile(fileName, userBytes, 0600)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	uidBytes, err := json.Marshal(UIDInfo{
		UID: uid,
	})

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	_, err = responseWriter.Write(uidBytes)

	if err != nil {
		innerErr := AuthClient.DeleteUser(Context, uid)

		if innerErr != nil {
			errString := innerErr.Error()
			_, after, found := strings.Cut(errString, "body: ")

			if !found {
				http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
				return
			}

			var firebaseInfo FirebaseInfo
			err := json.Unmarshal([]byte(after), &firebaseInfo)

			if err != nil {
				http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
			}

			firebaseError := firebaseInfo.Error
			http.Error(responseWriter, firebaseError.Message, firebaseError.Code)
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

	uid := uidInfo.UID

	if uid == "" {
		http.Error(responseWriter, "Missing UID", http.StatusBadRequest)
		return
	}

	fileName := "./data/" + uid + ".json"
	mutex := getMutex(fileName)
	mutex.Lock()
	defer mutex.Unlock()
	dataBytes, err := os.ReadFile(fileName)

	if err != nil {
		http.Error(responseWriter, "Invalid UID", http.StatusBadRequest)
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

	uid := uidInfo.UID

	if uid == "" {
		http.Error(responseWriter, "Missing UID", http.StatusBadRequest)
		return
	}

	fileName := "./data/" + uid + ".json"
	mutex := getMutex(fileName)
	mutex.Lock()
	defer mutex.Unlock()
	userBytes, err := os.ReadFile(fileName)

	if err != nil {
		http.Error(responseWriter, "Invalid UID", http.StatusBadRequest)
		return
	}

	var userInfo UserInfo
	err = json.Unmarshal(userBytes, &userInfo)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo.Count++
	userInfo.Timestamp = time.Now().UnixMicro()
	userBytes, err = json.Marshal(userInfo)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	err = os.WriteFile(fileName, userBytes, 0600)

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

func handleRename(responseWriter http.ResponseWriter, request *http.Request) {
	var renameInfo RenameInfo
	err := json.NewDecoder(request.Body).Decode(&renameInfo)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	uid := renameInfo.UID
	name := renameInfo.Name

	if uid == "" || name == "" {
		http.Error(responseWriter, "Missing UID and/or name", http.StatusBadRequest)
		return
	}

	fileName := "./data/" + uid + ".json"
	mutex := getMutex(fileName)
	mutex.Lock()
	defer mutex.Unlock()
	userBytes, err := os.ReadFile(fileName)

	if err != nil {
		http.Error(responseWriter, "Invalid UID", http.StatusBadRequest)
		return
	}

	var userInfo UserInfo
	err = json.Unmarshal(userBytes, &userInfo)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo.Name = name
	userBytes, err = json.Marshal(userInfo)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	err = os.WriteFile(fileName, userBytes, 0600)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	nameBytes, err := json.Marshal(NameInfo{
		Name: name,
	})

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	_, err = responseWriter.Write(nameBytes)

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
	http.HandleFunc("/rename", handleRename)

	log.Fatalln(http.ListenAndServe(":80", nil))
}
