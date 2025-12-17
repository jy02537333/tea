package teaapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCartRequiresAuth(t *testing.T) {
	baseURL := "http://localhost:9292/api/v1"
	resp, err := http.Get(baseURL + "/cart")
	if err != nil {
		t.Fatalf("cart request failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for cart, got %d", resp.StatusCode)
	}
}

func TestOrdersRequiresAuth(t *testing.T) {
	baseURL := "http://localhost:9292/api/v1"
	resp, err := http.Get(baseURL + "/orders")
	if err != nil {
		t.Fatalf("orders request failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for orders, got %d", resp.StatusCode)
	}
}

func TestLoginThenCartOK(t *testing.T) {
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
	var jsonResp struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		t.Fatalf("decode login json failed: %v", err)
	}
	if jsonResp.Data.Token == "" {
		t.Fatalf("token empty")
	}
	req, _ := http.NewRequest("GET", baseURL+"/cart", nil)
	req.Header.Set("Authorization", "Bearer "+jsonResp.Data.Token)
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("cart request failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK && resp2.StatusCode != http.StatusNoContent {
		t.Fatalf("cart expected 200/204, got %d", resp2.StatusCode)
	}
}
