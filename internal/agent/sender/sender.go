package sender

import (
	"fmt"
	"github.com/go-resty/resty/v2"
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

		client := resty.New()
		client.SetHeader("Content-type", "text/plain")

		resp, err := client.R().Post(url)
		if err != nil {
			fmt.Printf("failed to send request: %v\n", err)
			continue
		}

		if resp.StatusCode() != 200 {
			fmt.Printf("request %v failed: %v\n", url, err)
			continue
		}

	}
	fmt.Println("All metrics send to server")
}
