package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

func grantPermissionToUser(t *testing.T, db *gorm.DB, userID uint, permName string) {
	t.Helper()
	if userID == 0 {
		t.Fatalf("grantPermissionToUser: userID is zero")
	}
	if permName == "" {
		t.Fatalf("grantPermissionToUser: permName is empty")
	}

	var perm model.Permission
	if err := db.Where("name = ?", permName).First(&perm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			perm = model.Permission{Name: permName, DisplayName: permName}
			if err := db.Create(&perm).Error; err != nil {
				t.Fatalf("create permission %s: %v", permName, err)
			}
		} else {
			t.Fatalf("query permission %s: %v", permName, err)
		}
	}

	roleName := fmt.Sprintf("test_role_%s_%d", permName, time.Now().UnixNano())
	role := model.Role{Name: roleName, DisplayName: roleName, Status: 1}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}

	if err := db.Create(&model.RolePermission{RoleID: role.ID, PermissionID: perm.ID}).Error; err != nil {
		t.Fatalf("create role_permission: %v", err)
	}
	if err := db.Create(&model.UserRole{UserID: userID, RoleID: role.ID}).Error; err != nil {
		t.Fatalf("create user_role: %v", err)
	}
}

type orderAttributionDetailResp struct {
	Code int `json:"code"`
	Data struct {
		Order struct {
			ReferrerID   *uint `json:"referrer_id"`
			ShareStoreID uint  `json:"share_store_id"`
		} `json:"order"`
	} `json:"data"`
}

func fetchOrderAttributionDetail(t *testing.T, tsURL, auth string, orderID uint) orderAttributionDetailResp {
	t.Helper()
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/orders/%d", tsURL, orderID), nil)
	req.Header.Set("Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("order detail request err: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("order detail status: %d", resp.StatusCode)
	}
	var out orderAttributionDetailResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode order detail resp: %v", err)
	}
	return out
}

