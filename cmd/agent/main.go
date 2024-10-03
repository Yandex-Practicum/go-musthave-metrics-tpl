package main

import (
	"evgen3000/go-musthave-metrics-tpl.git/cmd/agent/client"
	"time"
)

func main() {
	agent := client.NewAgent(2*time.Second, 10*time.Second)
	agent.Start()

}
