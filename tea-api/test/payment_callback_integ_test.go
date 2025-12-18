package test

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"

    "tea-api/internal/config"
    "tea-api/internal/router"
    "tea-api/pkg/database"
)

// Test_PaymentCallback_Success simulates the minimal success path:
// dev-login -> add to cart -> create order -> unified-order -> callback SUCCESS -> order paid
func Test_PaymentCallback_Success(t *testing.T) {
    if err := config.LoadConfig("../configs/config.yaml"); err != nil {
        t.Fatalf("load config: %v", err)
    }
    _ = os.Setenv("TEA_AUTO_MIGRATE", "1")
    database.InitDatabase()

    r := router.SetupRouter()
    ts := httptest.NewServer(r)
    defer ts.Close()

    loginReq := map[string]string{"openid": "user_openid_callback"}
    b, _ := json.Marshal(loginReq)
    resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
    if err != nil { t.Fatalf("dev-login request err: %v", err) }
    defer resp.Body.Close()
    var login struct{ Code int; Data struct{ Token string } }
    json.NewDecoder(resp.Body).Decode(&login)
    if login.Code != 0 || login.Data.Token == "" { t.Fatalf("login failed: %+v", login) }
    auth := "Bearer " + login.Data.Token

    catReq := map[string]any{"name": "回调测试分类"}
    cb, _ := json.Marshal(catReq)
    req, _ := http.NewRequest("POST", ts.URL+"/api/v1/categories", bytes.NewReader(cb))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    resp2, err := http.DefaultClient.Do(req)
    if err != nil { t.Fatalf("create category err: %v", err) }
    var catResp struct{ Code int; Data struct{ ID uint }; Message string }
    json.NewDecoder(resp2.Body).Decode(&catResp)
    resp2.Body.Close()
    if catResp.Code != 0 || catResp.Data.ID == 0 { t.Fatalf("create category failed: code=%d msg=%s", catResp.Code, catResp.Message) }

    prodReq := map[string]any{"category_id": catResp.Data.ID, "name": "回调测试商品", "price": 9.90, "stock": 3, "status": 1}
    pb, _ := json.Marshal(prodReq)
    req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(pb))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    resp3, err := http.DefaultClient.Do(req)
    if err != nil { t.Fatalf("create product err: %v", err) }
    var prodResp struct{ Code int; Data struct{ ID uint }; Message string }
    json.NewDecoder(resp3.Body).Decode(&prodResp)
    resp3.Body.Close()
    if prodResp.Code != 0 || prodResp.Data.ID == 0 { t.Fatalf("create product failed: code=%d msg=%s", prodResp.Code, prodResp.Message) }

    addReq := map[string]any{"product_id": prodResp.Data.ID, "quantity": 1}
    ab, _ := json.Marshal(addReq)
    req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    respAdd, err := http.DefaultClient.Do(req)
    if err != nil { t.Fatalf("add cart err: %v", err) }
    if respAdd.StatusCode != 200 { t.Fatalf("add cart status: %d", respAdd.StatusCode) }
    respAdd.Body.Close()

    orderReq := map[string]any{"delivery_type": 1}
    ob, _ := json.Marshal(orderReq)
    req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    resp4, err := http.DefaultClient.Do(req)
    if err != nil { t.Fatalf("create order err: %v", err) }
    var orderResp struct{ Code int; Message string; Data struct{ ID uint } }
    json.NewDecoder(resp4.Body).Decode(&orderResp)
    resp4.Body.Close()
    if orderResp.Code != 0 || orderResp.Data.ID == 0 { t.Fatalf("create order failed: code=%d msg=%s", orderResp.Code, orderResp.Message) }

    ureq := map[string]any{"order_id": orderResp.Data.ID, "method": 1}
    ub, _ := json.Marshal(ureq)
    req, _ = http.NewRequest("POST", ts.URL+"/api/v1/payments/unified-order", bytes.NewReader(ub))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    resp5, err := http.DefaultClient.Do(req)
    if err != nil { t.Fatalf("unified-order err: %v", err) }
    var uresp struct{ Code int; Message string; Data struct{ PaymentNo string `json:"payment_no"` } }
    json.NewDecoder(resp5.Body).Decode(&uresp)
    resp5.Body.Close()
    if uresp.Code != 0 || uresp.Data.PaymentNo == "" { t.Fatalf("unified-order failed: code=%d msg=%s", uresp.Code, uresp.Message) }

    cbReq := map[string]any{"payment_no": uresp.Data.PaymentNo, "transaction_id": "txn_test_001", "trade_state": "SUCCESS", "sign": "", "test_mode": true}
    cbb, _ := json.Marshal(cbReq)
    resp6, err := http.Post(ts.URL+"/api/v1/payments/callback", "application/json", bytes.NewReader(cbb))
    if err != nil { t.Fatalf("callback err: %v", err) }
    if resp6.StatusCode != 200 { t.Fatalf("callback status: %d", resp6.StatusCode) }
    resp6.Body.Close()

    req, _ = http.NewRequest("GET", fmt.Sprintf(ts.URL+"/api/v1/orders/%d", orderResp.Data.ID), nil)
    req.Header.Set("Authorization", auth)
    resp7, err := http.DefaultClient.Do(req)
    if err != nil { t.Fatalf("get order detail err: %v", err) }
    var od struct{
        Code int
        Data struct{
            Order struct{
                Status   int `json:"status"`
                PayStatus int `json:"pay_status"`
            } `json:"order"`
        }
    }
    json.NewDecoder(resp7.Body).Decode(&od)
    resp7.Body.Close()
    if od.Code != 0 { t.Fatalf("order detail code=%d", od.Code) }
    if !(od.Data.Order.Status == 2 && od.Data.Order.PayStatus == 2) { t.Fatalf("order not marked paid: status=%d pay_status=%d", od.Data.Order.Status, od.Data.Order.PayStatus) }
}

