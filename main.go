package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var Mutexes = make(map[string]*sync.Mutex)

type TimeJson struct {
	Time    string `json:"time"`
	Version string `json:"version"`
}

type ErrorJson struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type SignUpJson struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type DataJson struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	Uid       string `json:"uid"`
	Count     int    `json:"count"`
	Timestamp int64  `json:"timestamp"`
}

type UidJson struct {
	Uid string `json:"uid"`
}

type SignInJson struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserJson struct {
	Error bool   `json:"error"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Uid   string `json:"uid"`
	Count int    `json:"count"`
}

type RenameJson struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}

func logError(level string, function string, code string, message string) {
	log.Println(level + " " + function + " " + code + " " + message)
}

func handleError(responseWriter http.ResponseWriter, level string, function string, code string,
	message string) {
	if level == "" || function == "" || (level != "2" && code == "") {
		logError("5", "sendError", "level == '' || function == '' || (level != '2' && code == '')",
			"")

		return
	}

	logError(level, function, code, message)

	if level != "4" {
		return
	}

	errorBytes, err := json.Marshal(ErrorJson{
		Error:   true,
		Message: message,
	})

	if err != nil {
		logError("5", "sendError", "Marshal(ErrorJson)", "")
		return
	}

	responseWriter.Header().Add("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Add("Content-Type", "application/json")
	_, err = responseWriter.Write(errorBytes)

	if err != nil {
		logError("5", "sendError", "Write(errorBytes)", "")
		return
	}

	logError("2", "sendError", "", "")
}

func handleTime(responseWriter http.ResponseWriter, request *http.Request) {
	function := "handleTime"

	timeBytes, err := json.Marshal(TimeJson{
		Time:    time.Now().UTC().Add(time.Hour * 9).Format(time.DateTime),
		Version: "2024.10.25.0",
	})

	if err != nil {
		handleError(responseWriter, "5", function, "Marshal(TimeJson)", "")
		return
	}

	responseWriter.Header().Add("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Add("Content-Type", "application/json")
	_, err = responseWriter.Write(timeBytes)

	if err != nil {
		handleError(responseWriter, "5", function, "Write(timeBytes)", "")
		return
	}

	handleError(responseWriter, "2", function, "", "")
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

func handleSignUp(responseWriter http.ResponseWriter, request *http.Request) {
	function := "handleSignUp"

	var signUpJson SignUpJson
	err := json.NewDecoder(request.Body).Decode(&signUpJson)

	if err != nil {
		handleError(responseWriter, "4", function, "Decode(&signUpJson)", "Request body is invalid")
		return
	}

	email := signUpJson.Email
	password := signUpJson.Password
	name := signUpJson.Name

	if email == "" || password == "" || name == "" {
		handleError(responseWriter, "4", function, "email == '' || password == '' || name == ''",
			"Email, password, or name is empty")

		return
	}

	uid, domain, found := strings.Cut(email, "@")

	if !found || uid == "" {
		handleError(responseWriter, "4", function, "!found || uid == ''", "Email is invalid")
		return
	}

	if !strings.HasSuffix(domain, "utah.edu") {
		handleError(responseWriter, "4", function, "!strings.HasSuffix(domain, 'utah.edu')",
			"Email does not end with utah.edu")

		return
	}

	fileName := "./data/" + uid + ".json"
	mutex := getMutex(fileName)
	mutex.Lock()
	defer mutex.Unlock()
	dataBytes, err := os.ReadFile(fileName)

	if err == nil {
		handleError(responseWriter, "4", function, "ReadFile(fileName)", "User already exists")
		return
	}

	dataJson := DataJson{
		Email:     email,
		Password:  password,
		Name:      name,
		Uid:       uid,
		Count:     0,
		Timestamp: time.Now().UnixMicro(),
	}

	dataBytes, err = json.Marshal(dataJson)

	if err != nil {
		handleError(responseWriter, "5", function, "Marshal(dataJson)", "")
		return
	}

	err = os.WriteFile(fileName, dataBytes, 0600)

	if err != nil {
		handleError(responseWriter, "5", function, "WriteFile(fileName, dataBytes, 0600)", "")
		return
	}

	userBytes, err := json.Marshal(UserJson{
		Error: false,
		Email: dataJson.Email,
		Name:  dataJson.Name,
		Uid:   dataJson.Uid,
		Count: dataJson.Count,
	})

	if err != nil {
		handleError(responseWriter, "5", function, "Marshal(UserJson)", "")
		return
	}

	responseWriter.Header().Add("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Add("Content-Type", "application/json")
	_, err = responseWriter.Write(userBytes)

	if err != nil {
		handleError(responseWriter, "5", function, "Write(userBytes)", "")
		return
	}

	handleError(responseWriter, "2", function, "", "")
}

func handleSignIn(responseWriter http.ResponseWriter, request *http.Request) {
	function := "handleSignIn"

	var signInJson SignInJson
	err := json.NewDecoder(request.Body).Decode(&signInJson)

	if err != nil {
		handleError(responseWriter, "4", function, "Decode(&signInJson)", "Request body is invalid")
		return
	}

	email := signInJson.Email
	password := signInJson.Password

	if email == "" || password == "" {
		handleError(responseWriter, "4", function, "email == '' || password == ''",
			"Email or password is empty")

		return
	}

	uid, domain, found := strings.Cut(email, "@")

	if !found || uid == "" {
		handleError(responseWriter, "4", function, "!found || uid == ''", "Email is invalid")
		return
	}

	if !strings.HasSuffix(domain, "utah.edu") {
		handleError(responseWriter, "4", function, "!strings.HasSuffix(domain, 'utah.edu')",
			"Email does not end with utah.edu")

		return
	}

	fileName := "./data/" + uid + ".json"
	mutex := getMutex(fileName)
	mutex.Lock()
	defer mutex.Unlock()
	dataBytes, err := os.ReadFile(fileName)

	if err != nil {
		handleError(responseWriter, "4", function, "ReadFile(fileName)", "User does not exist")
		return
	}

	var dataJson DataJson
	err = json.Unmarshal(dataBytes, &dataJson)

	if err != nil {
		handleError(responseWriter, "5", function, "Unmarshal(dataBytes, &dataJson)", "")
		return
	}

	// if email != dataJson.Email {
	// 	handleError(responseWriter, "4", function, "email != dataJson.Email",
	// 		"Email is incorrect")
	//
	// 	return
	// }

	if password != dataJson.Password {
		handleError(responseWriter, "4", function, "password != dataJson.Password",
			"Password is incorrect")

		return
	}

	userBytes, err := json.Marshal(UserJson{
		Error: false,
		Email: dataJson.Email,
		Name:  dataJson.Name,
		Uid:   dataJson.Uid,
		Count: dataJson.Count,
	})

	if err != nil {
		handleError(responseWriter, "5", function, "Marshal(UserJson)", "")
		return
	}

	responseWriter.Header().Add("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Add("Content-Type", "application/json")
	_, err = responseWriter.Write(userBytes)

	if err != nil {
		handleError(responseWriter, "5", function, "Write(userBytes)", "")
		return
	}

	handleError(responseWriter, "2", function, "", "")
}

func handleUser(responseWriter http.ResponseWriter, request *http.Request) {
	function := "handleUser"

	var uidJson UidJson
	err := json.NewDecoder(request.Body).Decode(&uidJson)

	if err != nil {
		handleError(responseWriter, "4", function, "Decode(&uidJson)", "Request body is invalid")
		return
	}

	uid := uidJson.Uid

	if uid == "" {
		handleError(responseWriter, "4", function, "uid == ''", "UID is empty")
		return
	}

	fileName := "./data/" + uid + ".json"
	mutex := getMutex(fileName)
	mutex.Lock()
	defer mutex.Unlock()
	dataBytes, err := os.ReadFile(fileName)

	if err != nil {
		handleError(responseWriter, "4", function, "ReadFile(fileName)", "User does not exist")
		return
	}

	var dataJson DataJson
	err = json.Unmarshal(dataBytes, &dataJson)

	if err != nil {
		handleError(responseWriter, "5", function, "Unmarshal(dataBytes, &dataJson)", "")
		return
	}

	userBytes, err := json.Marshal(UserJson{
		Error: false,
		Email: dataJson.Email,
		Name:  dataJson.Name,
		Uid:   dataJson.Uid,
		Count: dataJson.Count,
	})

	if err != nil {
		handleError(responseWriter, "5", function, "Marshal(UserJson)", "")
		return
	}

	responseWriter.Header().Add("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Add("Content-Type", "application/json")
	_, err = responseWriter.Write(userBytes)

	if err != nil {
		handleError(responseWriter, "5", function, "Write(userBytes)", "")
		return
	}

	handleError(responseWriter, "2", function, "", "")
}

func handleRename(responseWriter http.ResponseWriter, request *http.Request) {
	function := "handleRename"

	var renameJson RenameJson
	err := json.NewDecoder(request.Body).Decode(&renameJson)

	if err != nil {
		handleError(responseWriter, "4", function, "Decode(&renameJson)", "Request body is invalid")
		return
	}

	uid := renameJson.Uid
	name := renameJson.Name

	if uid == "" || name == "" {
		handleError(responseWriter, "4", function, "uid == '' || name == ''",
			"UID or name is empty")

		return
	}

	fileName := "./data/" + uid + ".json"
	mutex := getMutex(fileName)
	mutex.Lock()
	defer mutex.Unlock()
	dataBytes, err := os.ReadFile(fileName)

	if err != nil {
		handleError(responseWriter, "4", function, "ReadFile(fileName)", "User does not exist")
		return
	}

	var dataJson DataJson
	err = json.Unmarshal(dataBytes, &dataJson)

	if err != nil {
		handleError(responseWriter, "5", function, "Unmarshal(dataBytes, &dataJson)", "")
		return
	}

	dataJson.Name = name
	dataBytes, err = json.Marshal(dataJson)

	if err != nil {
		handleError(responseWriter, "5", function, "Marshal(dataJson)", "")
		return
	}

	err = os.WriteFile(fileName, dataBytes, 0600)

	if err != nil {
		handleError(responseWriter, "5", function, "WriteFile(fileName, dataBytes, 0600)", "")
		return
	}

	userBytes, err := json.Marshal(UserJson{
		Error: false,
		Email: dataJson.Email,
		Name:  dataJson.Name,
		Uid:   dataJson.Uid,
		Count: dataJson.Count,
	})

	if err != nil {
		handleError(responseWriter, "5", function, "Marshal(UserJson)", "")
		return
	}

	responseWriter.Header().Add("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Add("Content-Type", "application/json")
	_, err = responseWriter.Write(userBytes)

	if err != nil {
		handleError(responseWriter, "5", function, "Write(renameBytes)", "")
		return
	}

	handleError(responseWriter, "2", function, "", "")
}

func handleClick(responseWriter http.ResponseWriter, request *http.Request) {
	function := "handleClick"

	var uidJson UidJson
	err := json.NewDecoder(request.Body).Decode(&uidJson)

	if err != nil {
		handleError(responseWriter, "4", function, "Decode(&uidJson)", "Request body is invalid")
		return
	}

	uid := uidJson.Uid

	if uid == "" {
		handleError(responseWriter, "4", function, "uid == ''", "UID is empty")
		return
	}

	fileName := "./data/" + uid + ".json"
	mutex := getMutex(fileName)
	mutex.Lock()
	defer mutex.Unlock()
	dataBytes, err := os.ReadFile(fileName)

	if err != nil {
		handleError(responseWriter, "4", function, "ReadFile(fileName)", "User does not exist")
		return
	}

	var dataJson DataJson
	err = json.Unmarshal(dataBytes, &dataJson)

	if err != nil {
		handleError(responseWriter, "5", function, "Unmarshal(dataBytes, &dataJson)", "")
		return
	}

	dataJson.Count++
	dataJson.Timestamp = time.Now().UnixMicro()
	dataBytes, err = json.Marshal(dataJson)

	if err != nil {
		handleError(responseWriter, "5", function, "Marshal(dataJson)", "")
		return
	}

	err = os.WriteFile(fileName, dataBytes, 0600)

	if err != nil {
		handleError(responseWriter, "5", function, "WriteFile(fileName, dataBytes, 0600)", "")
		return
	}

	userBytes, err := json.Marshal(UserJson{
		Error: false,
		Email: dataJson.Email,
		Name:  dataJson.Name,
		Uid:   dataJson.Uid,
		Count: dataJson.Count,
	})

	if err != nil {
		handleError(responseWriter, "5", function, "Marshal(UserJson)", "")
		return
	}

	responseWriter.Header().Add("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Add("Content-Type", "application/json")
	_, err = responseWriter.Write(userBytes)

	if err != nil {
		handleError(responseWriter, "5", function, "Write(userBytes)", "")
		return
	}

	handleError(responseWriter, "2", function, "", "")
}

func main() {
	http.HandleFunc("/time", handleTime)
	http.HandleFunc("/signup", handleSignUp)
	http.HandleFunc("/signin", handleSignIn)
	http.HandleFunc("/user", handleUser)
	http.HandleFunc("/rename", handleRename)
	http.HandleFunc("/click", handleClick)

	log.Println("Listening on port 80")
	log.Fatalln(http.ListenAndServe(":80", nil))
}
