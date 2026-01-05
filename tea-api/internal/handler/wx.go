package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	envx "tea-test/pkg/env"

	"github.com/gin-gonic/gin"
)

type wxaCodeReq struct {
    Scene     string `json:"scene" binding:"required"`
    Page      string `json:"page" binding:"required"`
    Width     int    `json:"width"`
    IsHyaline bool   `json:"is_hyaline"`
}

// GetWxaCode 生成小程序码（wxacodeunlimit），返回 base64 PNG
// 环境变量：WECHAT_MINI_APPID / WECHAT_MINI_SECRET
func GetWxaCode(c *gin.Context) {
    var req wxaCodeReq
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "参数错误", "data": nil})
        return
    }
    appid := envx.Get("WECHAT_MINI_APPID", "")
    secret := envx.Get("WECHAT_MINI_SECRET", "")
    if appid == "" || secret == "" {
        c.JSON(http.StatusBadRequest, gin.H{"code": 4002, "message": "未配置微信小程序凭据", "data": nil})
        return
    }

    // 获取 access_token
    tokenURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appid, secret)
    httpClient := &http.Client{Timeout: 10 * time.Second}
    resp, err := httpClient.Get(tokenURL)
    if err != nil {
        c.JSON(http.StatusBadGateway, gin.H{"code": 5002, "message": "获取 access_token 失败", "data": nil})
        return
    }
    defer resp.Body.Close()
    var tokenResp struct {
        AccessToken string `json:"access_token"`
        ExpiresIn   int    `json:"expires_in"`
        ErrCode     int    `json:"errcode"`
        ErrMsg      string `json:"errmsg"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil || tokenResp.AccessToken == "" {
        c.JSON(http.StatusBadGateway, gin.H{"code": 5003, "message": "解析 access_token 响应失败", "data": tokenResp.ErrMsg})
        return
    }

    // 调用 wxacodeunlimit
    body := map[string]any{
        "scene":      req.Scene,
        "page":       req.Page,
        "width":      req.Width,
        "is_hyaline": req.IsHyaline,
    }
    if req.Width <= 0 {
        body["width"] = 240
    }
    payload, _ := json.Marshal(body)
    codeURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s", tokenResp.AccessToken)
    codeResp, err := httpClient.Post(codeURL, "application/json", bytes.NewReader(payload))
    if err != nil {
        c.JSON(http.StatusBadGateway, gin.H{"code": 5004, "message": "生成小程序码失败", "data": nil})
        return
    }
    defer codeResp.Body.Close()
    // 返回可能是 PNG 或 JSON 错误
    data, err := io.ReadAll(codeResp.Body)
    if err != nil {
        c.JSON(http.StatusBadGateway, gin.H{"code": 5005, "message": "读取响应失败", "data": nil})
        return
    }
    // 检查是否 JSON 错误
    if len(data) > 0 && data[0] == '{' {
        var errObj map[string]any
        _ = json.Unmarshal(data, &errObj)
        c.JSON(http.StatusBadGateway, gin.H{"code": 5006, "message": "生成小程序码失败", "data": errObj})
        return
    }
    b64 := base64.StdEncoding.EncodeToString(data)
    c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"image_base64": "data:image/png;base64," + b64}})
}
