package teaapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// Expect 403 for routes requiring specific permissions when using a normal user token.
func TestStoreAccountsForbidden(t *testing.T) {
	baseURL := "http://localhost:9292/api/v1"
	// login to get a normal user token
	body := []byte(`{"phone":"13800000000","code":"1234"}`)
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login expected 200, got %d", resp.StatusCode)
	}
	var login struct {
		Code int
		Data struct{ Token string }
	}
	if err := json.NewDecoder(resp.Body).Decode(&login); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if login.Data.Token == "" {
		t.Fatalf("empty token")
	}

	req, _ := http.NewRequest("GET", baseURL+"/stores/1/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+login.Data.Token)
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp2.StatusCode)
	}
}

func TestStoreWithdrawApplyForbidden(t *testing.T) {
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
	var login struct {
		Code int
		Data struct{ Token string }
	}
	if err := json.NewDecoder(resp.Body).Decode(&login); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if login.Data.Token == "" {
		t.Fatalf("empty token")
	}

	req, _ := http.NewRequest("POST", baseURL+"/stores/1/withdraws", bytes.NewReader([]byte(`{"amount":10}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+login.Data.Token)
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp2.StatusCode)
	}
}

func TestStoreCouponsCreateForbidden(t *testing.T) {
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
	var login struct {
		Code int
		Data struct{ Token string }
	}
	if err := json.NewDecoder(resp.Body).Decode(&login); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if login.Data.Token == "" {
		t.Fatalf("empty token")
	}

	req, _ := http.NewRequest("POST", baseURL+"/stores/1/coupons", bytes.NewReader([]byte(`{"name":"测试券"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+login.Data.Token)
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp2.StatusCode)
	}
}
