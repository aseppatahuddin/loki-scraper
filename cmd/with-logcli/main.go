package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"dario.cat/mergo"
	"github.com/prakasa1904/loki-scraper/internal/connection"
	"github.com/prakasa1904/loki-scraper/internal/constants"
	"github.com/prakasa1904/loki-scraper/internal/lokiparser"
	"github.com/prakasa1904/loki-scraper/internal/model"
)

func main() {
	debug := os.Getenv("LOKI_DEBUG")
	if debug == "" {
		log.Println("LOKI_DEBUG is not exist in environment variable!")
		return
	}

	bin := os.Getenv("LOKI_BINARY")
	if bin == "" {
		log.Println("LOKI_BINARY is not exist in environment variable!")
		return
	}

	query := os.Getenv("LOKI_QUERY")
	if query == "" {
		log.Println("LOKI_QUERY is not exist in environment variable!")
		return
	}
	startDate := os.Getenv("LOKI_START_DATE")
	if startDate == "" {
		log.Println("LOKI_START_DATE is not exist in environment variable!")
		return
	}
	endDate := os.Getenv("LOKI_END_DATE")
	if endDate == "" {
		log.Println("LOKI_END_DATE is not exist in environment variable!")
		return
	}
	limit := os.Getenv("LOKI_LIMIT")
	if limit == "" {
		log.Println("LOKI_LIMIT is not exist in environment variable!")
		return
	}

	// clickhouse table target
	dbTable := os.Getenv("CLICKHOUSE_TABLE")
	if dbTable == "" {
		log.Println("CLICKHOUSE_TABLE is not exist in environment variable!")
		return
	}

	cmd := exec.Command(bin, "query", "--limit", "0", "--batch", limit, "--output=jsonl", "--from", startDate, "--to", endDate, query)

	// Create a pipe to read the standard output of the command
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Error creating StdoutPipe: %v", err)
	}
	defer stdout.Close()

	// Start the command
	if err := cmd.Start(); err != nil {
		log.Fatalf("Error starting command: %v", err)
	}

	// db connection
	conn, err := connection.Connect()
	if err != nil {
		log.Panicln("Error DB connection: ", err.Error())
	}

	// Wait group to ensure the main function waits for the goroutine to finish.
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		// Use a scanner to read the output line by line (each line is a JSON object)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			// First, unmarshal the outer JSON into a temporary struct or map
			// to extract the escaped inner JSON string.
			var temp struct {
				Line      string `json:"line"`
				Timestamp string `json:"timestamp"`
			}
			if err := json.Unmarshal([]byte(line), &temp); err != nil {
				log.Fatalf("Error unmarshaling outer JSON: %v", err)
			}

			var entry model.LogEntry
			if err := json.Unmarshal([]byte(temp.Line), &entry); err != nil {
				log.Println("Error parsing JSON line: ", err.Error())
				continue
			}

			// Prepare the batch insert statement
			batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO "+dbTable+" (timestamp, job, namespace, node_name, pod, app, container, filename, stream, time, apiName, proxyResponseCode, destination, apiCreatorTenantDomain, platform, apiMethod, apiVersion, gatewayType, apiCreator, responseCacheHit, backendLatency, correlationId, requestMediationLatency, keyType, apiId, applicationName, targetResponseCode, requestTimestamp, applicationOwner, userAgent, eventType, apiResourceTemplate, regionId, responseLatency, responseMediationLatency, userIp, apiContext, applicationId, apiType)")
			if err != nil {
				log.Fatal("Error on prepare batch: ", err.Error())
			}

			if lokiparser.IfLogContains(entry.Log, constants.FILTER_LOG) {
				whitespaceClean := strings.Join(strings.Fields(entry.Log), " ")
				rawArrayString := strings.Split(whitespaceClean, "Value:")
				var metricsValue string
				if len(rawArrayString) == 2 {
					metricsValue = strings.TrimSpace(rawArrayString[1])

					metric, err := lokiparser.Parse(metricsValue)
					if err != nil {
						log.Panicln(err.Error())
					}

					// merge data metrics to clickhouse
					mergo.Merge(&entry, metric, mergo.WithOverride, mergo.WithoutDereference)

					t, err := time.Parse(time.RFC3339, metric.RequestTimestamp)
					if err != nil {
						log.Panicln(err.Error())
					}

					entry.Time = t.UTC() // format to UTC

					err = batch.AppendStruct(&entry) // Use AppendStruct for struct-based insertion
					if err != nil {
						log.Println("Failed to append structued data: ", err.Error())
					}

					// Send the batch
					if err := batch.Send(); err != nil {
						log.Fatal(err)
					}
				}
			}

			if debug == "true" {
				log.Println("entry.Job: ", entry.Job)
				log.Println("entry.Time: ", entry.Time)
				log.Println("entry.APIType: ", entry.APIType)
				log.Println("entry.UserAgent: ", entry.UserAgent)
				log.Println("entry.Namespace: ", entry.Namespace)
				log.Println("entry.APIContext: ", entry.APIContext)
			}
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			log.Fatalf("Error reading stdout: %v", err)
		}
	}()

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		log.Fatalf("Command finished with error: %v", err)
	}
	wg.Wait() // Wait for the scanner goroutine to finish processing all output

	fmt.Println("Command finished successfully.")
}
