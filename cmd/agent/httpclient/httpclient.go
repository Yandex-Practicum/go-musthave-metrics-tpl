package httpclient

import (
	"bytes"
	"compress/gzip"
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

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write(data)
	if err != nil {
		fmt.Println("Error compressing request:", err)
		return
	}
	err = gzipWriter.Close()
	if err != nil {
		fmt.Println("Error closing gzip writer:", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
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
