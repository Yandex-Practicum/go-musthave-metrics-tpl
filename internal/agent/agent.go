package agent

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
	"yandex-go-musthave-metrics-tpl/internal/model"
	"yandex-go-musthave-metrics-tpl/internal/service/agent"
)

func Start() {
	var (
		metrics map[string]model.GaugeT
		poolCnt = 0
		urlReq  *url.URL
	)
	serverHost := flag.String("a", "localhost:8080", "Set the path to the metrics endpoint")
	reportInterval := flag.Int("r", 10, "Set report interval in seconds")
	poolInterval := flag.Int("p", 2, "Set pool interval in seconds")
	flag.Parse()

	if val, ok := os.LookupEnv("ADDRESS"); ok {
		*serverHost = val
	}
	if val, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		valInt, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("Could not parse REPORT_INTERVAL environment variable, %v", err)
			return
		}
		*reportInterval = valInt
	}
	if val, ok := os.LookupEnv("POOL_INTERVAL"); ok {
		valInt, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("Could not parse POOL_INTERVAL environment variable, %v", err)
			return
		}
		*poolInterval = valInt
	}

	fmt.Println("Agent listening on", *serverHost)
	client := http.Client{}
	tPool := time.NewTicker((time.Duration)(*poolInterval) * time.Second)
	tSend := time.NewTicker((time.Duration)(*reportInterval) * time.Second)

	for {
		select {
		case <-tPool.C:
			fmt.Println("pool")
			metrics = agent.GetMetrics()
			poolCnt++
		case <-tSend.C:
			fmt.Println("send")
			for name, value := range metrics {
				urlReq = &url.URL{Scheme: "http", Host: *serverHost, Path: fmt.Sprintf("/update/%s/%s/%v", model.Gauge, name, value)}
				resp, err := client.Post(urlReq.String(), "text/plain", nil)
				if err != nil {
					log.Printf("Could not send request to server, %v/n", err)
				} else {
					resp.Body.Close()
				}
			}
			urlReq = &url.URL{Scheme: "http", Host: *serverHost, Path: fmt.Sprintf("/update/%s/%s/%v", model.Counter, "PollCount", poolCnt)}
			resp, err := client.Post(urlReq.String(), "text/plain", nil)
			if err != nil {
				log.Printf("Could not send request to server, %v/n", err)
			} else {
				resp.Body.Close()
			}

			poolCnt = 0
		}

	}
}
