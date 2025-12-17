package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type respWrapper struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

type product struct {
	ID int64 `json:"id"`
}

type orderCreateData struct {
	ID int64 `json:"id"`
}

type unifiedOrderData struct {
	PaymentNo string `json:"payment_no"`
	PrepayID  string `json:"prepay_id"`
	NonceStr  string `json:"nonce_str"`
	Timestamp int64  `json:"timestamp"`
	Sign      string `json:"sign"`
}

func main() {
	baseURLFlag := flag.String("base-url", "", "Base API URL, e.g. http://localhost:9292/api/v1")
	tokenFlag := flag.String("token", "", "Bearer token; falls back to TOKEN env or dev-login")
	productIDFlag := flag.Int64("product-id", 0, "Use a fixed product ID instead of fetching /products")
	quantityFlag := flag.Int("quantity", 1, "Quantity to add to cart")
	devOpenIDFlag := flag.String("dev-openid", "admin_openid", "OpenID used when auto-calling /user/dev-login")
	couponFlag := flag.Int64("coupon-id", 0, "User coupon ID to consume when creating the order")
	flag.Parse()

	baseURL := firstNonEmpty(*baseURLFlag, os.Getenv("BASE_URL"))
	if baseURL == "" {
		baseURL = "http://localhost:9292/api/v1"
	}

	client := &http.Client{}

	// 优先使用 flag/环境变量中的 TOKEN；若为空，则尝试通过 dev-login 自动获取
	token := firstNonEmpty(*tokenFlag, os.Getenv("TOKEN"))
	if token == "" {
		fmt.Println("TOKEN 为空，尝试调用 /user/dev-login 获取开发环境 admin token...")
		adminToken, err := fetchDevAdminToken(client, baseURL, *devOpenIDFlag)
		if err != nil {
			fmt.Println("获取 admin token 失败:", err)
			os.Exit(1)
		}
		fmt.Println("成功获取 admin token")
		token = adminToken
	}

	authHeader := "Bearer " + token

	var productID int64
	if *productIDFlag > 0 {
		productID = *productIDFlag
		fmt.Println("use product_id (flag):", productID)
	} else {
		// 1) 获取商品列表
		fmt.Println("1) GET /products")
		prodResp := doReq(client, "GET", baseURL+"/products?page=1&limit=1", authHeader, nil)
		fmt.Println("products raw data:", string(prodResp.Data))

		var products []product
		if err := json.Unmarshal(prodResp.Data, &products); err != nil {
			// 如果服务端包了一层 {data:[], page,...}，这里可以根据实际结构调整
			fmt.Println("warn: failed to unmarshal products as []product, please adjust parsing according to actual response structure")
		}

		if len(products) > 0 {
			productID = products[0].ID
		}
		if productID == 0 {
			fmt.Println("no product_id found from response; please check response structure or pass -product-id")
			return
		}
		fmt.Println("use product_id:", productID)
	}

	quantity := *quantityFlag
	if quantity <= 0 {
		fmt.Println("quantity <= 0, fallback to 1")
		quantity = 1
	}

	// 2) 加入购物车
	fmt.Println("2) POST /cart")
	bodyAdd := map[string]any{
		"product_id": productID,
		"quantity":   quantity,
	}
	doReq(client, "POST", baseURL+"/cart", authHeader, bodyAdd)

	// 3) 查看购物车
	fmt.Println("3) GET /cart")
	cartResp := doReq(client, "GET", baseURL+"/cart", authHeader, nil)
	fmt.Println("cart raw data:", string(cartResp.Data))

	// 4) 从购物车创建订单
	fmt.Println("4) POST /orders/from-cart")
	bodyOrder := map[string]any{
		"delivery_type":  1,
		"address_info":   "{}",
		"remark":         "测试下单",
		"user_coupon_id": *couponFlag,
		"store_id":       0,
		"order_type":     1,
	}
	if *couponFlag > 0 {
		fmt.Println("apply user_coupon_id:", *couponFlag)
	}
	orderResp := doReq(client, "POST", baseURL+"/orders/from-cart", authHeader, bodyOrder)
	fmt.Println("order create raw data:", string(orderResp.Data))

	var od orderCreateData
	if err := json.Unmarshal(orderResp.Data, &od); err != nil {
		fmt.Println("warn: failed to unmarshal order create data; please check response structure")
	}
	if od.ID == 0 {
		fmt.Println("no order_id found from response; please check response structure")
		return
	}
	orderID := od.ID
	fmt.Println("created order id:", orderID)

	// 5) 模拟统一下单
	fmt.Println("5) POST /payments/unified-order")
	orderReq := map[string]any{"order_id": orderID}
	payResp := doReq(client, "POST", baseURL+"/payments/unified-order", authHeader, orderReq)
	var payData unifiedOrderData
	if err := json.Unmarshal(payResp.Data, &payData); err != nil {
		fmt.Println("warn: failed to unmarshal unified order response; raw=", string(payResp.Data))
	}
	if payData.PaymentNo == "" {
		fmt.Println("unified order did not return payment_no")
		return
	}
	fmt.Println("payment_no:", payData.PaymentNo)

	// 6) 模拟第三方支付回调（test_mode 仅用于本地/CI）
	fmt.Println("6) POST /payments/callback (test_mode)")
	callbackBody := map[string]any{
		"app_id":         "tea-app-mock",
		"payment_no":     payData.PaymentNo,
		"transaction_id": fmt.Sprintf("mock_%s", payData.PaymentNo),
		"trade_state":    "SUCCESS",
		"paid_at":        time.Now().Format(time.RFC3339),
		"sign":           payData.Sign,
		"test_mode":      true,
	}
	doReq(client, "POST", baseURL+"/payments/callback", "", callbackBody)

	// 7) 查看订单详情
	fmt.Println("7) GET /orders/:id")
	detailURL := fmt.Sprintf("%s/orders/%d", baseURL, orderID)
	detailResp := doReq(client, "GET", detailURL, authHeader, nil)
	fmt.Println("order detail raw data:", string(detailResp.Data))

	fmt.Println("E2E single SKU order flow finished.")
}

// 在 local/dev 环境下通过 /user/dev-login 获取 admin_openid 的 token
func fetchDevAdminToken(client *http.Client, baseURL, openid string) (string, error) {
	url := baseURL + "/user/dev-login"
	body := map[string]any{"openid": openid}
	wrap := doReq(client, "POST", url, "", body)
	// LoginResponse 结构中的 data: {"token": "...", "userInfo": {...}}
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(wrap.Data, &payload); err != nil {
		return "", fmt.Errorf("unmarshal dev-login response failed: %w", err)
	}
	if payload.Token == "" {
		return "", fmt.Errorf("dev-login 返回中未找到 token")
	}
	return payload.Token, nil
}

func doReq(client *http.Client, method, url, auth string, body any) respWrapper {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		panic(err)
	}

	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	if res.StatusCode >= 400 {
		fmt.Printf("HTTP %d: %s\n", res.StatusCode, string(b))
		os.Exit(1)
	}

	var wrap respWrapper
	if err := json.Unmarshal(b, &wrap); err != nil {
		fmt.Println("warn: failed to unmarshal wrapper, raw:", string(b))
		return respWrapper{Code: 0, Data: b, Message: ""}
	}

	if wrap.Code != 0 {
		fmt.Printf("biz code=%d msg=%s\n", wrap.Code, wrap.Message)
	}

	return wrap
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
