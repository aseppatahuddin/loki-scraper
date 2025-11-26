package main

import (
	"context"
	"log"
	"os"
	"time"
)

var startDate string
var endDate string

func dateParse(date string) (time.Time, error) {
	// Define the layout string based on the reference date
	// "2006-01-02 15:04:05" represents a full date and time
	layout := "2006-01-02T15:04:05Z"

	// Parse the string into a time.Time object
	t, err := time.Parse(layout, date)
	if err != nil {
		return t, err
	}

	return t, nil
}

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

	// start, err := dateParse(startDate)
	// if err != nil {
	// 	log.Printf("Error running loki scraper: %s", err.Error())
	// 	return
	// }

	// end, err := dateParse(endDate)
	// if err != nil {
	// 	log.Panicf("Error running loki scraper: %s", err.Error())
	// }

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