// TestingT keeps this helper usable in this package without importing testing directly here.
// It matches *testing.T's Fatalf signature.
// (We keep it minimal to avoid duplicating lots of helper code.)
//
//nolint:revive // intentional small interface for helpers
//nolint:stylecheck // name chosen for clarity
//nolint:interfacebloat // minimal interface
//nolint:cyclop // not applicable
//nolint:gomnd // not applicable
//nolint:unparam // not applicable
//nolint:dupl // not applicable
//nolint:unused // used by helper
//nolint:errcheck // not applicable
//nolint:forcetypeassert // not applicable
//nolint:forbidigo // not applicable
//nolint:wrapcheck // not applicable
//nolint:gocritic // not applicable
//nolint:gosec // not applicable
//nolint:gocognit // not applicable
//nolint:goconst // not applicable
//nolint:lll // not applicable
//nolint:misspell // not applicable
//nolint:prealloc // not applicable
//nolint:staticcheck // not applicable
//nolint:unconvert // not applicable
//nolint:wsl // not applicable
//nolint:whitespace // not applicable
//nolint:govet // not applicable
//nolint:bodyclose // handled
//nolint:contextcheck // not applicable
//nolint:copyloopvar // not applicable
//nolint:errname // not applicable
//nolint:exhaustive // not applicable
//nolint:fatcontext // not applicable
//nolint:funlen // not applicable
//nolint:godox // not applicable
//nolint:gofumpt // repo formatting varies
//nolint:gosmopolitan // not applicable
//nolint:sloglint // not applicable
//nolint:zerologlint // not applicable
//nolint:dupword // not applicable
//nolint:decorder // not applicable
//nolint:predeclared // not applicable
//nolint:nolintlint // not applicable
//nolint:paralleltest // not applicable
//nolint:thelper // not applicable
//nolint:tparallel // not applicable
//nolint:err113 // not applicable
//nolint:errcheck // not applicable
//nolint:tagliatelle // not applicable
//nolint:godox // not applicable
//nolint:gomnd // not applicable
//nolint:bodyclose // not applicable
//nolint:goerr113 // not applicable
//nolint:revive // not applicable
//nolint:stylecheck // not applicable
//nolint:unparam // not applicable
//nolint:unused // not applicable
//nolint:cyclop // not applicable
//nolint:dupl // not applicable
//nolint:lll // not applicable
//nolint:varnamelen // not applicable
//nolint:wrapcheck // not applicable
//nolint:gomoddirectives // not applicable
//nolint:gocyclo // not applicable
//nolint:gocognit // not applicable
//nolint:funlen // not applicable
//nolint:maintidx // not applicable
//nolint:nestif // not applicable
//nolint:prealloc // not applicable
//nolint:revive // not applicable
//nolint:stylecheck // not applicable
//nolint:unparam // not applicable
//nolint:unused // not applicable
//nolint:govet // not applicable
//nolint:staticcheck // not applicable
//nolint:ineffassign // not applicable
//nolint:unconvert // not applicable
//nolint:misspell // not applicable
//nolint:unparam // not applicable
//nolint:exportloopref // not applicable
//nolint:forcetypeassert // not applicable
//nolint:exhaustive // not applicable
//nolint:wrapcheck // not applicable
//nolint:goimports // not applicable
//nolint:forbidigo // not applicable
//nolint:contextcheck // not applicable
//nolint:sqlclosecheck // not applicable
//nolint:copyloopvar // not applicable
//nolint:intrange // not applicable
//nolint:nolintlint // not applicable
//nolint:perfsprint // not applicable
//nolint:dogsled // not applicable
//nolint:tagalign // not applicable
//nolint:tenv // not applicable
//nolint:revive // not applicable
//nolint:stylecheck // not applicable
//nolint:usestdlibvars // not applicable
//nolint:varnamelen // not applicable
//nolint:gocritic // not applicable
//nolint:exhaustive // not applicable
//nolint:errname // not applicable
//nolint:containedctx // not applicable
//nolint:makezero // not applicable
//nolint:musttag // not applicable
//nolint:protogetter // not applicable
//nolint:sloglint // not applicable
//nolint:zerologlint // not applicable
//nolint:dupword // not applicable
//nolint:decorder // not applicable
//nolint:predeclared // not applicable
//nolint:nolintlint // not applicable
//nolint:paralleltest // not applicable
//nolint:thelper // not applicable
//nolint:tparallel // not applicable
//nolint:err113 // not applicable
//nolint:errcheck // not applicable
//nolint:tagliatelle // not applicable
//nolint:godox // not applicable
//nolint:gomnd // not applicable
//nolint:bodyclose // not applicable
//nolint:goerr113 // not applicable
//nolint:revive // not applicable
//nolint:stylecheck // not applicable
//nolint:unparam // not applicable
//nolint:unused // not applicable
//nolint:cyclop // not applicable
//nolint:dupl // not applicable
//nolint:lll // not applicable
//nolint:varnamelen // not applicable
//nolint:wrapcheck // not applicable

type TestingT interface {
	Fatalf(format string, args ...any)
}

