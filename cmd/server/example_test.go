// Copyright 2023 Your Name
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main contains the metrics server implementation.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/audit"
)

// ExampleJSONUpdate demonstrates how to update metrics using JSON API.
func Example_jsonUpdate() {
	// Create a new storage
	storage := NewMetricsStorage()

	// Create a new server
	server := NewServer(storage, &ServerConfig{
		Address:       "localhost:8080",
		StoreInterval: 0,
		FileStorage:   "",
		Restore:       false,
		DatabaseDSN:   "",
		AuditFile:     "",
		AuditURL:      "",
	}, []audit.Auditor{})

	// Create a JSON payload for updating a gauge
	payload := map[string]interface{}{
		"id":    "test_gauge",
		"type":  "gauge",
		"value": 42.5,
	}

	jsonPayload, _ := json.Marshal(payload)

	// Create a test request to update a metric via JSON
	req := httptest.NewRequest("POST", "/update", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	server.updateMetricJSONHandler(w, req)

	// Print the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	// Output:
	// Status: 200
	// Response: {"status":"ok"}
}

// ExampleJSONValue demonstrates how to retrieve metrics using JSON API.
func Example_jsonValue() {
	// Create a new storage
	storage := NewMetricsStorage()

	// First update a gauge metric
	storage.gauges["test_gauge"] = 42.5

	// Create a new server
	server := NewServer(storage, &ServerConfig{
		Address:       "localhost:8080",
		StoreInterval: 0,
		FileStorage:   "",
		Restore:       false,
		DatabaseDSN:   "",
		AuditFile:     "",
		AuditURL:      "",
	}, []audit.Auditor{})

	// Create a JSON payload for retrieving a metric
	payload := map[string]interface{}{
		"id":   "test_gauge",
		"type": "gauge",
	}

	jsonPayload, _ := json.Marshal(payload)

	// Create a test request to get a metric via JSON
	req := httptest.NewRequest("POST", "/value", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	server.valueMetricJSONHandler(w, req)

	// Print the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	// Output:
	// Status: 200
	// Response: {"id":"test_gauge","type":"gauge","value":42.5}
}

// ExamplePeriodicSave demonstrates periodic saving of metrics.
func Example_periodicSave() {
	// Create a new storage
	storage := NewMetricsStorage()

	// Create a new server with periodic save
	_ = NewServer(storage, &ServerConfig{
		Address:       "localhost:8080",
		StoreInterval: 0,
		FileStorage:   "metrics.json",
		Restore:       false,
		DatabaseDSN:   "",
		AuditFile:     "",
		AuditURL:      "",
	}, []audit.Auditor{})

	// Simulate updating a metric
	storage.gauges["example_gauge"] = 100.0

	// In real usage, the periodic save would happen automatically
	// This example shows how to configure it
	fmt.Println("Server configured with periodic save to metrics.json")

	// Output:
	// Server configured with periodic save to metrics.json
}
