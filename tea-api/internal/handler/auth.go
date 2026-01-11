package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	// 统一签发改用 utils.GenerateToken，避免与校验配置不一致
	pkgutils "tea-api/pkg/utils"
)

// RegisterAuthRoutes registers authentication endpoints.
func RegisterAuthRoutes(r *gin.RouterGroup) {
	r.POST("/auth/login", Login)
	// 开发/联调用途的验证码获取端点，用于前端登录页展示图形验证码
	// 返回字段包含：captcha_id、image_base64 等
	r.GET("/auth/captcha", AuthCaptcha)
}

// Login supports phone+code or wechat_code per PRD.
// TODO: integrate with SMS/WeChat providers and issue JWT.
func Login(c *gin.Context) {
	type Req struct {
		Phone      string `json:"phone"`
		Code       string `json:"code"`
		WechatCode string `json:"wechat_code"`
	}
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    4001,
			"message": "请求体解析失败或格式不合法",
			"data":    nil,
		})
		return
	}

	// 参数校验：必须满足 (phone && code) 或 (wechat_code)
	usingPhone := req.Phone != "" && req.Code != ""
	usingWechat := req.WechatCode != ""
	if !(usingPhone || usingWechat) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    4001,
			"message": "参数缺失：需提供手机号+验证码或微信登录code",
			"data":    nil,
		})
		return
	}

	// 校验占位（待接入真实服务）
	var userID uint
	if usingPhone {
		if !validateSMSCode(req.Phone, req.Code) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    4002,
				"message": "验证码校验失败",
				"data":    nil,
			})
			return
		}
		userID = resolveOrCreateUserByPhone(req.Phone)
	} else {
		if !validateWeChatCode(req.WechatCode) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    4002,
				"message": "微信登录校验失败",
				"data":    nil,
			})
			return
		}
		userID = resolveOrCreateUserByWechat(req.WechatCode)
	}

	// 使用统一签发函数，保证与校验端完全一致
	signed, err := pkgutils.GenerateToken(userID, "", "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "签发令牌失败", "data": nil})
		return
	}
	data := map[string]interface{}{
		"token": signed,
		"role":  "user",
		"name":  "访客",
		// iat/exp 可在前端解析 token 获取，或后续由 GenerateToken 返回结构体携带
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "OK", "data": data})
}

// --- 以下为占位实现，后续接入真实服务 ---

func validateSMSCode(phone, code string) bool {
	// Placeholder: 仅示例长度与简易匹配；接入短信服务后替换为真实校验
	if len(phone) < 6 || len(code) < 4 {
		return false
	}
	return true
}

func validateWeChatCode(wechatCode string) bool {
	// Placeholder: 仅示例长度校验；接入微信 jscode2session 后替换
	return len(wechatCode) >= 6
}

func resolveOrCreateUserByPhone(phone string) uint {
	// Placeholder: 依据手机号查找或注册用户，返回用户ID
	return 1
}

func resolveOrCreateUserByWechat(wechatCode string) uint {
	// Placeholder: 依据微信 code 交换 openid/unionid 并查找或注册用户
	return 2
}

//
