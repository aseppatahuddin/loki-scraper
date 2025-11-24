package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var currentState string
var startDate string
var endDate string

func dateParse(date string) (time.Time, error) {
	// Define the layout string based on the reference date
	// "2006-01-02 15:04:05" represents a full date and time
	layout := "2006-01-02 15:04:05"

	// Parse the string into a time.Time object
	t, err := time.Parse(layout, date)
	if err != nil {
		return t, err
	}

	return t, nil
}

func main() {
	// Use a context with a timeout of 10 seconds:
	appContext, appCancelContext := context.WithTimeout(context.Background(), 10*time.Second)
	// Use simple context based on OS:
	// appContext := context.Background()

	// add HTTP controller if need it, can be used to:
	// - Start
	// - Stop
	// httpServer()

	// set extra query to fetch loki data
	startDate = os.Getenv("LOKI_START_DATE")
	endDate = os.Getenv("LOKI_END_DATE")

	start, err := dateParse(startDate)
	if err != nil {
		log.Panicf("Error running loki scraper: %s", err.Error())
	}

	end, err := dateParse(endDate)
	if err != nil {
		log.Panicf("Error running loki scraper: %s", err.Error())
	}

	// get loki query, format times into RFC3339Nano and limit
	var lokiQuery = os.Getenv("LOKI_QUERY")
	var lokiStartDate = start.Format(time.RFC3339Nano)
	var lokiEndDate = end.Format(time.RFC3339Nano)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Starting loki scrapper")
	go func() {
		err := lokiParser(appContext, lokiQuery, lokiStartDate, lokiEndDate)
		if err != nil {
			log.Printf("Error running loki parser: %v\n", err)
		} else {
			log.Println("Scraper command executed successfully.")
		}
	}()

	select {
	case <-done:
		fmt.Println("Received interrupt signal. Stopping command...")
		appCancelContext()
	case err := <-done:
		if err != nil {
			log.Printf("Error running loki scraper: %v\n", err)
		} else {
			log.Println("loki scraper command executed successfully.")
		}
	}
}
