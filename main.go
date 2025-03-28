package main

import (
	"cmp"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

var mutexes = make(map[string]*sync.Mutex)
var leaderboard []LeaderboardItemJson

type ErrorJson struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type LeaderboardItemJson struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Uid       string `json:"uid"`
	Count     int    `json:"count"`
	Timestamp int64  `json:"timestamp"`
}

type DataJson struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	Uid       string `json:"uid"`
	Count     int    `json:"count"`
	Timestamp int64  `json:"timestamp"`
	Token     int64  `json:"token"`
}

type TimeJson struct {
	Time    string `json:"time"`
	Version string `json:"version"`
}

type SignUpJson struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type UserJson struct {
	Error bool   `json:"error"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Uid   string `json:"uid"`
	Count int    `json:"count"`
}

type SignInJson struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenJson struct {
	Error bool   `json:"error"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Uid   string `json:"uid"`
	Count int    `json:"count"`
	Token int64  `json:"token"`
}

type UidJson struct {
	Uid   string `json:"uid"`
	Token int64  `json:"token"`
}

type RenameJson struct {
	Uid   string `json:"uid"`
	Name  string `json:"name"`
	Token int64  `json:"token"`
}

type LeaderboardJson struct {
	Error       bool                  `json:"error"`
	Leaderboard []LeaderboardItemJson `json:"leaderboard"`
}

func logError(level string, function string, code string, message string) {
	log.Printf("%-4s %-24s %-48s %s\n", level, function, code, message)
}

