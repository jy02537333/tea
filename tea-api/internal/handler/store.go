package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/response"
)

type StoreHandler struct{ svc *service.StoreService }

func NewStoreHandler() *StoreHandler { return &StoreHandler{svc: service.NewStoreService()} }

func (h *StoreHandler) Create(c *gin.Context) {
	var req struct {
		Name          string  `json:"name" binding:"required"`
		Address       string  `json:"address"`
		Phone         string  `json:"phone"`
		Latitude      float64 `json:"latitude"`
		Longitude     float64 `json:"longitude"`
		BusinessHours string  `json:"business_hours"`
		Images        string  `json:"images"`
		Status        int     `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	st := &model.Store{Name: req.Name, Address: req.Address, Phone: req.Phone, Latitude: req.Latitude, Longitude: req.Longitude, BusinessHours: req.BusinessHours, Images: req.Images, Status: req.Status}
	if err := h.svc.CreateStore(st); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, st)
}

func (h *StoreHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法ID")
		return
	}
	var req struct {
		Name          string  `json:"name"`
		Address       string  `json:"address"`
		Phone         string  `json:"phone"`
		Latitude      float64 `json:"latitude"`
		Longitude     float64 `json:"longitude"`
		BusinessHours string  `json:"business_hours"`
		Images        string  `json:"images"`
		Status        int     `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	updates := map[string]any{
		"name":           req.Name,
		"address":        req.Address,
		"phone":          req.Phone,
		"latitude":       req.Latitude,
		"longitude":      req.Longitude,
		"business_hours": req.BusinessHours,
		"images":         req.Images,
		"status":         req.Status,
		"updated_at":     time.Now(),
	}
	if err := h.svc.UpdateStore(uint(id), updates); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

func (h *StoreHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法ID")
		return
	}
	if err := h.svc.DeleteStore(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

func (h *StoreHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法ID")
		return
	}
	st, err := h.svc.GetStore(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, st)
}

func (h *StoreHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	var statusPtr *int
	if v := c.Query("status"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			statusPtr = &n
		}
	}
	var latPtr, lngPtr *float64
	if v := c.Query("lat"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			latPtr = &f
		}
	}
	if v := c.Query("lng"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			lngPtr = &f
		}
	}

	list, total, err := h.svc.ListStores(page, limit, statusPtr, latPtr, lngPtr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessWithPagination(c, list, total, page, limit)
}

// OrderStats 门店订单统计（管理员）
func (h *StoreHandler) OrderStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法ID")
		return
	}
	ordSvc := service.NewOrderService()
	stats, err := ordSvc.GetStoreOrderStats(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, stats)
}
