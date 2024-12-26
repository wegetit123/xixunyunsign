package cmd_test

import (
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
