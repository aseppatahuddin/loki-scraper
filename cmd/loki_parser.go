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

// Define structs to match Loki API response
type LokiResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body %s", err.Error())
	}

	var lokiResp LokiResponse
	err = json.Unmarshal(body, &lokiResp)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON: %s", err.Error())
	}

	if lokiResp.Status == "success" {
		for _, stream := range lokiResp.Data.Result {
			log.Printf("Stream Labels: %v\n", stream.Stream)
			for _, entry := range stream.Values {
				timestamp, logLine := entry[0], entry[1]
				log.Printf("  Timestamp: %s, Log Line: %s\n", timestamp, logLine)
			}
		}
	} else {
		return fmt.Errorf("loki API error: %s", lokiResp.Status)
	}

	return nil
}
