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
		Rows int `json:"rows"`
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
	body := []byte(`{"csv":"Stage,Value,Source\nClosed Won,100,Referral\nClosed Lost,25,Web\nOpen,50,Partner\n","status_column":"Stage","value_column":"Value","source_column":"Source","won_statuses":["Closed Won"],"lost_statuses":["Closed Lost"]}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/report", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected report response: %d %s", response.Code, response.Body.String())
	}

	var result struct {
		ReportTitle string `json:"report_title"`
		TotalLeads int `json:"total_leads"`
		ClosedLeads int `json:"closed_leads"`
		OpenLeads int `json:"open_leads"`
		WinRatePercent float64 `json:"win_rate_percent"`
		OpenPipelineValue float64 `json:"open_pipeline_value"`
		MissingStatuses int `json:"missing_statuses"`
		MissingValues int `json:"missing_values"`
		MissingSources int `json:"missing_sources"`
		Observations []string `json:"observations"`
		RecommendedActions []string `json:"recommended_actions"`
		Limitations []string `json:"limitations"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.ReportTitle == "" || result.TotalLeads != 3 || result.ClosedLeads != 2 || result.OpenLeads != 1 {
		t.Fatalf("unexpected report counts/title: %+v", result)
	}
	if result.WinRatePercent != 50 || result.OpenPipelineValue != 50 {
		t.Fatalf("unexpected report metrics: %+v", result)
	}
	if result.MissingStatuses != 0 || result.MissingValues != 0 || result.MissingSources != 0 {
		t.Fatalf("unexpected data-quality counts: %+v", result)
	}
	if len(result.Observations) == 0 || len(result.RecommendedActions) == 0 || len(result.Limitations) == 0 {
		t.Fatalf("report is missing explanatory sections: %+v", result)
	}
}

func TestReportRequiresConfiguredBearerToken(t *testing.T) {
	server := NewWithToken(nil, "demo-secret")
	body := `{"csv":"Stage,Value\nOpen,50\n","status_column":"Stage","value_column":"Value","source_column":"","won_statuses":["Won"],"lost_statuses":["Lost"]}`

	request := httptest.NewRequest(http.MethodPost, "/api/report", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected report request without token to be unauthorized, got %d: %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/api/report", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer demo-secret")
	response = httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected authenticated report request to succeed, got %d: %s", response.Code, response.Body.String())
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

func TestCORSPreflightAllowsBearerAuthorization(t *testing.T) {
	server := NewWithToken([]string{"https://matteojose3112-bit.github.io"}, "demo-secret")
	request := httptest.NewRequest(http.MethodOptions, "/api/inspect", nil)
	request.Header.Set("Origin", "https://matteojose3112-bit.github.io")
	request.Header.Set("Access-Control-Request-Method", "POST")
	request.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected preflight success, got %d: %s", response.Code, response.Body.String())
	}
	allowed := strings.ToLower(response.Header().Get("Access-Control-Allow-Headers"))
	if !strings.Contains(allowed, "authorization") || !strings.Contains(allowed, "content-type") {
		t.Fatalf("preflight did not allow required headers: %q", allowed)
	}
}

func TestAPIRequiresConfiguredBearerToken(t *testing.T) {
	server := NewWithToken(nil, "demo-secret")
	body := `{"csv":"A\\nx"}`
	request := httptest.NewRequest(http.MethodPost, "/api/inspect", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized without token, got %d", response.Code)
	}
	request = httptest.NewRequest(http.MethodPost, "/api/inspect", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer demo-secret")
	response = httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected authorized request, got %d: %s", response.Code, response.Body.String())
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
