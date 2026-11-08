package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	server := New(nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected health response: %d %s", response.Code, response.Body.String())
	}
}

func TestInspectExecutesCSVInspector(t *testing.T) {
	server := New(nil)
	body := []byte(`{"csv":"Stage,Value\nClosed Won,100\nOpen,50\n"}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/inspect", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
	var result struct {
		Rows    int `json:"rows"`
		Columns []struct { Name string `json:"name"` } `json:"columns"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Rows != 2 || len(result.Columns) != 2 || result.Columns[0].Name != "Stage" {
		t.Fatalf("unexpected inspection result: %+v", result)
	}
}

func TestReportExecutesSalesOpsTool(t *testing.T) {
	server := New(nil)
	body := []byte(`{"csv":"Stage,Value,Source\nClosed Won,100,Referral\nOpen,50,Web\n","status_column":"Stage","value_column":"Value","source_column":"Source","won_statuses":["Closed Won"],"lost_statuses":["Closed Lost"]}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/report", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"open_pipeline_value":50`) {
		t.Fatalf("unexpected report response: %d %s", response.Code, response.Body.String())
	}
}

func TestRejectsUnknownFieldsAndUnapprovedOrigins(t *testing.T) {
	server := New([]string{"https://matteojose3112-bit.github.io"})
	request := httptest.NewRequest(http.MethodPost, "/api/inspect", strings.NewReader(`{"csv":"A\nx","unexpected":true}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://evil.example")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected origin rejection, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/inspect", strings.NewReader(`{"csv":"A\nx","unexpected":true}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected strict JSON rejection, got %d", response.Code)
	}
}

func TestRejectsMalformedCSV(t *testing.T) {
	server := New(nil)
	request := httptest.NewRequest(http.MethodPost, "/api/inspect", strings.NewReader(`{"csv":"A,B\n1"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected invalid CSV response, got %d: %s", response.Code, response.Body.String())
	}
}
