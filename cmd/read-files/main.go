package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"dario.cat/mergo"
	"github.com/prakasa1904/loki-scraper/internal/connection"
	"github.com/prakasa1904/loki-scraper/internal/lokiparser"
	"github.com/prakasa1904/loki-scraper/internal/model"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	var GOOGLE_APPLICATION_CREDENTIALS = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if GOOGLE_APPLICATION_CREDENTIALS == "" {
		log.Fatalf("no google application credentials set from env variable GOOGLE_APPLICATION_CREDENTIALS")
	}

	b, err := os.ReadFile(GOOGLE_APPLICATION_CREDENTIALS)
	if err != nil {
		log.Fatalf("read credentials file: %v", err)
	}

	var CLICKHOUSE_TABLE = os.Getenv("CLICKHOUSE_TABLE")
	if CLICKHOUSE_TABLE == "" {
		log.Fatalf("no clickhouse table set from env variable CLICKHOUSE_TABLE")
	}

	conn, err := connection.Connect()
	if err != nil {
		log.Panicln("Error DB connection: ", err.Error())
	}

	// Choose the scope you need
	scopes := []string{drive.DriveScope} // or drive.DriveReadonlyScope / drive.DriveFileScope

	// Create JWT config from JSON and build HTTP client
	jwtCfg, err := google.JWTConfigFromJSON(b, scopes...)
	if err != nil {
		log.Fatalf("JWTConfigFromJSON: %v", err)
	}
	client := jwtCfg.Client(ctx)

	// Create Drive service using the client
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("drive.NewService: %v", err)
	}

	// List first 10 files visible to the service account
	res, err := srv.Files.List().PageSize(10).
		Fields("files(id,name,owners,mimeType,createdTime)").Do()
	if err != nil {
		log.Fatalf("Files.List: %v", err)
	}
	if len(res.Files) == 0 {
		fmt.Println("No files found.")
		return
	}

	fmt.Println("Files:")
	for _, f := range res.Files {
		fmt.Printf("- %s (%s) owners=%s mime=%s created=%s\n", f.Name, f.Id, f.Owners[0].DisplayName, f.MimeType, f.CreatedTime)

		if f.MimeType == "text/plain" {
			readFileContent, err := readFileContent(srv, f.Id)
			if err != nil {
				log.Fatalf("readFileContent: %v", err)
			}

			reader := bufio.NewReader(bytes.NewReader(readFileContent))
			fmt.Printf("Content of %s:\n", f.Name)

			var dataLength int = 0
			for {
				line, err := reader.ReadString('\n')
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatalf("ReadString: %v", err)
				}

				// Prepare the batch insert statement
				batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO "+CLICKHOUSE_TABLE+" (timestamp, job, namespace, node_name, pod, app, container, filename, stream, time, apiName, proxyResponseCode, destination, apiCreatorTenantDomain, platform, apiMethod, apiVersion, gatewayType, apiCreator, responseCacheHit, backendLatency, correlationId, requestMediationLatency, keyType, apiId, applicationName, targetResponseCode, requestTimestamp, applicationOwner, userAgent, eventType, apiResourceTemplate, regionId, responseLatency, responseMediationLatency, userIp, apiContext, applicationId, apiType)")
				if err != nil {
					log.Fatal("Error on prepare batch: ", err.Error())
				}

				// parse loki line here
				var formattedLog model.LogEntryFromFile
				if err := json.Unmarshal([]byte(line), &formattedLog); err != nil {
					// skip this line if failed to parse
					log.Println("Skip data, error json.Unmarshal:", line)
					continue
				}

				// cleanup log and filter parsing metrics
				whitespaceClean := strings.Join(strings.Fields(formattedLog.Log), " ")
				rawArrayString := strings.Split(whitespaceClean, "Value:")

				var metricsValue string
				if len(rawArrayString) == 2 {
					metricsValue = strings.TrimSpace(rawArrayString[1])

					metrics, err := lokiparser.Parse(metricsValue)
					if err != nil {
						// skip this line if failed to parse
						log.Println("Skip data, error lokiparser.Parse:", line)
						continue
					}

					// merge data metrics to clickhouse structure
					mergo.Merge(&formattedLog, metrics, mergo.WithOverride, mergo.WithoutDereference)

					// format string to date time.Time
					parsedTime, err := time.Parse(time.RFC3339, formattedLog.TimeString)
					if err != nil {
						fmt.Println("Error parsing time:", err)
						continue
					}
					metrics.Time = parsedTime.UTC()

					// hardcoded data, not found in file backup
					metrics.Namespace = "splp-prod"
					metrics.NodeName = "unknown-node"
					metrics.Pod = "unknown-pod"

					// For debug purpose only, uncomment to see parsed data
					// fmt.Printf("----- Parsed Log Entry -----\n")
					// fmt.Printf("metrics.TimeString: %+v\n", formattedLog.TimeString)
					// fmt.Printf("metrics.Time: %+v\n", metrics.Time.String())
					// fmt.Printf("metrics.App: %+v\n", metrics.App)
					// fmt.Printf("metrics.Stream: %+v\n", metrics.Stream)
					// fmt.Printf("metrics.Filename: %+v\n", metrics.Filename)
					// fmt.Printf("metrics.Container: %+v\n", metrics.Container)
					// fmt.Printf("metrics.Platform: %+v\n", metrics.Platform)
					// fmt.Printf("metrics.APIVersion: %+v\n", metrics.APIVersion)
					// fmt.Printf("metrics.CorrelationID: %+v\n", metrics.CorrelationID)
					// fmt.Printf("metrics.UserAgent: %+v\n", metrics.UserAgent)
					// fmt.Printf("metrics.APIName: %+v\n", metrics.APIName)
					// fmt.Printf("metrics.APIMethod: %+v\n", metrics.APIMethod)
					// fmt.Printf("metrics.APICreator: %+v\n", metrics.APICreator)
					// fmt.Printf("metrics.APICreatorTenantDomain: %+v\n", metrics.APICreatorTenantDomain)
					// fmt.Printf("metrics.APIContext: %+v\n", metrics.APIContext)
					// fmt.Printf("metrics.Destination: %+v\n", metrics.Destination)
					// fmt.Printf("metrics.targetResponseCode: %+v\n", metrics.TargetResponseCode)
					// fmt.Printf("metrics.APIType: %+v\n", metrics.APIType)
					// fmt.Printf("metrics.ResponseMediationLatency: %+v\n", metrics.ResponseMediationLatency)
					// fmt.Printf("metrics.ApplicationOwner: %+v\n", metrics.ApplicationOwner)
					// fmt.Printf("----- Parsed Log Entry -----\n")

					err = batch.AppendStruct(metrics) // Use AppendStruct for struct-based insertion
					if err != nil {
						log.Println("Failed to append structued data: ", err.Error())
					}

					// send batch to clickhouse
					if err := batch.Send(); err != nil {
						log.Println("Failed to send batch: ", err.Error())
					} else {
						fmt.Printf("Successfully inserted batch to ClickHouse\n")
					}

					dataLength++
					log.Printf("We're in line numb: %d", dataLength)

					// time.Sleep(5 * time.Second) // just to slow down output for readability
				}
			}
		}
	}
}

func readFileContent(srv *drive.Service, fileID string) ([]byte, error) {
	resp, err := srv.Files.Get(fileID).Download()
	if err != nil {
		return nil, fmt.Errorf("Files.Get.Download: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %v", err)
	}
	return data, nil
}
