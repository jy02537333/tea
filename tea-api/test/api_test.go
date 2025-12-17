package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func Example_manual_run() {
	baseURL := "http://localhost:9292/api/v1"

	// 测试健康检查
	fmt.Println("=== 测试健康检查 ===")
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应体: %s\n\n", string(body))

	// 测试用户登录
	fmt.Println("=== 测试用户登录 ===")
	loginData := map[string]interface{}{
		"code": "test_code_123",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err = http.Post(baseURL+"/user/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("登录请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应体: %s\n\n", string(body))

	// 解析token
	var loginResp map[string]interface{}
	json.Unmarshal(body, &loginResp)

	if data, ok := loginResp["data"].(map[string]interface{}); ok {
		if token, ok := data["token"].(string); ok {
			// 测试获取用户信息
			fmt.Println("=== 测试获取用户信息 ===")
			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			req, _ := http.NewRequest("GET", baseURL+"/user/info", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err = client.Do(req)
			if err != nil {
				fmt.Printf("获取用户信息请求失败: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, _ = ioutil.ReadAll(resp.Body)
			fmt.Printf("状态码: %d\n", resp.StatusCode)
			fmt.Printf("响应体: %s\n", string(body))
		}
	}
}
