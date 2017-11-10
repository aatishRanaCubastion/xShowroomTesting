package models

import (
	"shared/router"
	"net/http"
	"encoding/json"
)

// Load forces the program to call all the init() funcs in each models file
func Load() {
	router.Get("/", WelCome)
}

func WelCome(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode("Welcome to xShowroom")
}
