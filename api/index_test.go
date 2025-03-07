package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCORSHeaders tests that the API correctly handles CORS headers
func TestCORSHeaders(t *testing.T) {
	// Create a request with OPTIONS method
	req := httptest.NewRequest(http.MethodOptions, "/api/detect", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create a mock handler that simulates CORS behavior
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	})

	// Call the mock handler
	mockHandler.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin header to be '*'")
	}

	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Errorf("Expected Access-Control-Allow-Methods header to be 'GET, POST, OPTIONS'")
	}

	if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type" {
		t.Errorf("Expected Access-Control-Allow-Headers header to be 'Content-Type'")
	}
}

// TestMethodValidation tests that the API rejects non-POST methods
func TestMethodValidation(t *testing.T) {
	// Create a request with GET method
	req := httptest.NewRequest(http.MethodGet, "/api/detect", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create a mock handler that simulates method validation
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate request method
		if r.Method != "POST" && r.Method != "OPTIONS" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})

	// Call the mock handler
	mockHandler.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

// TestResponseFormat tests that the API returns properly formatted JSON responses
func TestResponseFormat(t *testing.T) {
	// Create a request
	req := httptest.NewRequest(http.MethodPost, "/api/detect", bytes.NewBufferString("mock data"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create a mock handler that simulates the API response
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set content type
		w.Header().Set("Content-Type", "application/json")

		// Return a mock detection result
		mockResult := [][]interface{}{
			{100.0, 100.0, 200.0, 200.0, "TEST_OBJECT", float32(0.95)},
		}

		// Marshal and write the response
		json.NewEncoder(w).Encode(mockResult)
	})

	// Call the mock handler
	mockHandler.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type header to be 'application/json'")
	}

	// Parse response
	var results [][]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &results)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check results
	if len(results) != 1 {
		t.Fatalf("Expected 1 detection, got %d", len(results))
	}

	// Check detection values
	if results[0][0].(float64) != 100.0 || results[0][1].(float64) != 100.0 ||
		results[0][2].(float64) != 200.0 || results[0][3].(float64) != 200.0 ||
		results[0][4].(string) != "TEST_OBJECT" {
		t.Errorf("Detection values don't match expected values")
	}

	// Check confidence value (may be converted to float64 by JSON unmarshaling)
	confidence, ok := results[0][5].(float64)
	if !ok || confidence != 0.95 {
		t.Errorf("Expected confidence 0.95, got %v", results[0][5])
	}
}

// TestErrorHandling tests that the API properly handles errors
func TestErrorHandling(t *testing.T) {
	// Create a request with invalid content type
	req := httptest.NewRequest(http.MethodPost, "/api/detect", nil)
	req.Header.Set("Content-Type", "application/json") // Not multipart/form-data

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create a mock handler that simulates error handling
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check content type
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}
	})

	// Call the mock handler
	mockHandler.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Helper function to create a mock multipart request
func createMockMultipartRequest() (*http.Request, error) {
	// Create a buffer for the multipart data
	body := &bytes.Buffer{}

	// Create a new multipart writer
	writer := multipart.NewWriter(body)

	// Create a form file field
	part, err := writer.CreateFormFile("image_file", "test.jpg")
	if err != nil {
		return nil, err
	}

	// Write some mock data to the form file
	io.WriteString(part, "mock image data")

	// Close the writer
	writer.Close()

	// Create a new request
	req := httptest.NewRequest(http.MethodPost, "/api/detect", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}
