package handler

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"

    "tea-api/internal/model"
    "tea-api/pkg/database"
)

// TestAdminRefund_TriggersCommissionRollback 验证 AdminRefund 成功执行后，会按订单回滚未提现佣金（frozen/available -> reversed）
func TestAdminRefund_TriggersCommissionRollback(t *testing.T) {
    gin.SetMode(gin.TestMode)

    // 初始化内存 SQLite，并迁移必要模型
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("open sqlite: %v", err)
    }
    database.DB = db
    if err := db.AutoMigrate(
        &model.Order{},
        &model.OrderItem{},
        &model.Commission{},
        &model.CommissionTransaction{},
        &model.UserCoupon{},
    ); err != nil {
        t.Fatalf("auto migrate: %v", err)
    }

    // 构造状态=3(处理中)，支付=2(已付款) 的订单，满足 AdminRefund 条件；避免走库存回补分支
    ord := &model.Order{
        UserID:    1001,
        StoreID:   0,
        OrderNo:   "UT_REFUND_ROLLBACK_001",
        Status:    3,
        PayStatus: 2,
    }
    if err := db.Create(ord).Error; err != nil {
        t.Fatalf("create order: %v", err)
    }

    // 为该订单写入两条可回滚佣金（frozen/available）与一条已提现佣金（paid，不应回滚）
    availableAt := time.Now()
    frozen := &model.Commission{UserID: ord.UserID, OrderID: &ord.ID, CommissionType: "direct", Status: "frozen", NetAmount: decimal.NewFromInt(100), GrossAmount: decimal.NewFromInt(100), CalculationBasis: decimal.NewFromInt(1000)}
    available := &model.Commission{UserID: ord.UserID, OrderID: &ord.ID, CommissionType: "direct", Status: "available", NetAmount: decimal.NewFromInt(200), GrossAmount: decimal.NewFromInt(200), CalculationBasis: decimal.NewFromInt(2000), AvailableAt: &availableAt}
    paid := &model.Commission{UserID: ord.UserID, OrderID: &ord.ID, CommissionType: "direct", Status: "paid", NetAmount: decimal.NewFromInt(300), GrossAmount: decimal.NewFromInt(300), CalculationBasis: decimal.NewFromInt(3000)}
    if err := db.Create(frozen).Error; err != nil { t.Fatalf("create frozen: %v", err) }
    if err := db.Create(available).Error; err != nil { t.Fatalf("create available: %v", err) }
    if err := db.Create(paid).Error; err != nil { t.Fatalf("create paid: %v", err) }

    // 构造路由并注入 user_id，中间件模拟鉴权上下文
    r := gin.New()
    r.Use(func(c *gin.Context) { c.Set("user_id", uint(999)) })
    h := NewOrderHandler()
    r.POST("/api/v1/admin/orders/:id/refund", h.AdminRefund)

    // 触发 AdminRefund
    req := httptest.NewRequest(http.MethodPost,
        "/api/v1/admin/orders/"+jsonID(ord.ID)+"/refund",
        http.NoBody,
    )
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK {
        t.Fatalf("admin refund failed: status=%d body=%s", w.Code, w.Body.String())
    }
    // 响应必须是合法 JSON
    var body any
    if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
        t.Fatalf("response not JSON: %v", err)
    }

    // 校验佣金状态：frozen/available 被置为 reversed；paid 保持不变
    var rf, ra, rp model.Commission
    if err := db.First(&rf, frozen.ID).Error; err != nil { t.Fatalf("query frozen: %v", err) }
    if err := db.First(&ra, available.ID).Error; err != nil { t.Fatalf("query available: %v", err) }
    if err := db.First(&rp, paid.ID).Error; err != nil { t.Fatalf("query paid: %v", err) }
    if rf.Status != "reversed" || ra.Status != "reversed" {
        t.Fatalf("expected frozen/available reversed, got %s/%s", rf.Status, ra.Status)
    }
    if rp.Status != "paid" {
        t.Fatalf("paid commission should remain paid, got %s", rp.Status)
    }

    // 确认仅针对两条记录生成 adjust 流水
    var cnt int64
    if err := db.Model(&model.CommissionTransaction{}).
        Where("commission_id IN ? AND type = ?", []uint{frozen.ID, available.ID}, "adjust").
        Count(&cnt).Error; err != nil {
        t.Fatalf("count adjust: %v", err)
    }
    if cnt != 2 {
        t.Fatalf("expected 2 adjust transactions, got %d", cnt)
    }
    if err := db.Model(&model.CommissionTransaction{}).
        Where("commission_id = ? AND type = ?", paid.ID, "adjust").
        Count(&cnt).Error; err != nil {
        t.Fatalf("count paid adjust: %v", err)
    }
    if cnt != 0 {
        t.Fatalf("expected 0 adjust for paid, got %d", cnt)
    }
}

// jsonID 帮助函数：将 uint 转为字符串
func jsonID(id uint) string { return fmt.Sprintf("%d", id) }
