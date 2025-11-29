package main

import (
	"context"
	"log"
	"os"
)

var startDate string
var endDate string

func main() {
	// set extra query to fetch loki data
	startDate = os.Getenv("LOKI_START_DATE")
	if startDate == "" {
		log.Println("LOKI_START_DATE is not exist in environment variable!")
		return
	}
	endDate = os.Getenv("LOKI_END_DATE")
	if endDate == "" {
		log.Println("LOKI_START_DATE is not exist in environment variable!")
		return
	}

	// get loki query, format times into RFC3339Nano and limit
	var lokiQuery = os.Getenv("LOKI_QUERY")
	var lokiStartDate = startDate
	var lokiEndDate = endDate

	log.Println("Starting loki scrapper")
	// go func() {
	err := lokiParser(context.Background(), lokiQuery, lokiStartDate, lokiEndDate)
	if err != nil {
		log.Printf("Error running loki parser: %v\n", err)
	} else {
		log.Println("Scraper command executed successfully.")
	}
}
