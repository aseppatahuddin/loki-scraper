package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/prakasa1904/loki-scraper/internal/connection"
)

func main() {
	var LOKI_URL = os.Getenv("LOKI_URL")
	if LOKI_URL == "" {
		log.Println("LOKI_URL is not exist in environment variable!")
		return
	}

	var LOKI_QUERY = os.Getenv("LOKI_QUERY")
	if LOKI_QUERY == "" {
		log.Println("LOKI_QUERY is not exist in environment variable!")
		return
	}

	var LOKI_LIMIT = os.Getenv("LOKI_LIMIT")
	if LOKI_LIMIT == "" {
		log.Println("LOKI_LIMIT is not exist in environment variable!")
		return
	}

	intLimit, err := strconv.Atoi(LOKI_LIMIT)
	if err != nil {
		log.Println("Error to parse LOKI_LIMIT with error: ", err.Error())
		return
	}

	var LOKI_LAST_MONTH = os.Getenv("LOKI_LAST_MONTH")
	if LOKI_LAST_MONTH == "" {
		log.Println("LOKI_LAST_MONTH is not exist in environment variable!")
		return
	}

	intLastMonthAgo, err := strconv.Atoi(LOKI_LAST_MONTH)
	if err != nil {
		log.Println("Error to parse LOKI_LAST_MONTH with error: ", err.Error())
		return
	}

	if intLastMonthAgo > 0 {
		log.Println("LOKI_LAST_MONTH must negative value!")
		return
	}

	dbConnection, err := connection.Connect()
	if err != nil {
		log.Println("Error connect to DB with error: ", err.Error())
		return
	}

	// Range: last dynamic month
	end := time.Now()
	start := end.AddDate(0, intLastMonthAgo, 0)

	// Split into days
	days := splitByDays(start, end)

	// Worker pool (5 workers)
	workerCount := 5
	dayCh := make(chan [2]time.Time)
	wg := sync.WaitGroup{}

	// Start workers
	for i := 0; i < workerCount; i++ {
		go func() {
			for day := range dayCh {
				processDay(dbConnection, LOKI_URL, LOKI_QUERY, intLimit, day[0], day[1])
				wg.Done()
			}
		}()
	}

	// Push jobs
	for _, d := range days {
		wg.Add(1)
		dayCh <- d
	}

	close(dayCh)
	wg.Wait()

	fmt.Println("All days finished.")
}
