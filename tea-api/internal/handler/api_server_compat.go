package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"tea-api/internal/config"
	"tea-api/internal/service"
	pkgutils "tea-api/pkg/utils"
)

// This file provides compatibility handlers ported from the project's API-Server
// mock so that tea-api can serve the same admin routes (e.g. /api/v1/admin/menus,
// /api/v1/admin/products, /api/v1/admin/stores/:id/..., /api/v1/admin/orders/...)

type mockUser struct {
	ID        int    `json:"id"`
	Nickname  string `json:"nickname"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}

type mockOrder struct {
	ID        int     `json:"id"`
	OrderNo   string  `json:"order_no"`
	StoreID   int     `json:"store_id"`
	UserID    int     `json:"user_id"`
	PayAmount float64 `json:"pay_amount"`
	Status    int     `json:"status"`
	PayStatus int     `json:"pay_status"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type mockProduct struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	CategoryID  int     `json:"category_id"`
}

type mockStore struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status int    `json:"status"`
}

var compatUsers = []mockUser{
	{ID: 1, Nickname: "测试用户A", Phone: "13800000001", Role: "member", Status: 1, CreatedAt: "2025-01-01 10:00:00"},
	{ID: 2, Nickname: "测试用户B", Phone: "13800000002", Role: "member", Status: 1, CreatedAt: "2025-01-02 11:00:00"},
}

var compatOrders = []mockOrder{
	{ID: 1001, OrderNo: "T202511160001", StoreID: 1, UserID: 201, PayAmount: 88.00, Status: 2, PayStatus: 2, CreatedAt: "2025-11-15 10:00:00", UpdatedAt: "2025-11-15 10:05:00"},
	{ID: 1002, OrderNo: "T202511160002", StoreID: 1, UserID: 202, PayAmount: 56.50, Status: 3, PayStatus: 2, CreatedAt: "2025-11-15 11:00:00", UpdatedAt: "2025-11-15 11:20:00"},
	{ID: 1003, OrderNo: "T202511160003", StoreID: 2, UserID: 203, PayAmount: 120.00, Status: 4, PayStatus: 2, CreatedAt: "2025-11-14 09:00:00", UpdatedAt: "2025-11-14 09:30:00"},
}

var compatProducts = []mockProduct{
	{ID: 101, Name: "龙井", Description: "西湖龙井茶", Price: 88.0, Stock: 100, CategoryID: 1},
	{ID: 102, Name: "普洱", Description: "熟普洱砖", Price: 120.0, Stock: 50, CategoryID: 1},
	{ID: 201, Name: "陶瓷杯", Description: "手工陶瓷杯", Price: 35.0, Stock: 200, CategoryID: 2},
}

var compatStores = []mockStore{
	{ID: 1, Name: "杭州西湖店", Status: 1},
	{ID: 2, Name: "上海静安店", Status: 1},
}

var (
	captchaStore map[string]string
	captchaMu    sync.Mutex
)

// captchaStore is a tiny in-memory store for dev-mode captcha id->code mappings

// AdminMenus 提供 /api/v1/admin/menus 的兼容返回（mock）
func AdminMenus(c *gin.Context) {
	// minimal permissions support: unauthenticated => role 0
	menus := []map[string]interface{}{
		{"key": "dashboard", "title": "首页", "tab": "testing", "min_role": 0},
		{"key": "products", "title": "商品管理", "tab": "products", "min_role": 3},
		{"key": "categories", "title": "分类管理", "tab": "categories", "min_role": 3},
		{"key": "stores", "title": "门店管理", "tab": "stores", "min_role": 4},
		{"key": "finance", "title": "财务/退款", "tab": "finance", "min_role": 4},
		{"key": "users", "title": "用户管理", "tab": "users", "min_role": 3},
		{"key": "system", "title": "系统设置", "tab": "system", "min_role": 5},
	}
	c.JSON(http.StatusOK, gin.H{"data": menus})
}

// AdminProducts 返回 admin/products 的兼容 mock 列表
func AdminProducts(c *gin.Context) {
	// return in shape { code:0, data: [...] }
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": compatProducts})
}

// AdminProductDetail 返回单条 product（兼容）
func AdminProductDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	for _, p := range compatProducts {
		if p.ID == id {
			c.JSON(http.StatusOK, gin.H{"code": 0, "data": p})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 2, "message": "product not found"})
}

// AdminStoreOrders 返回 /api/v1/admin/stores/:id/orders 的 mock 数据
func AdminStoreOrders(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var res []mockOrder
	for _, o := range compatOrders {
		if o.StoreID == id {
			res = append(res, o)
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": res, "total": len(res), "page": 1, "page_size": len(res)})
}

// AdminStoreOrdersStats 返回简单门店统计
func AdminStoreOrdersStats(c *gin.Context) {
	// return a minimal stats object
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"total_orders": 2, "completed_amount": 144.5}})
}

