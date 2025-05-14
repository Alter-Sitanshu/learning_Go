package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

var r_mux sync.RWMutex

type User struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Sex     int    `json:"sex"`
}

var UserCache = make(map[int]User)

func main() {
	server := http.NewServeMux()
	server.HandleFunc("/", RootHandler)
	server.HandleFunc("POST /user", CreateNewUser)
	server.HandleFunc("GET /user/{id}", GetUser)
	server.HandleFunc("DELETE /user/{id}", DeleteUser)

	fmt.Println("Server started and listening to port 8000.")
	http.ListenAndServe(":8000", server)

}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to our Server\n")
}

func CreateNewUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "Missing Name(Required)", http.StatusBadRequest)
		return
	}

	r_mux.Lock()
	UserCache[len(UserCache)+1] = user
	r_mux.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		fmt.Println("Failed to write response:", err)
		return
	}

}

func GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r_mux.RLock()
	user, ok := UserCache[id]
	r_mux.RUnlock()

	if !ok {
		http.Error(w, "User Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		fmt.Println("Failed to write response:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r_mux.RLock()
	_, ok := UserCache[id]
	r_mux.RUnlock()

	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	delete(UserCache, id)
	w.WriteHeader(http.StatusNoContent)
}
