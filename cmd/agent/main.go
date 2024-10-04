package main

import (
	"evgen3000/go-musthave-metrics-tpl.git/cmd/agent/client"
	"flag"
	"time"
)

func main() {
	hostFlag := flag.String("a", "localhost:8080", "Host IP address and port.")
	reportInterval := flag.Int("r", 10, "Report interval in seconds.")
	pollInterval := flag.Int("p", 2, "Pool interval in seconds.")
	flag.Parse()
	
	agent := client.NewAgent(*hostFlag ,time.Duration(*pollInterval)*time.Second, time.Duration((*reportInterval)) * time.Second)
	agent.Start()

}