// AdminOrderDetail 返回 /api/v1/admin/orders/:id
func AdminOrderDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid order id"})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid order id"})
		return
	}
	for _, o := range compatOrders {
		if o.ID == id {
			c.JSON(http.StatusOK, gin.H{"code": 0, "data": o})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 2, "message": "order not found"})
}

// AdminOrdersList 返回 admin orders 列表（支持 ?store_id）
func AdminOrdersList(c *gin.Context) {
	storeID := 0
	if s := c.Query("store_id"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			storeID = v
		}
	}
	var filtered []mockOrder
	for _, o := range compatOrders {
		if storeID != 0 && o.StoreID != storeID {
			continue
		}
		filtered = append(filtered, o)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": filtered, "total": len(filtered), "page": 1, "page_size": len(filtered)})
}

// AdminUsers 返回 admin user 列表（支持 ?user_id）
func AdminUsers(c *gin.Context) {
	if s := c.Query("user_id"); s != "" {
		if id, err := strconv.Atoi(s); err == nil {
			for _, u := range compatUsers {
				if u.ID == id {
					c.JSON(http.StatusOK, gin.H{"data": []mockUser{u}})
					return
				}
			}
			c.JSON(http.StatusOK, gin.H{"data": []mockUser{}})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": compatUsers})
}

// AdminLogs 返回空的操作/访问日志结构（mock）
func AdminLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": []interface{}{}})
}

// OrdersAction 兼容 /orders/:id/* 的 POST 操作（deliver/complete/...）
func OrdersAction(c *gin.Context) {
	// Accept POST and return success
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// Health simple compatibility
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "茶心阁小程序API服务正常运行"})
}

// AuthCaptcha 返回真实的图片验证码（base64 图片 + captcha id）
func AuthCaptcha(c *gin.Context) {
	code := randomDigits(4)
	id := strconv.FormatInt(time.Now().UnixNano(), 10)
	imageData, err := generateCaptchaImage(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "captcha generation failed"})
		return
	}
	storeCaptchaCode(id, code)
	c.JSON(http.StatusOK, gin.H{"id": id, "image": imageData})
}

func storeCaptchaCode(id, code string) {
	captchaMu.Lock()
	defer captchaMu.Unlock()
	if captchaStore == nil {
		captchaStore = map[string]string{}
	}
	captchaStore[id] = code
}

func validateCaptcha(id, code string) bool {
	if id == "" || code == "" {
		return true
	}
	captchaMu.Lock()
	defer captchaMu.Unlock()
	if captchaStore == nil {
		return true
	}
	stored, ok := captchaStore[id]
	if !ok || stored == "" {
		return true
	}
	delete(captchaStore, id)
	return strings.EqualFold(stored, code)
}

