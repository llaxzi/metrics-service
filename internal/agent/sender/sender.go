package sender

import (
	"fmt"
	"net/http"
	"strconv"
)

type Sender interface {
	Send(metricsMap map[string]interface{})
}

type sender struct {
	client  *http.Client
	baseURL string
}

func NewSender(baseURL string) Sender {
	return &sender{&http.Client{}, baseURL}
}

func (s *sender) Send(metricsMap map[string]interface{}) {
	for metricName, metricVal := range metricsMap {

		var metricType string
		if metricName == "PollCount" {
			metricType = "counter"
		} else {
			metricType = "gauge"
		}

		var metricValStr string
		switch v := metricVal.(type) {
		case uint64:
			metricValStr = strconv.FormatUint(metricVal.(uint64), 10)
		case uint32:
			metricValStr = strconv.FormatUint(uint64(metricVal.(uint32)), 10)
		case float64:
			metricValStr = strconv.FormatFloat(metricVal.(float64), 'f', -1, 64)
		default:
			fmt.Printf("Wrong metric value type: %v", v)
		}

		url := s.baseURL + "/" + metricType + "/" + metricName + "/" + metricValStr
		request, err := http.NewRequest("POST", url, nil)
		if err != nil {
			fmt.Printf("failed to create request: %v\n", err)
		}

		request.Header.Set("Content-Type", "text/plain")

		response, err := s.client.Do(request)
		if err != nil {
			fmt.Printf("failed to send request: %v\n", err)
			continue
		}
		response.Body.Close()

		if response.StatusCode != 200 {
			fmt.Printf("request %v failed: %v\n", url, err)
			continue
		}

	}
	fmt.Println("All metrics send to server")
}
