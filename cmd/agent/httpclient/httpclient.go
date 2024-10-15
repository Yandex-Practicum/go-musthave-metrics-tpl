package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type HttpClient struct {
	host string
}

func NewHttpClient(host string) *HttpClient {
	return &HttpClient{host: host}
}

func (hc *HttpClient) SendMetrics(metricType, metricName string, value float64) {
	metricValue := strconv.FormatFloat(value, 'f', -1, 64)
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", hc.host, metricType, metricName, metricValue)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	fmt.Printf("Metrics %s (%s) with value %s sent successfully\n", metricName, metricType, metricValue)
}