func randomDigits(length int) string {
	if length <= 0 {
		length = 4
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	var builder strings.Builder
	for i := 0; i < length; i++ {
		builder.WriteByte(byte('0' + rnd.Intn(10)))
	}
	return builder.String()
}

func generateCaptchaImage(code string) (string, error) {
	const width = 120
	const height = 40
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bg := color.RGBA{R: 245, G: 247, B: 252, A: 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	// draw noise lines
	for i := 0; i < 4; i++ {
		start := image.Point{X: rnd.Intn(width), Y: rnd.Intn(height)}
		end := image.Point{X: rnd.Intn(width), Y: rnd.Intn(height)}
		col := color.RGBA{R: uint8(rnd.Intn(156)), G: uint8(rnd.Intn(156)), B: uint8(rnd.Intn(156)), A: 255}
		plotLine(img, start, end, col)
	}
	// noise dots
	for i := 0; i < 80; i++ {
		x := rnd.Intn(width)
		y := rnd.Intn(height)
		img.Set(x, y, color.RGBA{R: uint8(120 + rnd.Intn(120)), G: uint8(120 + rnd.Intn(120)), B: uint8(120 + rnd.Intn(120)), A: 255})
	}

	face := basicfont.Face7x13
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{R: 20, G: 40, B: 70, A: 255}),
		Face: face,
	}
	charSpacing := (width - 20) / len(code)
	for i, r := range code {
		x := 10 + i*charSpacing + rnd.Intn(6)
		y := height/2 + 10 + rnd.Intn(6) - 3
		drawer.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
		drawer.DrawString(string(r))
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + encoded, nil
}

func plotLine(img *image.RGBA, start, end image.Point, col color.Color) {
	dx := abs(end.X - start.X)
	dy := abs(end.Y - start.Y)
	sx := -1
	if start.X < end.X {
		sx = 1
	}
	sy := -1
	if start.Y < end.Y {
		sy = 1
	}
	err := dx - dy
	x := start.X
	y := start.Y
	for {
		img.Set(x, y, col)
		if x == end.X && y == end.Y {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// AuthLogin 为开发用途的登录接口，接受 openid 或 username/password，并检查 captcha（若提供）
func AuthLogin(c *gin.Context) {
	// only allow in dev/local
	env := config.Config.System.Env
	if env != "local" && env != "dev" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "该接口仅在开发环境可用"})
		return
	}

	var req map[string]interface{}

	// Read raw body first to allow tolerant parsing (JSON or form-encoded)
	raw, _ := io.ReadAll(c.Request.Body)
	if len(raw) > 0 {
		// try JSON first
		if err := json.Unmarshal(raw, &req); err != nil {
			// if not JSON, try parse form values from raw body (urlencoded)
			if vals, err2 := url.ParseQuery(string(raw)); err2 == nil && len(vals) > 0 {
				req = map[string]interface{}{}
				for k, v := range vals {
					if len(v) > 0 {
						req[k] = v[0]
					}
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
				return
			}
		}
	} else {
		// empty body: try form (e.g., Content-Type: application/x-www-form-urlencoded)
		if err := c.Request.ParseForm(); err == nil && len(c.Request.Form) > 0 {
			req = map[string]interface{}{}
			for k, v := range c.Request.Form {
				if len(v) > 0 {
					req[k] = v[0]
				}
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}
	}

	// optional captcha check (tolerant for dev compatibility)
	if idV, ok := req["captcha_id"]; ok {
		if codeV, ok2 := req["captcha_code"]; ok2 {
			idStr := fmt.Sprint(idV)
			codeStr := fmt.Sprint(codeV)
			if !validateCaptcha(idStr, codeStr) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid captcha"})
				return
			}
		}
	}

	// determine role by openid or username
	role := "user"
	name := "访客"
	openid := ""
	if v, ok := req["openid"]; ok {
		openid = fmt.Sprint(v)
		switch strings.ToLower(openid) {
		case "admin_openid", "admin":
			role = "admin"
			name = "超级管理员"
		case "store_openid", "store":
			role = "store"
			name = "门店管理员"
		case "partner_openid", "partner":
			role = "partner"
			name = "合伙人"
		default:
			role = "user"
			name = "访客"
		}
	} else if v, ok := req["username"]; ok {
		username := strings.TrimSpace(fmt.Sprint(v))
		pass := fmt.Sprint(req["password"])
		if username != "" && pass != "" {
			if realResp, err := service.NewUserService().LoginByUsername(username, pass); err == nil {
				claims, _ := pkgutils.ParseToken(realResp.Token)
				if claims != nil && claims.Role != "" {
					role = claims.Role
				}
				if info, ok := realResp.UserInfo.(service.UserInfo); ok {
					name = info.Nickname
					if name == "" {
						name = info.Phone
					}
					if name == "" {
						name = username
					}
					openid = info.OpenID
				}
				data := map[string]interface{}{"token": realResp.Token, "role": role, "name": name}
				c.JSON(http.StatusOK, gin.H{"data": data})
				return
			}
		}
		switch strings.ToLower(username) {
		case "admin":
			if pass == "pass" || pass == "admin" || pass == "Admin@123" {
				role = "admin"
				name = "超级管理员"
			}
		case "store":
			if pass == "pass" || pass == "store" {
				role = "store"
				name = "门店管理员"
			}
		case "partner":
			if pass == "pass" || pass == "partner" {
				role = "partner"
				name = "合伙人"
			}
		default:
			role = "user"
			name = "访客"
		}
	}

	// generate token via existing utils.GenerateToken (use dummy user id)
	token, err := pkgutils.GenerateToken(1000, openid, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	data := map[string]interface{}{"token": token, "role": role, "name": name, "iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix()}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// AuthMe 简单返回 token 中的 payload
func AuthMe(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
		return
	}
	parts := strings.Fields(auth)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		return
	}
	claims, err := pkgutils.ParseToken(parts[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": claims})
}

// AdminOrdersExport 返回 CSV 导出用于 Admin-FE 的导出流程（简化）
func AdminOrdersExport(c *gin.Context) {
	// build CSV from compatOrders
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=orders_export.csv")
	w := c.Writer
	csvw := csv.NewWriter(w)
	defer csvw.Flush()
	_ = csvw.Write([]string{"id", "order_no", "store_id", "user_id", "pay_amount", "status", "pay_status", "created_at"})
	for _, o := range compatOrders {
		_ = csvw.Write([]string{
			strconv.Itoa(o.ID), o.OrderNo, strconv.Itoa(o.StoreID), strconv.Itoa(o.UserID), fmt.Sprintf("%.2f", o.PayAmount), strconv.Itoa(o.Status), strconv.Itoa(o.PayStatus), o.CreatedAt,
		})
	}
}

// AdminLogsExport 返回操作日志导出（简化，CSV）
func AdminLogsExport(c *gin.Context) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=logs_export.csv")
	w := c.Writer
	csvw := csv.NewWriter(w)
	defer csvw.Flush()
	_ = csvw.Write([]string{"time", "user", "action", "module", "details"})
	// no logs in mock, write header only
}
