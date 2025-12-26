package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"dario.cat/mergo"
	"github.com/prakasa1904/loki-scraper/internal/connection"
	"github.com/prakasa1904/loki-scraper/internal/lokiclient"
	"github.com/prakasa1904/loki-scraper/internal/lokiparser"
	"github.com/prakasa1904/loki-scraper/internal/model"
)

type ExtendedLogEntry struct {
	APIName                  string `json:"apiName" ch:"apiName"`
	ProxyResponseCode        string `json:"proxyResponseCode" ch:"proxyResponseCode"`
	Destination              string `json:"destination" ch:"destination"`
	APICreatorTenantDomain   string `json:"apiCreatorTenantDomain" ch:"apiCreatorTenantDomain"`
	Platform                 string `json:"platform" ch:"platform"`
	APIMethod                string `json:"apiMethod" ch:"apiMethod"`
	APIVersion               string `json:"apiVersion" ch:"apiVersion"`
	GatewayType              string `json:"gatewayType" ch:"gatewayType"`
	APICreator               string `json:"apiCreator" ch:"apiCreator"`
	ResponseCacheHit         string `json:"responseCacheHit" ch:"responseCacheHit"`
	BackendLatency           string `json:"backendLatency" ch:"backendLatency"`
	CorrelationID            string `json:"correlationId" ch:"correlationId"`
	RequestMediationLatency  string `json:"requestMediationLatency" ch:"requestMediationLatency"`
	KeyType                  string `json:"keyType" ch:"keyType"`
	APIID                    string `json:"apiId" ch:"apiId"`
	ApplicationName          string `json:"applicationName" ch:"applicationName"`
	TargetResponseCode       string `json:"targetResponseCode" ch:"targetResponseCode"`
	RequestTimestamp         string `json:"requestTimestamp" ch:"requestTimestamp"`
	ApplicationOwner         string `json:"applicationOwner" ch:"applicationOwner"`
	UserAgent                string `json:"userAgent" ch:"userAgent"`
	EventType                string `json:"eventType" ch:"eventType"`
	APIResourceTemplate      string `json:"apiResourceTemplate" ch:"apiResourceTemplate"`
	RegionID                 string `json:"regionId" ch:"regionId"`
	ResponseLatency          string `json:"responseLatency" ch:"responseLatency"`
	ResponseMediationLatency string `json:"responseMediationLatency" ch:"responseMediationLatency"`
	UserIP                   string `json:"userIp" ch:"userIp"`
	APIContext               string `json:"apiContext" ch:"apiContext"`
	ApplicationID            string `json:"applicationId" ch:"applicationId"`
	APIType                  string `json:"apiType" ch:"apiType"`
}

type LogEntry struct {
	Timestamp string    `json:"timestamp" ch:"timestamp"`
	Job       string    `json:"job" ch:"job"`
	Namespace string    `json:"namespace" ch:"namespace"`
	NodeName  string    `json:"node_name" ch:"node_name"`
	Pod       string    `json:"pod" ch:"pod"`
	App       string    `json:"app" ch:"app"`
	Container string    `json:"container" ch:"container"`
	Filename  string    `json:"filename" ch:"filename"`
	Log       string    `json:"log"` // Skip raw data from clickhouse (hemat bebz...)
	Stream    string    `json:"stream" ch:"stream"`
	Time      time.Time `json:"time" ch:"time"`
	// extra data
	ExtendedLogEntry
}

type LokiStreamResponseChild struct {
	App       string `json:"app"`
	Container string `json:"container"`
	Filename  string `json:"filename"`
	Job       string `json:"job"`
	Namespace string `json:"namespace"`
	NodeName  string `json:"node_name"`
	Pod       string `json:"pod"`
}

// Define structs to match Loki API response
type LokiResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream         LokiStreamResponseChild `json:"stream"`
			RawValues      [][]string              `json:"values"`
			StructedValues []LogEntry              `json:"log"`
		} `json:"result"`
	} `json:"data"`
}

func setLimit() int {
	var limit = os.Getenv("LOKI_LIMIT")
	if limit == "" {
		return 0
	}

	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		return 0
	}

	return intLimit
}