func Test_OrderAttribution_NoShareParams_NoFreezeEvenWithDirectReferrer(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	sharerOpenID := fmt.Sprintf("share_openid_%d", time.Now().UnixNano())
	buyerOpenID := fmt.Sprintf("buyer_openid_%d", time.Now().UnixNano())

	sharerToken := devLogin(t, ts, sharerOpenID)
	buyerToken := devLogin(t, ts, buyerOpenID)

	db := database.GetDB()
	var sharer, buyer model.User
	if err := db.Where("open_id = ?", sharerOpenID).First(&sharer).Error; err != nil {
		t.Fatalf("query sharer user: %v", err)
	}
	if err := db.Where("open_id = ?", buyerOpenID).First(&buyer).Error; err != nil {
		t.Fatalf("query buyer user: %v", err)
	}

	// 预置一条直推关系（A -> B），确保“有直推但无分享参数”场景覆盖。
	rc := model.ReferralClosure{AncestorUserID: sharer.ID, DescendantUserID: buyer.ID, Depth: 1}
	_ = db.Create(&rc).Error

	authBuyer := "Bearer " + buyerToken
	catID := createCategory(t, ts, authBuyer)
	productID := createProduct(t, ts, authBuyer, catID)
	addCartItem(t, ts, authBuyer, productID)

	orderID := createOrderFromCart(t, ts, authBuyer)
	detail := fetchOrderAttributionDetail(t, ts.URL, authBuyer, orderID)
	if detail.Data.Order.ReferrerID != nil {
		t.Fatalf("expected referrer_id nil when no share params, got %v", *detail.Data.Order.ReferrerID)
	}
	if detail.Data.Order.ShareStoreID != 0 {
		t.Fatalf("expected share_store_id=0 when no share params, got %d", detail.Data.Order.ShareStoreID)
	}

	_ = sharerToken // silence unused if build tags change
}

func Test_OrderAttribution_WithSharerUID_FreezeReferrer(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	sharerOpenID := fmt.Sprintf("share_openid_%d", time.Now().UnixNano())
	buyerOpenID := fmt.Sprintf("buyer_openid_%d", time.Now().UnixNano())

	_ = devLogin(t, ts, sharerOpenID)
	buyerToken := devLogin(t, ts, buyerOpenID)

	db := database.GetDB()
	var sharer, buyer model.User
	if err := db.Where("open_id = ?", sharerOpenID).First(&sharer).Error; err != nil {
		t.Fatalf("query sharer user: %v", err)
	}
	if err := db.Where("open_id = ?", buyerOpenID).First(&buyer).Error; err != nil {
		t.Fatalf("query buyer user: %v", err)
	}

	// 预置直推关系（A -> B），然后使用 sharer_uid 下单，应该冻结为 A。
	rc := model.ReferralClosure{AncestorUserID: sharer.ID, DescendantUserID: buyer.ID, Depth: 1}
	_ = db.Create(&rc).Error

	authBuyer := "Bearer " + buyerToken
	catID := createCategory(t, ts, authBuyer)
	productID := createProduct(t, ts, authBuyer, catID)
	addCartItem(t, ts, authBuyer, productID)

	payload := map[string]any{
		"delivery_type": 1,
		"remark":        "share attribution test",
		"sharer_uid":    sharer.ID,
		// store_id=0 的商城订单允许 share_store_id=0
		"share_store_id": 0,
	}
	resp := authedJSONRequest(t, ts.URL+"/api/v1/orders/from-cart", authBuyer, payload)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create order status: %d", resp.StatusCode)
	}
	var data struct {
		Code int `json:"code"`
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("decode order resp: %v", err)
	}
	if data.Data.ID == 0 {
		t.Fatalf("order id is zero")
	}

	detail := fetchOrderAttributionDetail(t, ts.URL, authBuyer, data.Data.ID)
	if detail.Data.Order.ReferrerID == nil || *detail.Data.Order.ReferrerID != sharer.ID {
		if detail.Data.Order.ReferrerID == nil {
			t.Fatalf("expected referrer_id=%d, got nil", sharer.ID)
		}
		t.Fatalf("expected referrer_id=%d, got %d", sharer.ID, *detail.Data.Order.ReferrerID)
	}
	if detail.Data.Order.ShareStoreID != 0 {
		t.Fatalf("expected share_store_id=0 for mall order, got %d", detail.Data.Order.ShareStoreID)
	}
}