// Signature failure when not in test_mode
func Test_PaymentCallback_SignatureFailure(t *testing.T) {
    if err := config.LoadConfig("../configs/config.yaml"); err != nil {
        t.Fatalf("load config: %v", err)
    }
    _ = os.Setenv("TEA_AUTO_MIGRATE", "1")
    database.InitDatabase()

    r := router.SetupRouter()
    ts := httptest.NewServer(r)
    defer ts.Close()

    loginReq := map[string]string{"openid": "user_openid_sigfail"}
    b, _ := json.Marshal(loginReq)
    resp, _ := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
    var login struct{ Code int; Data struct{ Token string } }
    json.NewDecoder(resp.Body).Decode(&login)
    resp.Body.Close()
    auth := "Bearer " + login.Data.Token

    catReq := map[string]any{"name": "签名失败测试分类"}
    cb, _ := json.Marshal(catReq)
    req, _ := http.NewRequest("POST", ts.URL+"/api/v1/categories", bytes.NewReader(cb))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    resp2, _ := http.DefaultClient.Do(req)
    var catResp struct{ Code int; Data struct{ ID uint } }
    json.NewDecoder(resp2.Body).Decode(&catResp)
    resp2.Body.Close()

    prodReq := map[string]any{"category_id": catResp.Data.ID, "name": "签名失败测试商品", "price": 1.00, "stock": 1, "status": 1}
    pb, _ := json.Marshal(prodReq)
    req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(pb))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    resp3, _ := http.DefaultClient.Do(req)
    var prodResp struct{ Code int; Data struct{ ID uint } }
    json.NewDecoder(resp3.Body).Decode(&prodResp)
    resp3.Body.Close()

    addReq := map[string]any{"product_id": prodResp.Data.ID, "quantity": 1}
    ab, _ := json.Marshal(addReq)
    req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    _, _ = http.DefaultClient.Do(req)

    orderReq := map[string]any{"delivery_type": 1}
    ob, _ := json.Marshal(orderReq)
    req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    resp4, _ := http.DefaultClient.Do(req)
    var orderResp struct{ Code int; Data struct{ ID uint } }
    json.NewDecoder(resp4.Body).Decode(&orderResp)
    resp4.Body.Close()

    ureq := map[string]any{"order_id": orderResp.Data.ID, "method": 1}
    ub, _ := json.Marshal(ureq)
    req, _ = http.NewRequest("POST", ts.URL+"/api/v1/payments/unified-order", bytes.NewReader(ub))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    resp5, _ := http.DefaultClient.Do(req)
    var uresp struct{ Code int; Message string; Data struct{ PaymentNo string `json:"payment_no"` } }
    ubody, _ := io.ReadAll(resp5.Body)
    resp5.Body.Close()
    _ = json.Unmarshal(ubody, &uresp)
    t.Logf("unified-order body: %s", string(ubody))

    cbReq := map[string]any{"payment_no": uresp.Data.PaymentNo, "transaction_id": "txn_bad_sig", "trade_state": "SUCCESS", "sign": "BAD_SIGNATURE", "test_mode": false}
    cbb, _ := json.Marshal(cbReq)
    resp6, err := http.Post(ts.URL+"/api/v1/payments/callback", "application/json", bytes.NewReader(cbb))
    if err != nil { t.Fatalf("callback request err: %v", err) }
    if resp6.StatusCode == 200 { t.Fatalf("expected signature failure (400), got %d", resp6.StatusCode) }
    resp6.Body.Close()

    req, _ = http.NewRequest("GET", fmt.Sprintf(ts.URL+"/api/v1/orders/%d", orderResp.Data.ID), nil)
    req.Header.Set("Authorization", auth)
    r2, _ := http.DefaultClient.Do(req)
    var od struct{
        Code int
        Data struct{
            Order struct{
                Status   int `json:"status"`
                PayStatus int `json:"pay_status"`
            } `json:"order"`
        }
    }
    json.NewDecoder(r2.Body).Decode(&od)
    r2.Body.Close()
    if od.Data.Order.PayStatus == 2 { t.Fatalf("order should remain unpaid when signature invalid") }
}

