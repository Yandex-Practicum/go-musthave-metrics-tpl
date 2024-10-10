package main

import (
	"evgen3000/go-musthave-metrics-tpl.git/cmd/agent/client"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func gracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("Shutting down gracefully")
		os.Exit(0)
	}()
}

func main() {
	gracefulShutdown()

	hostEnv, isHostEnv := os.LookupEnv("ADDRESS")
	reportIntervalEnv, isReportIntervalEn := os.LookupEnv("REPORT_INTERVAL")
	pollIntervalEnv, isPollIntervalEnv := os.LookupEnv("POLL_INTERVAL")

	hostFlag := flag.String("a", "localhost:8080", "Host IP address and port.")
	reportIntervalFlag := flag.Int("r", 10, "Report interval in seconds.")
	pollIntervalFlag := flag.Int("p", 2, "Pool interval in seconds.")
	flag.Parse()

	if isReportIntervalEn && isPollIntervalEnv && isHostEnv {
		poolInterval, err := strconv.ParseInt(pollIntervalEnv, 10, 64)
		if err != nil {
			log.Fatal("Variable is not valid:", err)
		}
		reportInterval, err := strconv.ParseInt(reportIntervalEnv, 10, 64)
		if err != nil {
			log.Fatal("Variable is not valid:", err)
		}
		(client.NewAgent(hostEnv, time.Duration(poolInterval)*time.Second, time.Duration((reportInterval))*time.Second)).Start()
	} else {
		(client.NewAgent(*hostFlag, time.Duration(*pollIntervalFlag)*time.Second, time.Duration((*reportIntervalFlag))*time.Second)).Start()
	}

}