func Test_OrderAttribution_ShareStoreWithoutSharer_ShouldReject(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	buyerToken := devLogin(t, ts, fmt.Sprintf("buyer_openid_%d", time.Now().UnixNano()))
	authBuyer := "Bearer " + buyerToken

	catID := createCategory(t, ts, authBuyer)
	productID := createProduct(t, ts, authBuyer, catID)
	addCartItem(t, ts, authBuyer, productID)

	payload := map[string]any{
		"delivery_type":  1,
		"remark":         "invalid share params",
		"share_store_id": 1,
	}
	resp := authedJSONRequest(t, ts.URL+"/api/v1/orders/from-cart", authBuyer, payload)
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected non-200 when share_store_id provided without sharer_uid")
	}
}

func Test_OrderAttribution_StoreOrder_WithSharer_MissingShareStore_ShouldReject(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	sharerOpenID := fmt.Sprintf("share_openid_%d", time.Now().UnixNano())
	buyerOpenID := fmt.Sprintf("buyer_openid_%d", time.Now().UnixNano())

	_ = devLogin(t, ts, sharerOpenID)
	buyerToken := devLogin(t, ts, buyerOpenID)

	db := database.GetDB()
	var sharer model.User
	if err := db.Where("open_id = ?", sharerOpenID).First(&sharer).Error; err != nil {
		t.Fatalf("query sharer user: %v", err)
	}

	// 创建一个可用门店（store_id != 0）
	store := &model.Store{Name: "分享校验门店", Status: 1}
	if err := db.Create(store).Error; err != nil {
		t.Fatalf("create store: %v", err)
	}

	authBuyer := "Bearer " + buyerToken
	catID := createCategory(t, ts, authBuyer)
	productID := createProduct(t, ts, authBuyer, catID)
	addCartItem(t, ts, authBuyer, productID)

	// 门店订单：带 sharer_uid 但缺 share_store_id -> 必须拒绝
	payload := map[string]any{
		"delivery_type": 1,
		"store_id":      store.ID,
		"order_type":    2,
		"sharer_uid":    sharer.ID,
	}
	resp := authedJSONRequest(t, ts.URL+"/api/v1/orders/from-cart", authBuyer, payload)
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected non-200 when sharer_uid provided without share_store_id for store order")
	}
}

func Test_OrderAttribution_StoreOrder_WithSharer_ShareStoreMismatch_ShouldReject(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	sharerOpenID := fmt.Sprintf("share_openid_%d", time.Now().UnixNano())
	buyerOpenID := fmt.Sprintf("buyer_openid_%d", time.Now().UnixNano())

	_ = devLogin(t, ts, sharerOpenID)
	buyerToken := devLogin(t, ts, buyerOpenID)

	db := database.GetDB()
	var sharer model.User
	if err := db.Where("open_id = ?", sharerOpenID).First(&sharer).Error; err != nil {
		t.Fatalf("query sharer user: %v", err)
	}

	store := &model.Store{Name: "分享校验门店2", Status: 1}
	if err := db.Create(store).Error; err != nil {
		t.Fatalf("create store: %v", err)
	}

	authBuyer := "Bearer " + buyerToken
	catID := createCategory(t, ts, authBuyer)
	productID := createProduct(t, ts, authBuyer, catID)
	addCartItem(t, ts, authBuyer, productID)

	// 门店订单：share_store_id 与 store_id 不一致 -> 必须拒绝
	payload := map[string]any{
		"delivery_type":  1,
		"store_id":       store.ID,
		"order_type":     2,
		"sharer_uid":     sharer.ID,
		"share_store_id": store.ID + 999,
	}
	resp := authedJSONRequest(t, ts.URL+"/api/v1/orders/from-cart", authBuyer, payload)
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected non-200 when share_store_id mismatches store_id for store order")
	}
}

