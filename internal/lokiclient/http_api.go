package lokiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/prakasa1904/loki-scraper/internal/model"
	"moul.io/http2curl"
)

func setLimit(limit int) int {
	if limit == 0 {
		return 100
	}

	return limit
}

func setDirection(direction string) string {
	if direction == "" {
		return "forward"
	}

	return direction
}

func QueryRaw(lokiURL string, query string, start string, end string, limit string, direction string) model.LokiResponse {
	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", limit)
	params.Set("start", start)
	params.Set("end", end)
	params.Set("direction", setDirection(direction))

	reqURL := lokiURL + "?" + params.Encode()

	resp, err := http.Get(reqURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// convert to curl format for debuging purpose
	command, _ := http2curl.GetCurlCommand(resp.Request)
	fmt.Println("Curl Format:", command)

	body, _ := io.ReadAll(resp.Body)

	var out model.LokiResponse
	json.Unmarshal(body, &out)

	return out
}

func Query(lokiURL string, query string, start, end time.Time, limit int, direction string) model.LokiResponse {
	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", setLimit(limit)))
	params.Set("start", fmt.Sprint(start.UnixNano()))
	params.Set("end", fmt.Sprint(end.UnixNano()))
	params.Set("direction", setDirection(direction))

	reqURL := lokiURL + "?" + params.Encode()

	resp, err := http.Get(reqURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var LOKI_DEBUG = os.Getenv("LOKI_DEBUG")
	if LOKI_DEBUG == "true" {
		// convert to curl format for debuging purpose
		command, _ := http2curl.GetCurlCommand(resp.Request)
		fmt.Println("Curl Format:", command)
	}

	body, _ := io.ReadAll(resp.Body)

	var out model.LokiResponse
	json.Unmarshal(body, &out)

	return out
}
