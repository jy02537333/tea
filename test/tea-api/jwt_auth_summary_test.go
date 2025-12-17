package teaapi_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestSummaryRequiresAuth(t *testing.T) {
	baseURL := "http://localhost:9292/api/v1"
	req, _ := http.NewRequest("GET", baseURL+"/users/me/summary", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLoginThenSummaryOK(t *testing.T) {
	baseURL := "http://localhost:9292/api/v1"
	body := []byte(`{"phone":"13800000000","code":"1234"}`)
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login expected 200, got %d", resp.StatusCode)
	}
	raw, _ := ioutil.ReadAll(resp.Body)
	var jsonResp struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &jsonResp); err != nil {
		t.Fatalf("decode login json failed: %v", err)
	}
	if jsonResp.Data.Token == "" {
		t.Fatalf("token empty")
	}
	req, _ := http.NewRequest("GET", baseURL+"/users/me/summary", nil)
	req.Header.Set("Authorization", "Bearer "+jsonResp.Data.Token)
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("summary request failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("summary expected 200, got %d", resp2.StatusCode)
	}
}
