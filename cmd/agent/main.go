package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"evgen3000/go-musthave-metrics-tpl.git/cmd/agent/collector"
	"evgen3000/go-musthave-metrics-tpl.git/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Получен сигнал: %v. Завершаем работу...", sig)
		cancel()
	}()

	c := config.GetAgentConfig()
	agent := collector.NewAgent(c.Host, time.Duration(c.PoolInterval)*time.Second, time.Duration(c.ReportInterval)*time.Second)
	agent.Start(ctx)
}
