package model

import "time"

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

type LogEntryFromFile struct {
	Timestamp  string    `json:"timestamp" ch:"timestamp"`
	Job        string    `json:"job" ch:"job"`
	Namespace  string    `json:"namespace" ch:"namespace"`
	NodeName   string    `json:"node_name" ch:"node_name"`
	Pod        string    `json:"pod" ch:"pod"`
	App        string    `json:"app" ch:"app"`
	Container  string    `json:"container" ch:"container"`
	Filename   string    `json:"filename" ch:"filename"`
	Log        string    `json:"log"` // Skip raw data from clickhouse (hemat bebz...)
	Stream     string    `json:"stream" ch:"stream"`
	Time       time.Time `ch:"time"`
	TimeString string    `json:"time"`
	// extra data
	ExtendedLogEntry
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
