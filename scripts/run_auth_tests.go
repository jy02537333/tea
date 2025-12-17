package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// This program performs an automated dev-login and then runs authenticated probes
// against a list of endpoints. It writes both the raw dev-login JSON and a human
// readable results file in the repository directory.

func main() {
	// allow overriding base URL and output dir for flexibility
	var base string
	var outDir string
	flag.StringVar(&base, "base", "http://localhost:9292/api/v1", "API base URL")
	flag.StringVar(&outDir, "out", ".", "Output directory (will contain devlogin_resp.json and results)")
	flag.Parse()

	// normalize outDir to absolute path
	if !filepath.IsAbs(outDir) {
		wd, _ := os.Getwd()
		outDir = filepath.Join(wd, outDir)
	}
	_ = os.MkdirAll(outDir, 0o755)

	outDev := filepath.Join(outDir, "devlogin_resp.json")
	outResults := filepath.Join(outDir, "admin_api_test_results_auth.txt")

	client := &http.Client{Timeout: 10 * time.Second}

	// wait for health to be available
	if !waitForHealth(client, base+"/health", 20, 1*time.Second) {
		log.Printf("health endpoint not available at %s, continuing but likely failures will follow", base+"/health")
	}

	// attempt dev-login with retries
	token := ""
	devLoginURL := base + "/user/dev-login"
	devBody := map[string]string{"openid": "admin_openid"}
	var devRespBody []byte
	for i := 0; i < 5; i++ {
		b, err := json.Marshal(devBody)
		if err != nil {
			log.Fatalf("marshal dev body: %v", err)
		}
		req, _ := http.NewRequest("POST", devLoginURL, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		req = req.WithContext(ctx)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("dev-login attempt %d failed: %v", i+1, err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		devRespBody, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		// save raw dev response each attempt (last wins)
		_ = os.WriteFile(outDev, devRespBody, 0o644)
		// try to extract token
		token = extractToken(devRespBody)
		if token != "" {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}

	// fallback: allow token from env var for environments where dev-login is disabled
	if token == "" {
		token = os.Getenv("DEVLOGIN_TOKEN")
		if token != "" {
			log.Printf("using token from DEVLOGIN_TOKEN env var")
		}
	}

	// endpoints to probe
	endpoints := []string{
		"/health", "/auth/captcha", "/auth/login", "/auth/me", "/admin/menus", "/admin/users",
		"/products", "/products/1", "/categories", "/categories/1", "/stores",
		"/admin/stores/1/orders", "/admin/stores/1/orders/stats", "/admin/stores/1/products", "/admin/stores/1/products/1",
		"/admin/products", "/admin/orders", "/admin/orders/export", "/admin/orders/1",
		"/orders/1/deliver", "/orders/1/complete", "/orders/1/cancel", "/orders/1/admin-cancel", "/orders/1/refund", "/orders/1/refund/start", "/orders/1/refund/confirm",
		"/admin/logs/operations?module=finance&order_id=1&limit=5", "/admin/logs/operations", "/admin/logs/operations/export", "/admin/logs/access", "/admin/logs/access/export",
		"/admin/refunds", "/admin/refunds/export",
	}

	postPatterns := map[string]bool{
		"/orders/1/deliver":        true,
		"/orders/1/complete":       true,
		"/orders/1/cancel":         true,
		"/orders/1/admin-cancel":   true,
		"/orders/1/refund":         true,
		"/orders/1/refund/start":   true,
		"/orders/1/refund/confirm": true,
	}

	var out strings.Builder
	out.WriteString(fmt.Sprintf("Authenticated API test results - %s\n\n", time.Now().Format(time.RFC3339)))
	out.WriteString("Dev-login raw response: " + outDev + "\n\n")
	if token == "" {
		out.WriteString("WARNING: token not found in dev-login response and DEVLOGIN_TOKEN not set; probes will be unauthenticated.\n\n")
	} else {
		out.WriteString(fmt.Sprintf("Token length: %d\n\n", len(token)))
	}

	for _, p := range endpoints {
		url := base + p
		method := "GET"
		if postPatterns[p] {
			method = "POST"
		}
		out.WriteString("URL: " + url + "\n")
		out.WriteString("METHOD: " + method + "\n")

		var req *http.Request
		if method == "GET" {
			req, _ = http.NewRequest("GET", url, nil)
		} else {
			req, _ = http.NewRequest("POST", url, bytes.NewReader([]byte("{}")))
			req.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		req = req.WithContext(ctx)
		resp, err := client.Do(req)
		cancel()
		if err != nil {
			out.WriteString("STATUS: ERROR\n")
			out.WriteString("BODY_PREVIEW:\n" + err.Error() + "\n")
			out.WriteString("----\n")
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		out.WriteString(fmt.Sprintf("STATUS: %d\n", resp.StatusCode))
		preview := string(body)
		if len(preview) > 2000 {
			preview = preview[:2000] + "..."
		}
		if strings.TrimSpace(preview) == "" {
			preview = "(empty body)"
		}
		out.WriteString("BODY_PREVIEW:\n" + preview + "\n")
		out.WriteString("----\n")
	}

	// write results
	if err := os.WriteFile(outResults, []byte(out.String()), 0o644); err != nil {
		log.Printf("failed writing results: %v", err)
	} else {
		log.Printf("Wrote results to %s", outResults)
	}
}

func extractToken(b []byte) string {
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return ""
	}
	// common shapes
	if data, ok := m["data"].(map[string]interface{}); ok {
		if t, ok := data["token"].(string); ok && t != "" {
			return t
		}
		if t, ok := data["access_token"].(string); ok && t != "" {
			return t
		}
		if t, ok := data["Token"].(string); ok && t != "" {
			return t
		}
	}
	if t, ok := m["token"].(string); ok && t != "" {
		return t
	}
	if t, ok := m["access_token"].(string); ok && t != "" {
		return t
	}
	return ""
}

func waitForHealth(client *http.Client, url string, attempts int, delay time.Duration) bool {
	for i := 0; i < attempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		resp, err := client.Do(req)
		cancel()
		if err == nil && resp != nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return true
			}
		}
		time.Sleep(delay)
	}
	return false
}
