package cmd_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"xixuanyunsign/cmd"
)

func TestMonthReportUploadSelectFile(t *testing.T) {
	// Define mock data
	expectedURI := "https://example.com/uploaded/image.jpg"
	mockResponseBody := []byte(fmt.Sprintf(`{
		"code": 0,
		"message": "success",
		"data": {
			"uri": "%s"
		}
	}`, expectedURI))

	// Create mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/file/form?token=dummy_token", r.URL.String())

		// Read request body
		_, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)

		// Validate request body (optional for more comprehensive testing)
		// ...

		// Write mock response
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseBody)
	}))
	defer ts.Close()

	// Test function
	uri := cmd.MonthReportUploadSelectFile("dummy_file_path", "dummy_token")
	assert.Equal(t, expectedURI, uri)
}

func TestMonthReportUploadSelectFile_Error(t *testing.T) {
	// Define mock data
	mockResponseBody := []byte(`{
		"code": 1,
		"message": "upload failed"
	}`)

	// Create mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(mockResponseBody)
	}))
	defer ts.Close()

	// Test function with error
	uri := cmd.MonthReportUploadSelectFile("dummy_file_path", "dummy_token")
	assert.Empty(t, uri)
}

func TestMonthReportUploadSelectFile_InvalidJSON(t *testing.T) {
	// Define mock data
	mockResponseBody := []byte("invalid json")

	// Create mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseBody)
	}))
	defer ts.Close()

	// Test function with invalid JSON response
	uri := cmd.MonthReportUploadSelectFile("dummy_file_path", "dummy_token")
	assert.Empty(t, uri)
}
func TestPostRequest(t *testing.T) {
	// Create a test server to mock the API endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check HTTP method
		if r.Method != http.MethodPost {
			t.Errorf("Expected method POST but got %s", r.Method)
		}

		// Read and validate the request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		defer r.Body.Close()

		// Validate specific form data fields
		if !bytes.Contains(body, []byte("business_type=month")) {
			t.Errorf("Expected 'business_type=month' in request body, but it was missing")
		}
		if !bytes.Contains(body, []byte("start_date=2024/12/01")) {
			t.Errorf("Expected 'start_date=2024/12/01' in request body, but it was missing")
		}
		if !bytes.Contains(body, []byte("end_date=2024/12/31")) {
			t.Errorf("Expected 'end_date=2024/12/31' in request body, but it was missing")
		}

		// Respond with a mock response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	// Create the request
	formData := "business_type=month&start_date=2024/12/01&end_date=2024/12/31"
	req, err := http.NewRequest("POST", server.URL, bytes.NewBufferString(formData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200 but got %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Println("Response Body:", string(body))
}
