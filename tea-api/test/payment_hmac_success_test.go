package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"tea-api/internal/config"
	"tea-api/pkg/utils"
)

// 严格 HMAC 成功路径：从配置读取 APIKey，对原始回调报文做 HMAC-SHA256
func Test_PaymentCallback_HMAC_Strict_Success(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// 前置：登录 -> 创建分类/商品 -> 加入购物车 -> 创建订单 -> 统一下单（仅为拿到 payment_no）
	token := devLogin(t, ts, fmt.Sprintf("hmac_user_%d", time.Now().UnixNano()))
	authHeader := "Bearer " + token

	catID := createCategory(t, ts, authHeader)
	productID := createProduct(t, ts, authHeader, catID)
	addCartItem(t, ts, authHeader, productID)
	orderID := createOrderFromCart(t, ts, authHeader)

	uo := callUnifiedOrder(t, ts, authHeader, orderID)
	if uo.Data.PaymentNo == "" {
		t.Fatalf("payment_no empty")
	}

	// 构造不含 sign 的原始回调体并计算 HMAC（按服务端规范：移除 sign，按 key 排序，紧凑 JSON）
	callbackBody := map[string]any{
		"app_id":         "tea-app-mock",
		"paid_at":        time.Now().Format(time.RFC3339),
		"payment_no":     uo.Data.PaymentNo,
		"trade_state":    "SUCCESS",
		"transaction_id": fmt.Sprintf("mock_%s", uo.Data.PaymentNo),
	}
	// 规范化编码
	keys := make([]string, 0, len(callbackBody))
	for k := range callbackBody {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var bldr strings.Builder
	bldr.WriteString("{")
	for i, k := range keys {
		v := callbackBody[k]
		bz, _ := json.Marshal(v)
		bldr.WriteString("\"")
		bldr.WriteString(k)
		bldr.WriteString("\":")
		bldr.Write(bz)
		if i < len(keys)-1 {
			bldr.WriteString(",")
		}
	}
	bldr.WriteString("}")
	canonical := bldr.String()

	secret := config.Config.WeChat.APIKey
	if secret == "" {
		t.Fatalf("wechat api_key empty in config")
	}
	sign := utils.HMACSHA256Hex(secret, canonical)

	// 将 sign 注入原始体（不改变已有字段顺序，只附加 sign）
	callbackBody["sign"] = sign
	body, _ := json.Marshal(callbackBody)

	resp, err := http.Post(ts.URL+"/api/v1/payments/callback", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("callback request err: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var respBody bytes.Buffer
		_, _ = respBody.ReadFrom(resp.Body)
		t.Fatalf("callback status: %d body: %s", resp.StatusCode, respBody.String())
	}
}
