package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"dario.cat/mergo"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/prakasa1904/loki-scraper/internal/constants"
	"github.com/prakasa1904/loki-scraper/internal/lokiclient"
	"github.com/prakasa1904/loki-scraper/internal/lokiparser"
	"github.com/prakasa1904/loki-scraper/internal/model"
)

func splitByDays(start, end time.Time) [][2]time.Time {
	var days [][2]time.Time
	cur := start

	for cur.Before(end) {
		next := cur.Add(24 * time.Hour)
		if next.After(end) {
			next = end
		}
		days = append(days, [2]time.Time{cur, next})
		cur = next
	}
	return days
}

func processDay(dbConnection driver.Conn, lokiURL string, query string, limit int, dayStart, dayEnd time.Time) {
	fmt.Printf("Start day: %s\n", dayStart.Format("2006-01-02"))

	var CLICKHOUSE_TABLE = os.Getenv("CLICKHOUSE_TABLE")
	if CLICKHOUSE_TABLE == "" {
		log.Println("no clickhouse table set from env variable.")
		return
	}

	current := dayStart

	// Prepare the batch insert statement
	batch, err := dbConnection.PrepareBatch(context.Background(), "INSERT INTO "+CLICKHOUSE_TABLE+" (timestamp, job, namespace, node_name, pod, app, container, filename, stream, time, apiName, proxyResponseCode, destination, apiCreatorTenantDomain, platform, apiMethod, apiVersion, gatewayType, apiCreator, responseCacheHit, backendLatency, correlationId, requestMediationLatency, keyType, apiId, applicationName, targetResponseCode, requestTimestamp, applicationOwner, userAgent, eventType, apiResourceTemplate, regionId, responseLatency, responseMediationLatency, userIp, apiContext, applicationId, apiType)")
	if err != nil {
		log.Fatal("Error on prepare batch: ", err.Error())
	}

	for {
		resp := lokiclient.Query(lokiURL, query, current, dayEnd, limit, "")
		if len(resp.Data.Result) == 0 {
			break
		}

		batchCount := 0
		var lastTS string

		for i, stream := range resp.Data.Result {
			for _, entry := range stream.RawValues {
				timestamp, logLine := entry[0], entry[1]
				lastTS = timestamp
				batchCount++

				// parse response
				var structuredResponse model.LogEntry
				if err := json.Unmarshal([]byte(logLine), &structuredResponse); err != nil {
					log.Println("parsing loki log error: ", err.Error())
					break
				}

				// format for specific logline, by check constants
				if lokiparser.IfLogContains(structuredResponse.Log, constants.FILTER_LOG) {
					// cleanup log and filter parsing metrics
					whitespaceClean := strings.Join(strings.Fields(logLine), " ")
					rawArrayString := strings.Split(whitespaceClean, "Value:")

					var metricsValue string
					if len(rawArrayString) == 2 {
						metricsValue = strings.TrimSpace(rawArrayString[1])
						metric, err := lokiparser.Parse(metricsValue)
						if err != nil {
							log.Println("parsing loki log error: ", err.Error())
							// break
						}

						// merge data metrics to clickhouse format
						mergo.Merge(&structuredResponse, metric, mergo.WithOverride, mergo.WithoutDereference)

						// set database time to time format
						t, err := time.Parse(time.RFC3339, structuredResponse.RequestTimestamp)
						if err != nil {
							log.Println("parsing date to time error: ", err.Error())
							// break
						}

						structuredResponse.APIType = "HTTP" // temporary hardcode, parser issue
						structuredResponse.Time = t.UTC()   // format to UTC
					}

					// add log ID, log workspace
					structuredResponse.Timestamp = timestamp
					structuredResponse.Job = stream.Stream.Job
					structuredResponse.Namespace = stream.Stream.Namespace
					structuredResponse.NodeName = stream.Stream.NodeName
					structuredResponse.Pod = stream.Stream.Pod
					structuredResponse.App = stream.Stream.App
					structuredResponse.Container = stream.Stream.Container
					structuredResponse.Filename = stream.Stream.Filename

					// push to structed json
					resp.Data.Result[i].StructedValues = append(resp.Data.Result[i].StructedValues, structuredResponse)
				}
			}
		}

		// push to clickhouse
		for _, results := range resp.Data.Result {
			for _, structedData := range results.StructedValues {
				err := batch.AppendStruct(&structedData) // Use AppendStruct for struct-based insertion
				if err != nil {
					log.Println("Failed to append structued data: ", err.Error())
				}
			}
		}

		if batchCount < limit {
			break // no more data in this day
		}

		// Move to next batch
		nextStart := lokiparser.ParseStrToTimeNano(lastTS).Add(time.Nanosecond)
		current = nextStart
	}

	// Send the batch
	if err := batch.Send(); err != nil {
		log.Println(err)
	}

	fmt.Printf("Finished day: %s\n", dayStart.Format("2006-01-02"))
}
