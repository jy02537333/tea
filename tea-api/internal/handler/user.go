package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"tea-api/internal/config"
	"tea-api/internal/service"
	"tea-api/pkg/utils"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(),
	}
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.InvalidParam(c, err.Error())
		return
	}

	// Support multiple login flows:
	// - code (WeChat / SSO) -> userService.Login
	// - openid / username/password -> fall back to compatibility handler AuthLogin
	if req.Code != "" {
		resp, err := h.userService.Login(req.Code)
		if err != nil {
			utils.Error(c, utils.CodeError, "登录失败: "+err.Error())
			return
		}
		utils.Success(c, resp)
		return
	}

	// If request contains username/password: in dev env reuse the compatibility
	// handler; in non-dev env perform production password verification.
	if req.Username != "" && req.Password != "" {
		env := config.Config.System.Env
		if env == "local" || env == "dev" {
			AuthLogin(c)
			return
		}

		// production flow: verify password against stored hash
		resp, err := h.userService.LoginByUsername(req.Username, req.Password)
		if err != nil {
			utils.Error(c, utils.CodeError, "登录失败: "+err.Error())
			return
		}
		utils.Success(c, resp)
		return
	}

	// If request contains openid or captcha-only dev flows, fallback to compat handler
	if req.OpenID != "" || req.CaptchaID != "" || req.CaptchaCode != "" {
		AuthLogin(c)
		return
	}

	utils.InvalidParam(c, "invalid login payload: require 'code' or dev credentials/openid")
}

// DevLogin 开发环境登录（通过 openid 直接登录）
func (h *UserHandler) DevLogin(c *gin.Context) {
	// 仅在本地/开发环境开放
	env := config.Config.System.Env
	if env != "local" && env != "dev" {
		utils.Forbidden(c, "该接口仅在开发环境可用")
		return
	}

	var req struct {
		OpenID string `json:"openid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.OpenID == "" {
		utils.InvalidParam(c, "openid 不能为空")
		return
	}

	resp, err := h.userService.LoginByOpenID(req.OpenID)
	if err != nil {
		utils.Error(c, utils.CodeError, "登录失败: "+err.Error())
		return
	}
	utils.Success(c, resp)
}

// GetUserInfo 获取用户信息
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		utils.ServerError(c, "用户ID格式错误")
		return
	}

	userInfo, err := h.userService.GetUserInfo(uid)
	if err != nil {
		utils.Error(c, utils.CodeError, "获取用户信息失败: "+err.Error())
		return
	}

	utils.Success(c, userInfo)
}

// UpdateUserInfo 更新用户信息
func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		utils.ServerError(c, "用户ID格式错误")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.InvalidParam(c, err.Error())
		return
	}

	// 过滤允许更新的字段
	allowedFields := map[string]bool{
		"nickname": true,
		"avatar":   true,
		"gender":   true,
		"birthday": true,
		"province": true,
		"city":     true,
		"country":  true,
	}

	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowedFields[key] {
			filteredUpdates[key] = value
		}
	}

	if err := h.userService.UpdateUserInfo(uid, filteredUpdates); err != nil {
		utils.Error(c, utils.CodeError, "更新用户信息失败: "+err.Error())
		return
	}

	utils.Success(c, "更新成功")
}

// ChangePassword 用户修改自己的密码（需要登录）
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		utils.ServerError(c, "用户ID格式错误")
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"omitempty"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.InvalidParam(c, "invalid payload")
		return
	}

	if err := h.userService.ChangePassword(uid, req.OldPassword, req.NewPassword); err != nil {
		utils.Error(c, utils.CodeError, "修改密码失败: "+err.Error())
		return
	}

	utils.Success(c, "密码已更新")
}

// GetUserByID 根据ID获取用户信息
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.InvalidParam(c, "用户ID格式错误")
		return
	}

	userInfo, err := h.userService.GetUserInfo(uint(userID))
	if err != nil {
		utils.Error(c, utils.CodeError, "获取用户信息失败: "+err.Error())
		return
	}

	utils.Success(c, userInfo)
}

// AdminListUsers 管理员获取用户列表
func (h *UserHandler) AdminListUsers(c *gin.Context) {
	// 优先支持按 user_id 精确查询，供 Admin-FE "查看用户" 一键定位使用
	if idStr := c.Query("user_id"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil || id == 0 {
			utils.InvalidParam(c, "user_id 格式错误")
			return
		}
		userInfo, err := h.userService.GetUserInfo(uint(id))
		if err != nil {
			// 若未找到或其他错误，统一按空数组响应，方便前端回退到本地过滤
			// 这里不区分 ErrRecordNotFound，保持接口简单
			utils.Success(c, gin.H{"data": []interface{}{}})
			return
		}
		utils.Success(c, gin.H{"data": []interface{}{userInfo}})
		return
	}

	// 未指定 user_id 时，保持原有分页行为
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	users, total, err := h.userService.ListUsersPaged(page, size)
	if err != nil {
		utils.Error(c, utils.CodeError, "获取用户列表失败: "+err.Error())
		return
	}
	utils.PageSuccess(c, users, total, page, size)
}

// Refresh 刷新JWT Token
func (h *UserHandler) Refresh(c *gin.Context) {
	// 从Authorization头提取Bearer token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.Unauthorized(c, "缺少Authorization头")
		return
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		utils.Unauthorized(c, "Token格式错误")
		return
	}
	oldToken := parts[1]
	newToken, err := utils.RefreshToken(oldToken)
	if err != nil {
		utils.Error(c, utils.CodeTokenInvalid, "刷新失败: "+err.Error())
		return
	}
	utils.Success(c, gin.H{"token": newToken})
}
