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

// Package handlers contains HTTP handlers for metrics operations.
package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/storage"
)

// ExampleUpdateGauge demonstrates how to update a gauge metric.
func Example_updateGauge() {
	// Create a new storage
	storage := storage.NewMemStorage()

	// Create a new handler
	handler := NewMetricHandlers(storage)

	// Create a test request to update a gauge metric
	req := httptest.NewRequest("POST", "/update/gauge/test_gauge/42.5", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.UpdateHandler(w, req)

	// Print the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	// Output:
	// Status: 200
	// Response: OK
}

// ExampleUpdateCounter demonstrates how to update a counter metric.
func Example_updateCounter() {
	// Create a new storage
	storage := storage.NewMemStorage()

	// Create a new handler
	handler := NewMetricHandlers(storage)

	// Create a test request to update a counter metric
	req := httptest.NewRequest("POST", "/update/counter/test_counter/10", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.UpdateHandler(w, req)

	// Print the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	// Output:
	// Status: 200
	// Response: OK
}

// ExampleGetGauge demonstrates how to retrieve a gauge metric value.
func Example_getGauge() {
	// Create a new storage
	storage := storage.NewMemStorage()

	// First update a gauge metric
	storage.UpdateGauge("test_gauge", 42.5)

	// Create a new handler
	handler := NewMetricHandlers(storage)

	// Create a test request to get a gauge metric value
	req := httptest.NewRequest("GET", "/value/gauge/test_gauge", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.ValueHandler(w, req)

	// Print the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	// Output:
	// Status: 200
	// Response: 42.5
}

// ExampleGetCounter demonstrates how to retrieve a counter metric value.
func Example_getCounter() {
	// Create a new storage
	storage := storage.NewMemStorage()

	// First update a counter metric
	storage.UpdateCounter("test_counter", 10)

	// Create a new handler
	handler := NewMetricHandlers(storage)

	// Create a test request to get a counter metric value
	req := httptest.NewRequest("GET", "/value/counter/test_counter", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.ValueHandler(w, req)

	// Print the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	// Output:
	// Status: 200
	// Response: 10
}

// TestExamples verifies that the examples compile and run correctly.
func TestExamples(t *testing.T) {
	// This test ensures that the examples compile and don't panic
	// We don't actually run the examples, but we verify they compile

	// Test that the basic functionality works
	storage := storage.NewMemStorage()
	handler := NewMetricHandlers(storage)

	// Test update gauge
	req := httptest.NewRequest("POST", "/update/gauge/test/42.5", nil)
	w := httptest.NewRecorder()
	handler.UpdateHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test get gauge
	req = httptest.NewRequest("GET", "/value/gauge/test", nil)
	w = httptest.NewRecorder()
	handler.ValueHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
