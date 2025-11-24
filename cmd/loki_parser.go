package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type LogEntry struct {
	Timestamp string    `json:"timestamp" ch:"timestamp"`
	Log       string    `json:"log" ch:"log"`
	Stream    string    `json:"stream" ch:"stream"`
	Time      time.Time `json:"time" ch:"time"`
}

// Define structs to match Loki API response
type LokiResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream         map[string]string `json:"stream"`
			RawValues      [][]string        `json:"values"`
			StructedValues []LogEntry        `json:"log"`
		} `json:"result"`
	} `json:"data"`
}

func setLimit() string {
	var limit = os.Getenv("LOKI_LIMIT")
	if limit == "" {
		return "100"
	}

	return limit
}

func lokiParser(ctx context.Context, query string, startDate string, endDate string) error {
	var LOKI_URL = os.Getenv("LOKI_URL")
	if LOKI_URL == "" {
		return fmt.Errorf("no Loki URL set from env LOKI_URL")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", LOKI_URL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %s", err.Error())
	}

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("start", startDate)
	q.Add("end", endDate)
	q.Add("limit", setLimit())
	req.URL.RawQuery = q.Encode()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("loki API returned non-OK status: %s", resp.Status)
	}

	log.Println("Start to receiving loki data from URL ", req.URL.String())

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body %s", err.Error())
	}

	var lokiResp LokiResponse
	err = json.Unmarshal(body, &lokiResp)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON: %s", err.Error())
	}

	// on success, formating data
	if lokiResp.Status == "success" {
		for i, stream := range lokiResp.Data.Result {
			// log.Printf("Stream Labels: %v\n", stream.Stream)
			for _, entry := range stream.RawValues {
				timestamp, logLine := entry[0], entry[1]

				var structedJson LogEntry
				err := json.Unmarshal([]byte(logLine), &structedJson)
				if err != nil {
					return fmt.Errorf("parsing loki log error: %s", err.Error())
				}

				// add log ID from timestamp
				structedJson.Timestamp = timestamp

				// push to structed json
				lokiResp.Data.Result[i].StructedValues = append(lokiResp.Data.Result[i].StructedValues, structedJson)
			}
		}
	} else {
		return fmt.Errorf("loki API error: %s", lokiResp.Status)
	}

	conn, err := connect()
	if err != nil {
		log.Panicln("Error DB connection: ", err.Error())
	}

	// Prepare the batch insert statement
	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO log_entry (timestamp, log, stream, time)")
	if err != nil {
		log.Fatal("Error on prepare batch: ", err.Error())
	}
	defer batch.Close()

	for _, results := range lokiResp.Data.Result {
		for _, structedData := range results.StructedValues {
			err := batch.AppendStruct(&structedData) // Use AppendStruct for struct-based insertion
			if err != nil {
				log.Fatal("Failed to append structued data: ", err.Error())
			}
		}
	}

	// Send the batch
	if err := batch.Send(); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully inserted data.")

	return nil
}