func lokiParser(ctx context.Context, query string, startDate string, endDate string) error {
	var LOKI_URL = os.Getenv("LOKI_URL")
	if LOKI_URL == "" {
		return fmt.Errorf("no Loki URL set from env LOKI_URL")
	}

	var CLICKHOUSE_TABLE = os.Getenv("CLICKHOUSE_TABLE")
	if CLICKHOUSE_TABLE == "" {
		return fmt.Errorf("no clickhouse table set from env variable CLICKHOUSE_TABLE")
	}

	start, _ := time.Parse(time.RFC3339, startDate)
	end, _ := time.Parse(time.RFC3339, endDate)

	conn, err := connection.Connect()
	if err != nil {
		log.Panicln("Error DB connection: ", err.Error())
	}

	current := start

	for {
		// If reached or passed end date → stop
		if current.After(end) {
			fmt.Println("Finished fetching all logs.")
			return nil
		}

		lokiResp := lokiclient.Query(LOKI_URL, query, current, end, setLimit(), "")

		if len(lokiResp.Data.Result) == 0 {
			fmt.Println("No more results.")
			return nil
		}

		// Prepare the batch insert statement
		batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO "+CLICKHOUSE_TABLE+" (timestamp, job, namespace, node_name, pod, app, container, filename, stream, time, apiName, proxyResponseCode, destination, apiCreatorTenantDomain, platform, apiMethod, apiVersion, gatewayType, apiCreator, responseCacheHit, backendLatency, correlationId, requestMediationLatency, keyType, apiId, applicationName, targetResponseCode, requestTimestamp, applicationOwner, userAgent, eventType, apiResourceTemplate, regionId, responseLatency, responseMediationLatency, userIp, apiContext, applicationId, apiType)")
		if err != nil {
			log.Fatal("Error on prepare batch: ", err.Error())
		}

		// on success, formating data
		if lokiResp.Status == "success" {
			for i, stream := range lokiResp.Data.Result {
				for _, entry := range stream.RawValues {
					timestamp, logLine := entry[0], entry[1]

				// parse response
				var structuredResponse model.LogEntry
				if json.Valid([]byte(logLine)) {
					_ = json.Unmarshal([]byte(logLine), &structuredResponse)
				}
				if structuredResponse.Log == "" {
					structuredResponse.Log = logLine
				}

				// cleanup log and filter parsing metrics
				whitespaceClean := strings.Join(strings.Fields(structuredResponse.Log), " ")
				rawArrayString := strings.Split(whitespaceClean, "Value:")
				var metricsValue string
					if len(rawArrayString) == 2 {
						metricsValue = strings.TrimSpace(rawArrayString[1])

						metric, err := lokiparser.Parse(metricsValue)
						if err != nil {
							log.Println("Error to parse loki metrics, with error: ", err.Error())
						}

						// merge data metrics to clickhouse
						mergo.Merge(&structuredResponse, metric, mergo.WithOverride, mergo.WithoutDereference)

						// convert date string from metric to date, and set to the database

						ns, err := strconv.ParseInt(timestamp, 10, 64)
						if err != nil {
							// fallback: try parse float seconds
							f, err2 := strconv.ParseFloat(timestamp, 64)
							if err2 == nil {
								ns = int64(f * 1e9)
							} else {
								// skip if cannot parse
								continue
							}
						}

						// parse with time.RFC3339, to follow clickhouse datetime format
						t := time.Unix(0, ns)
						parsedTime, err := time.Parse(time.RFC3339, t.Format(time.RFC3339))
						if err != nil {
							log.Println("Error parsing date:", err)
							continue
						}
						structuredResponse.Time = parsedTime.UTC()
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
					lokiResp.Data.Result[i].StructedValues = append(lokiResp.Data.Result[i].StructedValues, structuredResponse)
				}
			}
		} else {
			log.Println("loki API error: ", lokiResp.Status)
			break
		}

		for _, results := range lokiResp.Data.Result {
			for _, structedData := range results.StructedValues {
				err := batch.AppendStruct(&structedData) // Use AppendStruct for struct-based insertion
				if err != nil {
					log.Println("Failed to append structued data: ", err.Error())
				}
			}
		}

		lastTS := lokiResp.Data.Result[0].RawValues[len(lokiResp.Data.Result[0].RawValues)-1][0]
		current = timeFromLoki(lastTS).Add(1 * time.Nanosecond)

		// Send the batch
		if err := batch.Send(); err != nil {
			log.Fatalln("Failed to store data to Clickhose, error: ", err.Error())
		}
	}

	return nil
}

// Loki timestamp = nanoseconds string
func timeFromLoki(ns string) time.Time {
	n, _ := time.ParseDuration(ns + "ns")
	return time.Unix(0, n.Nanoseconds())
}

// function to append string to file
// func appendToFile(filename string, data string) error {
// 	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()

// 	if _, err := f.WriteString(data + "\n"); err != nil {
// 		return err
// 	}
// 	return nil
// }
