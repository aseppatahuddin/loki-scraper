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
	"github.com/prakasa1904/loki-scraper/internal/connection"
	"github.com/prakasa1904/loki-scraper/internal/constants"
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

func setLimit() string {
	var limit = os.Getenv("LOKI_LIMIT")
	if limit == "" {
		return ""
	}

	return limit
}

func lokiParser(ctx context.Context, query string, startDate string, endDate string) error {
	var LOKI_URL = os.Getenv("LOKI_URL")
	if LOKI_URL == "" {
		return fmt.Errorf("no Loki URL set from env LOKI_URL")
	}

	var CLICKHOUSE_TABLE = os.Getenv("CLICKHOUSE_TABLE")
	if CLICKHOUSE_TABLE == "" {
		return fmt.Errorf("no clickhouse table set from env variable.")
	}

	lokiResp := lokiclient.QueryRaw(LOKI_URL, query, startDate, endDate, setLimit(), "")

	// on success, formating data
	if lokiResp.Status == "success" {
		for i, stream := range lokiResp.Data.Result {
			for _, entry := range stream.RawValues {
				timestamp, logLine := entry[0], entry[1]

				// parse response
				var structuredResponse model.LogEntry
				if err := json.Unmarshal([]byte(logLine), &structuredResponse); err != nil {
					return fmt.Errorf("parsing loki log error: %s", err.Error())
				}

				if lokiparser.IfLogContains(structuredResponse.Log, constants.FILTER_LOG) {
					// cleanup log and filter parsing metrics
					whitespaceClean := strings.Join(strings.Fields(structuredResponse.Log), " ")
					rawArrayString := strings.Split(whitespaceClean, "Value:")
					var metricsValue string
					if len(rawArrayString) == 2 {
						metricsValue = strings.TrimSpace(rawArrayString[1])

						metric, err := lokiparser.Parse(metricsValue)
						if err != nil {
							log.Panicln(err.Error())
						}

						// merge data metrics to clickhouse
						mergo.Merge(&structuredResponse, metric, mergo.WithOverride, mergo.WithoutDereference)

						// convert date string from metric to date, and set to the database
						// loc, _ := time.LoadLocation("Asia/Jakarta") // Uncommnet to set UTC+7
						t, err := time.Parse(time.RFC3339, metric.RequestTimestamp)
						if err != nil {
							log.Panicln(err.Error())
						}

						// tLocal := t.In(loc)        // Uncommnet to set UTC+7
						// structuredResponse.Time = tLocal // Uncommnet to set UTC+7
						structuredResponse.Time = t.UTC() // format to UTC
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
		}
	} else {
		return fmt.Errorf("loki API error: %s", lokiResp.Status)
	}

	conn, err := connection.Connect()
	if err != nil {
		log.Panicln("Error DB connection: ", err.Error())
	}

	// Prepare the batch insert statement
	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO "+CLICKHOUSE_TABLE+" (timestamp, job, namespace, node_name, pod, app, container, filename, stream, time, apiName, proxyResponseCode, destination, apiCreatorTenantDomain, platform, apiMethod, apiVersion, gatewayType, apiCreator, responseCacheHit, backendLatency, correlationId, requestMediationLatency, keyType, apiId, applicationName, targetResponseCode, requestTimestamp, applicationOwner, userAgent, eventType, apiResourceTemplate, regionId, responseLatency, responseMediationLatency, userIp, apiContext, applicationId, apiType)")
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
