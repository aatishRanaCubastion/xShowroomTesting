package models

import (
	"shared/router"
	"net/http"
	"encoding/json"
)

type Response struct {
	Data          interface{}    `json:"data"`
	StatusCode    uint            `json:"status_code"`
	StatusMessage string        `json:"status_message"`
}

// Load forces the program to call all the init() funcs in each models file
func Load() {
	router.Get("/", Welcome)
}

func Welcome(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode("Welcome to xShowroom")
}