// Idempotent duplicate SUCCESS callbacks
func Test_PaymentCallback_IdempotentSuccess(t *testing.T) {
    if err := config.LoadConfig("../configs/config.yaml"); err != nil {
        t.Fatalf("load config: %v", err)
    }
    _ = os.Setenv("TEA_AUTO_MIGRATE", "1")
    database.InitDatabase()

    r := router.SetupRouter()
    ts := httptest.NewServer(r)
    defer ts.Close()

    loginReq := map[string]string{"openid": "user_openid_idem"}
    b, _ := json.Marshal(loginReq)
    resp, _ := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
    var login struct{ Code int; Data struct{ Token string } }
    json.NewDecoder(resp.Body).Decode(&login)
    resp.Body.Close()
    auth := "Bearer " + login.Data.Token

    catReq := map[string]any{"name": "幂等测试分类"}
    cb, _ := json.Marshal(catReq)
    req, _ := http.NewRequest("POST", ts.URL+"/api/v1/categories", bytes.NewReader(cb))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    resp2, _ := http.DefaultClient.Do(req)
    var catResp struct{ Code int; Data struct{ ID uint } }
    json.NewDecoder(resp2.Body).Decode(&catResp)
    resp2.Body.Close()

    prodReq := map[string]any{"category_id": catResp.Data.ID, "name": "幂等测试商品", "price": 2.00, "stock": 2, "status": 1}
    pb, _ := json.Marshal(prodReq)
    req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(pb))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    resp3, _ := http.DefaultClient.Do(req)
    var prodResp struct{ Code int; Data struct{ ID uint } }
    json.NewDecoder(resp3.Body).Decode(&prodResp)
    resp3.Body.Close()

    addReq := map[string]any{"product_id": prodResp.Data.ID, "quantity": 1}
    ab, _ := json.Marshal(addReq)
    req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    _, _ = http.DefaultClient.Do(req)

    orderReq := map[string]any{"delivery_type": 1}
    ob, _ := json.Marshal(orderReq)
    req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    resp4, _ := http.DefaultClient.Do(req)
    var orderResp struct{ Code int; Data struct{ ID uint } }
    json.NewDecoder(resp4.Body).Decode(&orderResp)
    resp4.Body.Close()

    ureq := map[string]any{"order_id": orderResp.Data.ID, "method": 1}
    ub, _ := json.Marshal(ureq)
    req, _ = http.NewRequest("POST", ts.URL+"/api/v1/payments/unified-order", bytes.NewReader(ub))
    req.Header.Set("Authorization", auth)
    req.Header.Set("Content-Type", "application/json")
    resp5, _ := http.DefaultClient.Do(req)
    var uresp struct{ Code int; Data struct{ PaymentNo string `json:"payment_no"` } ; Message string }
    json.NewDecoder(resp5.Body).Decode(&uresp)
    resp5.Body.Close()
    if uresp.Code != 0 || uresp.Data.PaymentNo == "" { t.Fatalf("unified-order failed: code=%d msg=%s", uresp.Code, uresp.Message) }

    cbReq := map[string]any{"payment_no": uresp.Data.PaymentNo, "transaction_id": "txn_idem_1", "trade_state": "SUCCESS", "sign": "", "test_mode": true}
    cbb, _ := json.Marshal(cbReq)
    resp6, _ := http.Post(ts.URL+"/api/v1/payments/callback", "application/json", bytes.NewReader(cbb))
    if resp6.StatusCode != 200 {
        var rb bytes.Buffer
        io.Copy(&rb, resp6.Body)
        resp6.Body.Close()
        t.Fatalf("first callback expected 200, got %d; body=%s", resp6.StatusCode, rb.String())
    }
    resp6.Body.Close()

    resp7, _ := http.Post(ts.URL+"/api/v1/payments/callback", "application/json", bytes.NewReader(cbb))
    if resp7.StatusCode != 200 { t.Fatalf("duplicate callback expected 200, got %d", resp7.StatusCode) }
    resp7.Body.Close()

    req, _ = http.NewRequest("GET", fmt.Sprintf(ts.URL+"/api/v1/orders/%d", orderResp.Data.ID), nil)
    req.Header.Set("Authorization", auth)
    r2, _ := http.DefaultClient.Do(req)
    var od struct{
        Code int
        Data struct{
            Order struct{
                Status   int `json:"status"`
                PayStatus int `json:"pay_status"`
            } `json:"order"`
        }
    }
    bodyBytes, _ := io.ReadAll(r2.Body)
    r2.Body.Close()
    _ = json.Unmarshal(bodyBytes, &od)
    t.Logf("order detail body: %s", string(bodyBytes))
    if !(od.Data.Order.Status == 2 && od.Data.Order.PayStatus == 2) { t.Fatalf("order not paid after idempotent callbacks: status=%d pay_status=%d", od.Data.Order.Status, od.Data.Order.PayStatus) }
}