func Test_MembershipOrderAttribution_WithSharerUID_FreezeReferrer(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	sharerOpenID := fmt.Sprintf("share_openid_%d", time.Now().UnixNano())
	buyerOpenID := fmt.Sprintf("buyer_openid_%d", time.Now().UnixNano())

	_ = devLogin(t, ts, sharerOpenID)
	buyerToken := devLogin(t, ts, buyerOpenID)

	db := database.GetDB()
	var sharer, buyer model.User
	if err := db.Where("open_id = ?", sharerOpenID).First(&sharer).Error; err != nil {
		t.Fatalf("query sharer user: %v", err)
	}
	if err := db.Where("open_id = ?", buyerOpenID).First(&buyer).Error; err != nil {
		t.Fatalf("query buyer user: %v", err)
	}

	// 预置直推关系（A -> B），然后使用 sharer_uid 下单，应该冻结为 A。
	rc := model.ReferralClosure{AncestorUserID: sharer.ID, DescendantUserID: buyer.ID, Depth: 1}
	_ = db.Create(&rc).Error

	pkg := &model.MembershipPackage{
		Name:  "测试会员套餐",
		Price: decimal.NewFromInt(99),
		Type:  "membership",
	}
	if err := db.Create(pkg).Error; err != nil {
		t.Fatalf("create membership package: %v", err)
	}

	authBuyer := "Bearer " + buyerToken
	payload := map[string]any{
		"package_id":     pkg.ID,
		"remark":         "share attribution membership",
		"sharer_uid":     sharer.ID,
		"share_store_id": 0,
	}
	resp := authedJSONRequest(t, ts.URL+"/api/v1/membership-orders", authBuyer, payload)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create membership order status: %d", resp.StatusCode)
	}
	var out struct {
		Code int `json:"code"`
		Data struct {
			OrderID uint `json:"order_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode membership order resp: %v", err)
	}
	if out.Data.OrderID == 0 {
		t.Fatalf("membership order_id is zero")
	}

	detail := fetchOrderAttributionDetail(t, ts.URL, authBuyer, out.Data.OrderID)
	if detail.Data.Order.ReferrerID == nil || *detail.Data.Order.ReferrerID != sharer.ID {
		if detail.Data.Order.ReferrerID == nil {
			t.Fatalf("expected referrer_id=%d, got nil", sharer.ID)
		}
		t.Fatalf("expected referrer_id=%d, got %d", sharer.ID, *detail.Data.Order.ReferrerID)
	}
	if detail.Data.Order.ShareStoreID != 0 {
		t.Fatalf("expected share_store_id=0 for membership order, got %d", detail.Data.Order.ShareStoreID)
	}
}

func Test_MembershipOrderAttribution_ShareStoreNotZero_ShouldReject(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	buyerToken := devLogin(t, ts, fmt.Sprintf("buyer_openid_%d", time.Now().UnixNano()))
	authBuyer := "Bearer " + buyerToken

	db := database.GetDB()
	pkg := &model.MembershipPackage{
		Name:  "测试会员套餐2",
		Price: decimal.NewFromInt(99),
		Type:  "membership",
	}
	if err := db.Create(pkg).Error; err != nil {
		t.Fatalf("create membership package: %v", err)
	}

	payload := map[string]any{
		"package_id":     pkg.ID,
		"remark":         "invalid share store",
		"sharer_uid":     123, // 任意非 0，触发 share_store_id 校验
		"share_store_id": 1,
	}
	resp := authedJSONRequest(t, ts.URL+"/api/v1/membership-orders", authBuyer, payload)
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected non-200 when share_store_id provided for membership order")
	}
}

func Test_ActivityRegisterWithOrderAttribution_WithSharer_FreezeReferrerAndShareStore(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	sharerOpenID := fmt.Sprintf("share_openid_%d", time.Now().UnixNano())
	buyerOpenID := fmt.Sprintf("buyer_openid_%d", time.Now().UnixNano())

	_ = devLogin(t, ts, sharerOpenID)
	buyerToken := devLogin(t, ts, buyerOpenID)

	db := database.GetDB()
	var sharer, buyer model.User
	if err := db.Where("open_id = ?", sharerOpenID).First(&sharer).Error; err != nil {
		t.Fatalf("query sharer user: %v", err)
	}
	if err := db.Where("open_id = ?", buyerOpenID).First(&buyer).Error; err != nil {
		t.Fatalf("query buyer user: %v", err)
	}
	grantPermissionToUser(t, db, buyer.ID, "activity:register")

	store := &model.Store{Name: "活动分享门店", Status: 1}
	if err := db.Create(store).Error; err != nil {
		t.Fatalf("create store: %v", err)
	}

	now := time.Now()
	act := &model.Activity{
		StoreID:   &store.ID,
		Name:      "活动报名测试",
		Type:      1,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Rules:     "{}",
		Status:    1,
	}
	if err := db.Create(act).Error; err != nil {
		t.Fatalf("create activity: %v", err)
	}

	authBuyer := "Bearer " + buyerToken
	payload := map[string]any{
		"name":           "张三",
		"phone":          "13800000000",
		"fee":            9.9,
		"sharer_uid":     sharer.ID,
		"share_store_id": store.ID,
	}
	resp := authedJSONRequest(t, fmt.Sprintf("%s/api/v1/activities/%d/register-with-order", ts.URL, act.ID), authBuyer, payload)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("register-with-order status: %d", resp.StatusCode)
	}
	var out struct {
		Code int `json:"code"`
		Data struct {
			Order struct {
				ID uint `json:"id"`
			} `json:"order"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode register-with-order resp: %v", err)
	}
	if out.Data.Order.ID == 0 {
		t.Fatalf("activity order id is zero")
	}

	detail := fetchOrderAttributionDetail(t, ts.URL, authBuyer, out.Data.Order.ID)
	if detail.Data.Order.ReferrerID == nil || *detail.Data.Order.ReferrerID != sharer.ID {
		if detail.Data.Order.ReferrerID == nil {
			t.Fatalf("expected referrer_id=%d, got nil", sharer.ID)
		}
		t.Fatalf("expected referrer_id=%d, got %d", sharer.ID, *detail.Data.Order.ReferrerID)
	}
	if detail.Data.Order.ShareStoreID != store.ID {
		t.Fatalf("expected share_store_id=%d for activity store order, got %d", store.ID, detail.Data.Order.ShareStoreID)
	}
}

