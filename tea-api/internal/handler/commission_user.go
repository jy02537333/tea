package handler

import (
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"

    svc "tea-api/internal/service/commission"
    "tea-api/internal/model"
    "tea-api/pkg/database"
    "tea-api/pkg/utils"
)

// SC4: 用户侧佣金相关接口（预览计算、创建落库、列表、汇总）

type CommissionUserHandler struct{}

func NewCommissionUserHandler() *CommissionUserHandler { return &CommissionUserHandler{} }

// 计算预览请求
type commissionCalcItem struct {
    SKUId          int64 `json:"sku_id"`
    UnitPriceCents int64 `json:"unit_price_cents"`
    Quantity       int   `json:"quantity"`
    DiscountCents  int64 `json:"discount_cents"`
}

type commissionCalcReq struct {
    PayerUserID             int64                `json:"payer_user_id" binding:"required"`
    ReferrerUserID          int64                `json:"referrer_user_id" binding:"required"`
    Items                   []commissionCalcItem `json:"items" binding:"required"`
    ShippingCents           int64                `json:"shipping_cents"`
    CouponCents             int64                `json:"coupon_cents"`
    OrderLevelDiscountCents int64                `json:"order_level_discount_cents"`
    DirectRate              *float64             `json:"direct_rate"`      // 可选，默认 0.30
    IndirectRate            *float64             `json:"indirect_rate"`    // 可选，默认 0.10
    HoldPeriodDays          *int                 `json:"hold_period_days"` // 可选，默认 7
}

type commissionCalcResp struct {
    Records      []svc.CommissionRecord `json:"records"`
    BasisCents   int64                  `json:"calculation_basis_cents"`
    DirectRate   float64                `json:"direct_rate"`
    IndirectRate float64                `json:"indirect_rate"`
}

// POST /api/v1/commissions/calculate
func (h *CommissionUserHandler) Calculate(c *gin.Context) {
    var req commissionCalcReq
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": utils.CodeInvalidParam, "message": "参数错误", "data": nil})
        return
    }

    direct := 0.30
    if req.DirectRate != nil {
        direct = *req.DirectRate
    }
    indirect := 0.10
    if req.IndirectRate != nil {
        indirect = *req.IndirectRate
    }
    hold := 7
    if req.HoldPeriodDays != nil && *req.HoldPeriodDays > 0 {
        hold = *req.HoldPeriodDays
    }

    // 计算订单总额（分）：按 items 单价*数量 相加；折扣统一累加到订单折扣
    var total int64
    var itemDiscountSum int64
    for _, it := range req.Items {
        total += it.UnitPriceCents * int64(it.Quantity)
        if it.DiscountCents > 0 {
            itemDiscountSum += it.DiscountCents
        }
    }

    order := svc.Order{
        ID:             0,
        UserID:         req.PayerUserID,
        TotalAmount:    total,
        ShippingAmount: req.ShippingCents,
        CouponAmount:   req.CouponCents,
        DiscountAmount: req.OrderLevelDiscountCents + itemDiscountSum,
        Items:          nil,
    }

    recs := svc.BuildCommissionRecords(order, req.ReferrerUserID, direct, indirect, hold)

    // 额外返回基数，便于前端解释
    basis := order.TotalAmount - order.ShippingAmount - order.CouponAmount - order.DiscountAmount
    if basis < 0 {
        basis = 0
    }

    c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": commissionCalcResp{
        Records: recs, BasisCents: basis, DirectRate: direct, IndirectRate: indirect,
    }})
}

// 创建落库请求（两种形态，择一提供）：
// 1) 与 Calculate 同构的订单参数，后端计算并保存；
// 2) 直接提供 records 列表（分为单位），按给定值保存。
type commissionCreateReq struct {
    // 形态1
    PayerUserID             *int64               `json:"payer_user_id"`
    ReferrerUserID          *int64               `json:"referrer_user_id"`
    Items                   []commissionCalcItem `json:"items"`
    ShippingCents           *int64               `json:"shipping_cents"`
    CouponCents             *int64               `json:"coupon_cents"`
    OrderLevelDiscountCents *int64               `json:"order_level_discount_cents"`
    DirectRate              *float64             `json:"direct_rate"`
    IndirectRate            *float64             `json:"indirect_rate"`
    HoldPeriodDays          *int                 `json:"hold_period_days"`
    OrderID                 *int64               `json:"order_id"`

    // 形态2
    Records []struct {
        UserID                 int64      `json:"user_id"`
        OrderID                int64      `json:"order_id"`
        OrderItemID            int64      `json:"order_item_id"`
        CommissionType         string     `json:"commission_type"`
        SourceUserID           int64      `json:"source_user_id"`
        Rate                   float64    `json:"rate"`
        CalculationBasisCents  int64      `json:"calculation_basis_cents"`
        GrossCents             int64      `json:"gross_cents"`
        FeeCents               int64      `json:"fee_cents"`
        NetCents               int64      `json:"net_cents"`
        AvailableAt            *time.Time `json:"available_at"`
    } `json:"records"`
}

