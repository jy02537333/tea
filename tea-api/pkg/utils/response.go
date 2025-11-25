package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageResponse 分页响应结构
type PageResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
}

// 响应码常量
const (
	CodeSuccess      = 0    // 成功
	CodeError        = 1    // 通用错误
	CodeInvalidParam = 1001 // 参数错误
	CodeUnauthorized = 1002 // 未授权
	CodeForbidden    = 1003 // 禁止访问
	CodeNotFound     = 1004 // 资源不存在
	CodeServerError  = 1005 // 服务器错误
	CodeTokenExpired = 1006 // Token过期
	CodeTokenInvalid = 1007 // Token无效
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// PageSuccess 分页成功响应
func PageSuccess(c *gin.Context, data interface{}, total int64, page, size int) {
	c.JSON(http.StatusOK, PageResponse{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

// InvalidParam 参数错误响应
func InvalidParam(c *gin.Context, message string) {
	Error(c, CodeInvalidParam, message)
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
	})
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
	})
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeNotFound,
		Message: message,
	})
}

// ServerError 服务器错误响应
func ServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeServerError,
		Message: message,
	})
}
