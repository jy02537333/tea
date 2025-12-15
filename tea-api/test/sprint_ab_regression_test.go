package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tea-api/internal/config"
	"tea-api/internal/router"
	"tea-api/pkg/database"
)

// Test_SprintAB_Regression_Cart_Order_Coupon 测试 Sprint A/B 关键接口的稳定性
// 测试范围：购物车、下单、可用券列表
func Test_SprintAB_Regression_Cart_Order_Coupon(t *testing.T) {
	// 加载配置并初始化数据库
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	database.InitDatabase()

	// 设置路由
	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1. 登录获取 token
	t.Run("登录获取token", func(t *testing.T) {
		loginReq := map[string]string{"openid": "sprint_ab_regression_user"}
		b, _ := json.Marshal(loginReq)
		resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatalf("登录请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Fatalf("登录状态码错误: %d", resp.StatusCode)
		}

		var login struct {
			Code int `json:"code"`
			Data struct {
				Token string `json:"token"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&login); err != nil {
			t.Fatalf("解析登录响应失败: %v", err)
		}
		if login.Code != 0 || login.Data.Token == "" {
			t.Fatalf("登录失败: code=%d, token=%s", login.Code, login.Data.Token)
		}
		t.Logf("✓ 登录成功，获得 token: %s...", login.Data.Token[:20])
	})

	// 登录以便后续测试使用
	loginReq := map[string]string{"openid": "sprint_ab_regression_user"}
	b, _ := json.Marshal(loginReq)
	loginResp, _ := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
	var login struct {
		Code int
		Data struct {
			Token string
			User  struct{ ID uint }
		}
	}
	json.NewDecoder(loginResp.Body).Decode(&login)
	loginResp.Body.Close()
	auth := "Bearer " + login.Data.Token

	// 2. 测试购物车接口
	t.Run("购物车_获取购物车", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/v1/cart", nil)
		req.Header.Set("Authorization", auth)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("获取购物车请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("获取购物车状态码错误: %d", resp.StatusCode)
		} else {
			t.Logf("✓ 获取购物车成功")
		}
	})

	// 3. 测试商品列表（下单前置）
	t.Run("下单_获取商品列表", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/v1/products?page=1&size=10", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("获取商品列表请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("获取商品列表状态码错误: %d", resp.StatusCode)
		} else {
			var result struct {
				Code int
				Data []map[string]interface{}
			}
			json.NewDecoder(resp.Body).Decode(&result)
			t.Logf("✓ 获取商品列表成功，商品数: %d", len(result.Data))
		}
	})

	// 4. 测试订单列表
	t.Run("下单_获取订单列表", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/v1/orders", nil)
		req.Header.Set("Authorization", auth)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("获取订单列表请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("获取订单列表状态码错误: %d", resp.StatusCode)
		} else {
			var result struct {
				Code int
				Data []map[string]interface{}
			}
			json.NewDecoder(resp.Body).Decode(&result)
			t.Logf("✓ 获取订单列表成功，订单数: %d", len(result.Data))
		}
	})

	// 5. 测试优惠券列表
	t.Run("优惠券_获取可用券列表", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/v1/coupons", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("获取优惠券列表请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("获取优惠券列表状态码错误: %d", resp.StatusCode)
		} else {
			var result struct {
				Code int
				Data []map[string]interface{}
			}
			json.NewDecoder(resp.Body).Decode(&result)
			t.Logf("✓ 获取优惠券列表成功，优惠券数: %d", len(result.Data))
		}
	})

	// 6. 测试用户优惠券
	t.Run("优惠券_获取用户优惠券", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/v1/user/coupons", nil)
		req.Header.Set("Authorization", auth)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("获取用户优惠券请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("获取用户优惠券状态码错误: %d", resp.StatusCode)
		} else {
			var result struct {
				Code int
				Data []map[string]interface{}
			}
			json.NewDecoder(resp.Body).Decode(&result)
			t.Logf("✓ 获取用户优惠券成功，券数: %d", len(result.Data))
		}
	})

	// 7. 测试分类列表（购物辅助）
	t.Run("基础_获取分类列表", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/v1/categories", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("获取分类列表请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("获取分类列表状态码错误: %d", resp.StatusCode)
		} else {
			t.Logf("✓ 获取分类列表成功")
		}
	})

	// 8. 测试门店列表（门店下单辅助）
	t.Run("基础_获取门店列表", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/v1/stores", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("获取门店列表请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("获取门店列表状态码错误: %d", resp.StatusCode)
		} else {
			t.Logf("✓ 获取门店列表成功")
		}
	})

	// 9. 测试用户信息（支付前置）
	t.Run("用户_获取用户信息", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/v1/user/info", nil)
		req.Header.Set("Authorization", auth)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("获取用户信息请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("获取用户信息状态码错误: %d", resp.StatusCode)
		} else {
			t.Logf("✓ 获取用户信息成功")
		}
	})

	// 10. 测试健康检查
	t.Run("健康_健康检查", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/v1/health", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("健康检查请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("健康检查状态码错误: %d", resp.StatusCode)
		} else {
			t.Logf("✓ 健康检查通过")
		}
	})
}
