package models

import (
	"shared/router"
	"net/http"
	"encoding/json"
)

type Response struct {
	StatusCode    uint            `json:"status_code"`
	StatusMessage string        `json:"status_message"`
	Data          interface{}    `json:"data"`
}

// Load forces the program to call all the init() funcs in each models file
func Load() {
	router.Get("/", Welcome)
}

func Welcome(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{2000, "Welcome to xShowroom", nil})
}