// POST /api/v1/commissions
func (h *CommissionUserHandler) Create(c *gin.Context) {
    var req commissionCreateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.Error(c, utils.CodeInvalidParam, "参数错误")
        return
    }

    var toSave []svc.CommissionRecord

    if len(req.Records) > 0 { // 形态2：直接落库
        for _, r := range req.Records {
            toSave = append(toSave, svc.CommissionRecord{
                UserID:           r.UserID,
                OrderID:          r.OrderID,
                OrderItemID:      r.OrderItemID,
                CommissionType:   r.CommissionType,
                SourceUserID:     r.SourceUserID,
                Rate:             r.Rate,
                CalculationBasis: r.CalculationBasisCents,
                GrossAmount:      r.GrossCents,
                Fee:              r.FeeCents,
                NetAmount:        r.NetCents,
                AvailableAt:      valueOrNow(r.AvailableAt),
            })
        }
    } else if req.PayerUserID != nil && req.ReferrerUserID != nil { // 形态1：计算后落库
        direct := 0.30
        if req.DirectRate != nil {
            direct = *req.DirectRate
        }
        indirect := 0.10
        if req.IndirectRate != nil {
            indirect = *req.IndirectRate
        }
        hold := 7
        if req.HoldPeriodDays != nil && *req.HoldPeriodDays > 0 {
            hold = *req.HoldPeriodDays
        }

        var total int64
        var itemDiscountSum int64
        for _, it := range req.Items {
            total += it.UnitPriceCents * int64(it.Quantity)
            if it.DiscountCents > 0 {
                itemDiscountSum += it.DiscountCents
            }
        }
        shipping := int64(0)
        if req.ShippingCents != nil {
            shipping = *req.ShippingCents
        }
        coupon := int64(0)
        if req.CouponCents != nil {
            coupon = *req.CouponCents
        }
        orderDisc := int64(0)
        if req.OrderLevelDiscountCents != nil {
            orderDisc = *req.OrderLevelDiscountCents
        }

        order := svc.Order{
            ID:             valueOrZero(req.OrderID),
            UserID:         *req.PayerUserID,
            TotalAmount:    total,
            ShippingAmount: shipping,
            CouponAmount:   coupon,
            DiscountAmount: orderDisc + itemDiscountSum,
        }
        toSave = svc.BuildCommissionRecords(order, *req.ReferrerUserID, direct, indirect, hold)
    } else if req.OrderID != nil { // 最小占位：仅 order_id
        // 不做任何创建，返回 created=0，便于最小集成流水通过
        utils.Success(c, gin.H{"created": 0})
        return
    } else {
        utils.Error(c, utils.CodeInvalidParam, "缺少必要参数：records 或 订单信息")
        return
    }

    if err := svc.SaveCommissionRecords(toSave); err != nil {
        utils.Error(c, utils.CodeError, err.Error())
        return
    }
    utils.Success(c, gin.H{"created": len(toSave)})
}

func valueOrNow(t *time.Time) time.Time {
    if t == nil || t.IsZero() {
        return time.Now().Add(7 * 24 * time.Hour)
    }
    return *t
}

func valueOrZero(p *int64) int64 {
    if p == nil {
        return 0
    }
    return *p
}

func parseParamUint(c *gin.Context, key string) uint {
    v := c.Param(key)
    if v == "" {
        return 0
    }
    if x, err := strconv.ParseUint(v, 10, 64); err == nil {
        return uint(x)
    }
    return 0
}

func parseQueryInt(c *gin.Context, key string, dft int) int {
    v := c.Query(key)
    if v == "" {
        return dft
    }
    if x, err := strconv.Atoi(v); err == nil {
        return x
    }
    return dft
}

// GET /api/v1/users/:id/commissions?page=&size=
func (h *CommissionUserHandler) ListUserCommissions(c *gin.Context) {
    // 授权：仅允许本人查看
    uidParam := parseParamUint(c, "id")
    current, _ := c.Get("user_id")
    if cu, ok := current.(uint); !ok || cu == 0 || cu != uidParam {
        c.JSON(http.StatusForbidden, gin.H{"code": 4031, "message": "无权访问该资源"})
        return
    }

    page := parseQueryInt(c, "page", 1)
    size := parseQueryInt(c, "size", 20)
    if size > 100 {
        size = 100
    }
    offset := (page - 1) * size

    db := database.GetDB()
    var rows []model.Commission
    var total int64
    db.Model(&model.Commission{}).Where("user_id = ?", uidParam).Count(&total)
    if err := db.Where("user_id = ?", uidParam).Order("id desc").Limit(size).Offset(offset).Find(&rows).Error; err != nil {
        utils.Error(c, utils.CodeError, err.Error())
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"total": total, "list": rows}})
}

// GET /api/v1/users/:id/commissions/summary
func (h *CommissionUserHandler) UserCommissionSummary(c *gin.Context) {
    uidParam := parseParamUint(c, "id")
    current, _ := c.Get("user_id")
    if cu, ok := current.(uint); !ok || cu == 0 || cu != uidParam {
        c.JSON(http.StatusForbidden, gin.H{"code": 4031, "message": "无权访问该资源"})
        return
    }

    db := database.GetDB()
    type row struct{ Status string; Sum decimal.Decimal }
    var rows []row
    if err := db.Table("commissions").Select("status, SUM(net_amount) as sum").Where("user_id = ?", uidParam).Group("status").Scan(&rows).Error; err != nil {
        utils.Error(c, utils.CodeError, err.Error())
        return
    }
    out := gin.H{"frozen": "0", "available": "0", "paid": "0", "reversed": "0"}
    for _, r := range rows {
        out[r.Status] = r.Sum.StringFixed(2)
    }
    utils.Success(c, gin.H{"summary": out})
}
