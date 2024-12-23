package sender

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"metrics-service/internal/server/models"
	"net/http"
	"strconv"
)

type Sender interface {
	SendJSON(metricsMap map[string]interface{}) error
}

type sender struct {
	client  *http.Client
	baseURL string
}

func NewSender(baseURL string) Sender {
	return &sender{&http.Client{}, baseURL}
}

func (s *sender) Send(metricsMap map[string]interface{}) error {
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
			return fmt.Errorf("wrong metric value type: %v", v)
		}

		url := s.baseURL + "/" + metricType + "/" + metricName + "/" + metricValStr

		client := resty.New()
		client.SetHeader("Content-type", "text/plain")

		resp, err := client.R().Post(url)
		if err != nil {
			return fmt.Errorf("failed to send request: %v", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("request %v failed: %v", url, err)
		}

	}
	fmt.Println("All metrics send to server")
	return nil
}

func (s *sender) SendJSON(metricsMap map[string]interface{}) error {
	for metricName, metricValI := range metricsMap {

		var body models.Metrics

		if metricName == "PollCount" {
			metricType := "counter"

			metricVal, ok := metricValI.(int64)
			if !ok {
				return fmt.Errorf("invalid type for counter metric: %v", metricName)
			}

			body = models.Metrics{ID: metricName, MType: metricType, Delta: &metricVal}
		} else {
			metricType := "gauge"

			metricVal, ok := metricValI.(float64)
			if !ok {
				return fmt.Errorf("invalid type for gauge metric: %v", metricName)
			}

			body = models.Metrics{ID: metricName, MType: metricType, Value: &metricVal}
		}

		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}

		// Сжатие данных в gzip
		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)

		if _, err = gzipWriter.Write(jsonData); err != nil {
			return fmt.Errorf("failed to gzip data: %v", err)
		}
		if err = gzipWriter.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %v", err)
		}

		url := s.baseURL

		client := resty.New()
		client.SetHeader("Content-type", "application/json")
		client.SetHeader("Content-Encoding", "gzip")

		// Resty автоматически сериализует в json
		resp, err := client.R().SetBody(buf.Bytes()).Post(url)
		if err != nil {
			return fmt.Errorf("failed to send request: %v", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("request %v failed: %v", url, err)
		}

	}
	fmt.Println("All metrics send to server")
	return nil
}
