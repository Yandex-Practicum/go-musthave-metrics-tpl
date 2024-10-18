package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

type HTTPClient struct {
	host string
}

func NewHTTPClient(host string) *HTTPClient {
	return &HTTPClient{host: host}
}

func (hc *HTTPClient) SendMetrics(data []byte) {
	url := fmt.Sprintf("http://%s/update/", hc.host)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

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

	fmt.Printf("Metrics %s sent successfully\n", string(data))
}