func handleError(responseWriter http.ResponseWriter, level string, function string, code string,
	message string) {
	logError(level, function, code, message)

	function = "handleError"

	if level != "4" {
		return
	}

	errorBytes, err := json.Marshal(ErrorJson{
		Error:   true,
		Message: message,
	})

	if err != nil {
		logError("5", function, "Marshal(ErrorJson)", "")
		return
	}

	responseWriter.Header().Add("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Add("Content-Type", "application/json")
	_, err = responseWriter.Write(errorBytes)

	if err != nil {
		logError("5", function, "Write(errorBytes)", "")
		return
	}

	logError("2", function, "", "")
}

func getMutex(fileName string) *sync.Mutex {
	mutex, exists := mutexes[fileName]

	if exists {
		return mutex
	}

	mutex = &sync.Mutex{}
	mutexes[fileName] = mutex
	return mutex
}

func updateLeaderboard(leaderboardItemJson LeaderboardItemJson) {
	mutex := getMutex("leaderboard")
	mutex.Lock()
	defer mutex.Unlock()

	leaderboard = slices.DeleteFunc(leaderboard, func(item LeaderboardItemJson) bool {
		return strings.Compare(item.Uid, leaderboardItemJson.Uid) == 0
	})

	leaderboard = append(leaderboard, leaderboardItemJson)

	slices.SortFunc(leaderboard, func(a, b LeaderboardItemJson) int {
		if difference := cmp.Compare(b.Count, a.Count); difference != 0 {
			return difference
		}

		return cmp.Compare(a.Timestamp, b.Timestamp)
	})

	if len(leaderboard) == 11 {
		leaderboard = leaderboard[:10]
	}
}

func initializeLeaderboard() {
	function := "initializeLeaderboard"

	entries, err := os.ReadDir("./data")

	if err != nil {
		logError("5", function, "ReadDir('./data')", err.Error())
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := "./data/" + entry.Name()
		mutex := getMutex(fileName)
		mutex.Lock()
		defer mutex.Unlock()
		dataBytes, err := os.ReadFile(fileName)

		if err != nil {
			logError("5", function, "ReadFile(fileName)", err.Error())
			os.Exit(1)
		}

		var dataJson DataJson
		err = json.Unmarshal(dataBytes, &dataJson)

		if err != nil {
			logError("5", function, "Unmarshal(dataBytes, &dataJson)", err.Error())
			os.Exit(1)
		}

		updateLeaderboard(LeaderboardItemJson{
			Email:     dataJson.Email,
			Name:      dataJson.Name,
			Uid:       dataJson.Uid,
			Count:     dataJson.Count,
			Timestamp: dataJson.Timestamp,
		})
	}

	logError("2", function, "", "")
}

func handleTime(responseWriter http.ResponseWriter, request *http.Request) {
	function := "handleTime"

	timeBytes, err := json.Marshal(TimeJson{
		Time:    time.Now().UTC().Add(time.Hour * 9).Format(time.DateTime),
		Version: "2025.3.5.0",
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
		Token:     0,
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

	updateLeaderboard(LeaderboardItemJson{
		Email:     dataJson.Email,
		Name:      dataJson.Name,
		Uid:       dataJson.Uid,
		Count:     dataJson.Count,
		Timestamp: dataJson.Timestamp,
	})

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
	// 	handleError(responseWriter, "4", function, "email != dataJson.Email", "Email is incorrect")
	//
	// 	return
	// }

	if password != dataJson.Password {
		handleError(responseWriter, "4", function, "password != dataJson.Password",
			"Password is incorrect")

		return
	}

	dataJson.Token = time.Now().UnixMicro()
	dataBytes, err = json.Marshal(dataJson)

	if err != nil {
		handleError(responseWriter, "5", function, "Marshal(dataJson)", "")
		return
	}

	err = os.WriteFile(fileName, dataBytes, 0600)

	userBytes, err := json.Marshal(TokenJson{
		Error: false,
		Email: dataJson.Email,
		Name:  dataJson.Name,
		Uid:   dataJson.Uid,
		Count: dataJson.Count,
		Token: dataJson.Token,
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
	token := uidJson.Token

	if uid == "" || token == 0 {
		handleError(responseWriter, "4", function, "uid == '' || token == 0",
			"UID or token is empty")

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

	if token != dataJson.Token {
		handleError(responseWriter, "4", function, "token != dataJson.Token", "Token is invalid")
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
	token := renameJson.Token

	if uid == "" || name == "" || token == 0 {
		handleError(responseWriter, "4", function, "uid == '' || name == '' || token == 0",
			"UID, name, or token is empty")

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

	if token != dataJson.Token {
		handleError(responseWriter, "4", function, "token != dataJson.Token", "Token is invalid")
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

	updateLeaderboard(LeaderboardItemJson{
		Email:     dataJson.Email,
		Name:      dataJson.Name,
		Uid:       dataJson.Uid,
		Count:     dataJson.Count,
		Timestamp: dataJson.Timestamp,
	})

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
	token := uidJson.Token

	if uid == "" || token == 0 {
		handleError(responseWriter, "4", function, "uid == '' || token == 0",
			"UID or token is empty")

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

	if token != dataJson.Token {
		handleError(responseWriter, "4", function, "token != dataJson.Token", "Token is invalid")
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

	updateLeaderboard(LeaderboardItemJson{
		Email:     dataJson.Email,
		Name:      dataJson.Name,
		Uid:       dataJson.Uid,
		Count:     dataJson.Count,
		Timestamp: dataJson.Timestamp,
	})

	responseWriter.Header().Add("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Add("Content-Type", "application/json")
	_, err = responseWriter.Write(userBytes)

	if err != nil {
		handleError(responseWriter, "5", function, "Write(userBytes)", "")
		return
	}

	handleError(responseWriter, "2", function, "", "")
}

func handleLeaderboard(responseWriter http.ResponseWriter, request *http.Request) {
	function := "handleLeaderboard"

	var uidJson UidJson
	err := json.NewDecoder(request.Body).Decode(&uidJson)

	if err != nil {
		handleError(responseWriter, "4", function, "Decode(&uidJson)", "Request body is invalid")
		return
	}

	uid := uidJson.Uid
	token := uidJson.Token

	if uid == "" || token == 0 {
		handleError(responseWriter, "4", function, "uid == '' || token == 0",
			"UID or token is empty")

		return
	}

	fileName := "./data/" + uid + ".json"
	mutex := getMutex(fileName)
	mutex.Lock()
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

	if token != dataJson.Token {
		handleError(responseWriter, "4", function, "token != dataJson.Token", "Token is invalid")
		return
	}

	leaderboardItemJson := LeaderboardItemJson{
		Email:     dataJson.Email,
		Name:      dataJson.Name,
		Uid:       dataJson.Uid,
		Count:     dataJson.Count,
		Timestamp: dataJson.Timestamp,
	}

	mutex.Unlock()
	mutex = getMutex("leaderboard")
	mutex.Lock()
	defer mutex.Unlock()

	var tempLeaderboard []LeaderboardItemJson
	contains := false

	for _, item := range leaderboard {
		tempLeaderboard = append(tempLeaderboard, item)

		if item.Uid == leaderboardItemJson.Uid {
			contains = true
		}
	}

	if !contains {
		if len(tempLeaderboard) == 10 {
			tempLeaderboard = tempLeaderboard[:9]
		}

		tempLeaderboard = append(tempLeaderboard, leaderboardItemJson)
	}

	leaderboardBytes, err := json.Marshal(LeaderboardJson{
		Error:       false,
		Leaderboard: tempLeaderboard,
	})

	if err != nil {
		handleError(responseWriter, "5", function, "Marshal(LeaderboardJson)", "")
		return
	}

	responseWriter.Header().Add("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Add("Content-Type", "application/json")
	_, err = responseWriter.Write(leaderboardBytes)

	if err != nil {
		handleError(responseWriter, "5", function, "Write(leaderboardBytes)", "")
		return
	}

	handleError(responseWriter, "2", function, "", "")
}

func main() {
	initializeLeaderboard()

	http.HandleFunc("/time", handleTime)
	http.HandleFunc("/signup", handleSignUp)
	http.HandleFunc("/signin", handleSignIn)
	http.HandleFunc("/user", handleUser)
	http.HandleFunc("/rename", handleRename)
	http.HandleFunc("/click", handleClick)
	http.HandleFunc("/leaderboard", handleLeaderboard)

	logError("2", "main", "", "")
	log.Fatalln(http.ListenAndServe(":80", nil))
}
