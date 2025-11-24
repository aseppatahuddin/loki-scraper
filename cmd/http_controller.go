package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func httpstartQueryHandler(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type header
	w.Header().Set("Content-Type", "text/plain")

	// Set the HTTP status code
	w.WriteHeader(http.StatusOK)

	// Write the response body
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:]) // r.URL.Path[1:] gets the path after the leading '/'
}

// func to start HTTP controller
func httpServer() error {
	var httpControllerPort = os.Getenv("PORT")
	if httpControllerPort == "" {
		httpControllerPort = "9000"
	}

	http.HandleFunc("/", httpstartQueryHandler)

	// Start the HTTP server on port 8080
	log.Println("Starting server on :" + httpControllerPort)
	err := http.ListenAndServe(":"+httpControllerPort, nil)
	if err != nil {
		return fmt.Errorf("listen and serve: %s", err.Error())
	}

	return nil
}
