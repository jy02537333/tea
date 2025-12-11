package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
)

// fakeOrderService implements orderService for testing AdminStoreOrders without touching the real DB.
type fakeOrderService struct {
	lastStoreID    uint
	lastStatus     int
	lastPage       int
	lastLimit      int
	lastStart      *time.Time
	lastEnd        *time.Time
	lastOrderID    uint
	ordersToReturn []model.Order
	totalToReturn  int64
	errToReturn    error
}

func (f *fakeOrderService) CreateOrderFromCart(userID uint, deliveryType int, addressInfo, remark string, userCouponID uint, storeID uint, orderType int) (*model.Order, error) {
	return nil, nil
}
func (f *fakeOrderService) ListOrders(userID uint, status int, page, limit int, storeID uint) ([]model.Order, int64, error) {
	return nil, 0, nil
}
func (f *fakeOrderService) GetOrder(userID, orderID uint) (*model.Order, []model.OrderItem, error) {
	return nil, nil, nil
}
func (f *fakeOrderService) AdminListOrders(status int, page, limit int, storeID uint) ([]model.Order, int64, error) {
	return nil, 0, nil
}
func (f *fakeOrderService) GetOrderAdmin(orderID uint) (*model.Order, []model.OrderItem, error) {
	return nil, nil, nil
}
func (f *fakeOrderService) CancelOrder(userID, orderID uint, reason string) error { return nil }
func (f *fakeOrderService) MarkPaid(userID, orderID uint) error                   { return nil }
func (f *fakeOrderService) StartDelivery(userID, orderID uint) error              { return nil }
func (f *fakeOrderService) Complete(userID, orderID uint) error                   { return nil }
func (f *fakeOrderService) Receive(userID, orderID uint) error                    { return nil }
func (f *fakeOrderService) AdminCancelOrder(orderID uint, reason string) error    { return nil }
func (f *fakeOrderService) AdminRefundOrder(orderID uint, reason string) error    { return nil }
func (f *fakeOrderService) AdminRefundStart(orderID uint, reason string) error    { return nil }
func (f *fakeOrderService) AdminRefundConfirm(orderID uint, reason string) error  { return nil }

func (f *fakeOrderService) AdminListStoreOrders(storeID uint, status int, page, limit int, startTime, endTime *time.Time, orderID uint) ([]model.Order, int64, error) {
	f.lastStoreID = storeID
	f.lastStatus = status
	f.lastPage = page
	f.lastLimit = limit
	f.lastStart = startTime
	f.lastEnd = endTime
	f.lastOrderID = orderID
	return f.ordersToReturn, f.totalToReturn, f.errToReturn
}

// TestAdminStoreOrders_ParsesParams verifies basic parameter parsing and delegation to service.
func TestAdminStoreOrders_ParsesParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fakeSvc := &fakeOrderService{
		ordersToReturn: []model.Order{{BaseModel: model.BaseModel{ID: 1}}},
		totalToReturn:  1,
	}

	h := &OrderHandler{svc: fakeSvc}
	r := gin.New()
	r.GET("/api/v1/admin/stores/:id/orders", h.AdminStoreOrders)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/stores/123/orders?page=2&page_size=50&status=4&order_id=99", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
	if fakeSvc.lastStoreID != 123 || fakeSvc.lastStatus != 4 || fakeSvc.lastPage != 2 || fakeSvc.lastLimit != 50 || fakeSvc.lastOrderID != 99 {
		t.Fatalf("unexpected service args: store=%d status=%d page=%d limit=%d orderID=%d", fakeSvc.lastStoreID, fakeSvc.lastStatus, fakeSvc.lastPage, fakeSvc.lastLimit, fakeSvc.lastOrderID)
	}
}
