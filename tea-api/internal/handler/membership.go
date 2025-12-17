package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/database"
	"tea-api/pkg/response"
)

// membershipOrderService 定义创建会员订单所需的服务接口，便于后续测试注入 fake 实现。
type membershipOrderService interface {
	CreateMembershipOrder(userID, packageID uint, remark string) (*model.Order, error)
}

// MembershipHandler 小程序/用户侧会员相关接口
type MembershipHandler struct {
	orderSvc membershipOrderService
}

func NewMembershipHandler() *MembershipHandler {
	return &MembershipHandler{orderSvc: service.NewOrderService()}
}

// ListPackages GET /api/v1/membership-packages
// 提供给小程序等前台使用的会员套餐列表，只读接口。
// 默认按 type=membership 过滤，并按 id 倒序返回分页结果。
func (h *MembershipHandler) ListPackages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	typeFilter := strings.TrimSpace(c.DefaultQuery("type", "membership"))

	db := database.GetDB()
	q := db.Model(&model.MembershipPackage{})
	if typeFilter != "" {
		q = q.Where("type = ?", typeFilter)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	var list []model.MembershipPackage
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&list).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithPagination(c, list, total, page, limit)
}

// createMembershipOrderReq 创建会员订单请求体
type createMembershipOrderReq struct {
	PackageID uint   `json:"package_id" binding:"required"`
	Remark    string `json:"remark"`
}

// CreateOrder POST /api/v1/membership-orders
// 为当前登录用户基于指定会员套餐创建一笔会员订单。
func (h *MembershipHandler) CreateOrder(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))

	var req createMembershipOrderReq
	if err := c.ShouldBindJSON(&req); err != nil || req.PackageID == 0 {
		response.BadRequest(c, "参数错误")
		return
	}

	order, err := h.orderSvc.CreateMembershipOrder(userID, req.PackageID, req.Remark)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	payAmt, _ := order.PayAmount.Float64()
	response.Success(c, gin.H{
		"order_id":   order.ID,
		"order_no":   order.OrderNo,
		"pay_amount": payAmt,
	})
}