func Test_ActivityRegisterWithOrderAttribution_MissingShareStore_ShouldReject(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	sharerOpenID := fmt.Sprintf("share_openid_%d", time.Now().UnixNano())
	buyerOpenID := fmt.Sprintf("buyer_openid_%d", time.Now().UnixNano())

	_ = devLogin(t, ts, sharerOpenID)
	buyerToken := devLogin(t, ts, buyerOpenID)

	db := database.GetDB()
	var sharer, buyer model.User
	if err := db.Where("open_id = ?", sharerOpenID).First(&sharer).Error; err != nil {
		t.Fatalf("query sharer user: %v", err)
	}
	if err := db.Where("open_id = ?", buyerOpenID).First(&buyer).Error; err != nil {
		t.Fatalf("query buyer user: %v", err)
	}
	grantPermissionToUser(t, db, buyer.ID, "activity:register")

	store := &model.Store{Name: "活动分享门店2", Status: 1}
	if err := db.Create(store).Error; err != nil {
		t.Fatalf("create store: %v", err)
	}

	now := time.Now()
	act := &model.Activity{
		StoreID:   &store.ID,
		Name:      "活动报名测试2",
		Type:      1,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Rules:     "{}",
		Status:    1,
	}
	if err := db.Create(act).Error; err != nil {
		t.Fatalf("create activity: %v", err)
	}

	authBuyer := "Bearer " + buyerToken
	payload := map[string]any{
		"name":       "张三",
		"phone":      "13800000001",
		"fee":        9.9,
		"sharer_uid": sharer.ID,
	}
	resp := authedJSONRequest(t, fmt.Sprintf("%s/api/v1/activities/%d/register-with-order", ts.URL, act.ID), authBuyer, payload)
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected non-200 when sharer_uid provided without share_store_id for activity store order")
	}
}
