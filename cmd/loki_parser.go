package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"dario.cat/mergo"
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

				if containsAny(structedJson.Log) {
					log.Println("structedJson.LogstructedJson.Log")
					log.Println(structedJson.Log)
					log.Println("structedJson.LogstructedJson.Log")
					// cleanup log and filter parsing metrics
					whitespaceClean := strings.Join(strings.Fields(structedJson.Log), " ")
					rawArrayString := strings.Split(whitespaceClean, "Value:")
					var metricsValue string
					if len(rawArrayString) == 2 {
						metricsValue = strings.TrimSpace(rawArrayString[1])

						fmt.Println("metricsValuemetricsValuemetricsValue")
						fmt.Println(metricsValue)
						fmt.Println("metricsValuemetricsValuemetricsValue")

						metric, err := ConvertToApiLog(metricsValue)
						if err != nil {
							log.Panicln(err.Error())
						}

						log.Printf("Data App ID: %s", metric.APIID)
						log.Printf("Data Application ID: %s", metric.ApplicationID)
						log.Printf("Data User Agent: %s", metric.UserAgent)
						log.Printf("Data User IP: %s", metric.UserIP)

						// merge data metrics to clickhouse
						mergo.Merge(&structedJson, metric, mergo.WithOverride, mergo.WithoutDereference)
					}

					// add log ID, log workspace
					structedJson.Timestamp = timestamp
					structedJson.Job = stream.Stream.Job
					structedJson.Namespace = stream.Stream.Namespace
					structedJson.NodeName = stream.Stream.NodeName
					structedJson.Pod = stream.Stream.Pod
					structedJson.App = stream.Stream.App
					structedJson.Container = stream.Stream.Container
					structedJson.Filename = stream.Stream.Filename

					// push to structed json
					lokiResp.Data.Result[i].StructedValues = append(lokiResp.Data.Result[i].StructedValues, structedJson)
				}
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
	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO log_entry (timestamp, job, namespace, node_name, pod, app, container, filename, stream, time, apiName, proxyResponseCode, destination, apiCreatorTenantDomain, platform, apiMethod, apiVersion, gatewayType, apiCreator, responseCacheHit, backendLatency, correlationId, requestMediationLatency, keyType, apiId, applicationName, targetResponseCode, requestTimestamp, applicationOwner, userAgent, eventType, apiResourceTemplate, regionId, responseLatency, responseMediationLatency, userIp, apiContext, applicationId, apiType)")
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
