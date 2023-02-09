

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AddReminder struct {
	UserID      int `json:"userid"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phonenumber"`
	Message     string `json:"message"`
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	var body AddReminder
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Do something with the request body
	fmt.Println("Name:", body.UserID)
	fmt.Println("Name:", body.Name)
	fmt.Println("Age:", body.PhoneNumber)
	fmt.Println("Name:", body.Message)

	// Write the response
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "Data received")
}

// Create a new instance of the RequestBody struct
reminder:= AddReminder{
	UserID: "1"
	Name: "Tom"
	PhoneNumber: "+44 123456789"
	Message: "First message"
}
