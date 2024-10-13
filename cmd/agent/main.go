package main

import (
	"evgen3000/go-musthave-metrics-tpl.git/cmd/agent/provider"
	"evgen3000/go-musthave-metrics-tpl.git/internal/config"
	"log"
	"os"
	"os/signal"
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
	c := config.GetAgentConfig()
	(provider.NewAgent(c.Host, time.Duration(c.PoolInterval)*time.Second, time.Duration(c.ReportInterval)*time.Second)).Start()

}
