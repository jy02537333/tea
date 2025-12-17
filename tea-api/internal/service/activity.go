package service

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

type ActivityService struct{ db *gorm.DB }

func NewActivityService() *ActivityService { return &ActivityService{db: database.GetDB()} }

// CreateActivity 平台/门店通用的活动创建逻辑
func (s *ActivityService) CreateActivity(a *model.Activity) error {
	if a.Name == "" {
		return errors.New("活动名称必填")
	}
	if a.Type < 1 || a.Type > 3 {
		return errors.New("非法的活动类型")
	}
	if a.EndTime.Before(a.StartTime) {
		return errors.New("结束时间需大于开始时间")
	}
	return s.db.Create(a).Error
}

// CreateStoreActivity 创建门店活动（强制绑定门店ID）
func (s *ActivityService) CreateStoreActivity(storeID uint, a *model.Activity) error {
	if storeID == 0 {
		return errors.New("无效的门店ID")
	}
	a.StoreID = &storeID
	return s.CreateActivity(a)
}

// ListStoreActivities 按门店列出活动，支持按状态过滤（简化：不分页）
func (s *ActivityService) ListStoreActivities(storeID uint, status int) ([]model.Activity, error) {
	if storeID == 0 {
		return nil, errors.New("无效的门店ID")
	}
	var list []model.Activity
	q := s.db.Model(&model.Activity{}).Where("store_id = ?", storeID)
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// UpdateStoreActivity 门店编辑自己的活动
func (s *ActivityService) UpdateStoreActivity(storeID, activityID uint, upd *model.Activity) (*model.Activity, error) {
	if storeID == 0 || activityID == 0 {
		return nil, errors.New("无效的参数")
	}
	var act model.Activity
	if err := s.db.First(&act, activityID).Error; err != nil {
		return nil, err
	}
	if act.StoreID == nil || *act.StoreID != storeID {
		return nil, errors.New("无权操作该活动")
	}

	// 可编辑字段
	act.Name = upd.Name
	act.Type = upd.Type
	act.StartTime = upd.StartTime
	act.EndTime = upd.EndTime
	act.Rules = upd.Rules
	act.Status = upd.Status
	act.Priority = upd.Priority
	act.Description = upd.Description

	// 复用创建时的基础校验
	if act.Name == "" {
		return nil, errors.New("活动名称必填")
	}
	if act.Type < 1 || act.Type > 3 {
		return nil, errors.New("非法的活动类型")
	}
	if act.EndTime.Before(act.StartTime) {
		return nil, errors.New("结束时间需大于开始时间")
	}
	if err := s.db.Save(&act).Error; err != nil {
		return nil, err
	}
	return &act, nil
}

// ListUserActivities 用户侧：按门店列出当前可报名活动（启用且在时间范围内）
func (s *ActivityService) ListUserActivities(storeID uint) ([]model.Activity, error) {
	if storeID == 0 {
		return nil, errors.New("无效的门店ID")
	}
	var list []model.Activity
	now := time.Now()
	q := s.db.Model(&model.Activity{}).
		Where("store_id = ? AND status = 1 AND start_time <= ? AND end_time >= ?", storeID, now, now).
		Order("priority desc, id desc")
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ListActivityRegistrations 按门店+活动分页查询报名记录，支持按状态过滤
func (s *ActivityService) ListActivityRegistrations(storeID, activityID uint, page, limit int, status *int) ([]model.ActivityRegistration, int64, error) {
	if storeID == 0 || activityID == 0 {
		return nil, 0, errors.New("无效的门店或活动ID")
	}
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}

	q := s.db.Model(&model.ActivityRegistration{}).
		Select("activity_registrations.*, orders.status AS order_status, orders.pay_status AS order_pay_status").
		Joins("LEFT JOIN orders ON orders.id = activity_registrations.order_id").
		Where("activity_registrations.store_id = ? AND activity_registrations.activity_id = ?", storeID, activityID)
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ActivityRegistration
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// RefundActivityRegistration 简化版退款：仅更新报名记录状态与退款信息，实际打款由线下/其他系统处理
func (s *ActivityService) RefundActivityRegistration(storeID, activityID, regID uint, reason string, operatorUserID uint) (*model.ActivityRegistration, error) {
	if storeID == 0 || activityID == 0 || regID == 0 {
		return nil, errors.New("无效的参数")
	}
	var reg model.ActivityRegistration
	if err := s.db.First(&reg, regID).Error; err != nil {
		return nil, err
	}
	if reg.StoreID != storeID || reg.ActivityID != activityID {
		return nil, errors.New("无权操作该报名记录")
	}
	if reg.Status == 3 {
		return nil, errors.New("该报名已退款")
	}

	now := time.Now()
	reg.Status = 3
	// 简化：默认退款金额 = 报名费用，后续可支持部分退款
	reg.RefundAmount = reg.Fee
	reg.RefundReason = reason
	reg.RefundedAt = &now
	// 如需记录操作人，可在后续为 ActivityRegistration 增加 processed_by 字段
	if err := s.db.Save(&reg).Error; err != nil {
		return nil, err
	}
	return &reg, nil
}

// RegisterActivityWithOrder 用户报名并生成订单（用于需要支付的活动报名场景）
func (s *ActivityService) RegisterActivityWithOrder(activityID, userID uint, name, phone string, fee decimal.Decimal) (*model.ActivityRegistration, *model.Order, error) {
	if activityID == 0 || userID == 0 {
		return nil, nil, errors.New("无效的活动或用户ID")
	}
	if fee.LessThan(decimal.NewFromInt(0)) {
		return nil, nil, errors.New("报名费用不能为负数")
	}
	var act model.Activity
	if err := s.db.First(&act, activityID).Error; err != nil {
		return nil, nil, err
	}
	if act.Status != 1 {
		return nil, nil, errors.New("活动未启用或已下线")
	}
	now := time.Now()
	if now.Before(act.StartTime) || now.After(act.EndTime) {
		return nil, nil, errors.New("当前不在活动报名时间内")
	}
	if act.StoreID == nil || *act.StoreID == 0 {
		return nil, nil, errors.New("当前活动暂不支持报名")
	}

	// 防止重复报名：同一用户对同一活动仅允许一条“已报名”记录
	var existing model.ActivityRegistration
	if err := s.db.Where("activity_id = ? AND user_id = ? AND status = 1", activityID, userID).First(&existing).Error; err == nil {
		return nil, nil, errors.New("您已报名该活动")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	}

	var reg model.ActivityRegistration
	var order model.Order
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		// 创建一个简化订单，金额即为报名费用
		order = model.Order{
			OrderNo:        "", // 由数据库或后续逻辑填充，如需可接入统一订单号生成器
			UserID:         userID,
			StoreID:        *act.StoreID,
			TotalAmount:    fee,
			PayAmount:      fee,
			DiscountAmount: decimal.NewFromInt(0),
			DeliveryFee:    decimal.NewFromInt(0),
			Status:         1,
			PayStatus:      1,
			OrderType:      1,
			DeliveryType:   1,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		reg = model.ActivityRegistration{
			StoreID:      *act.StoreID,
			ActivityID:   activityID,
			UserID:       userID,
			UserName:     name,
			UserPhone:    phone,
			Status:       1,
			Fee:          fee,
			OrderID:      &order.ID,
			RefundAmount: decimal.NewFromInt(0),
		}
		if err := tx.Create(&reg).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}
	return &reg, &order, nil
}

// RegisterActivity 用户报名活动（简化版：仅记录报名信息，不直接创建订单/支付流水）
func (s *ActivityService) RegisterActivity(activityID, userID uint, name, phone string) (*model.ActivityRegistration, error) {
	if activityID == 0 || userID == 0 {
		return nil, errors.New("无效的活动或用户ID")
	}
	var act model.Activity
	if err := s.db.First(&act, activityID).Error; err != nil {
		return nil, err
	}
	if act.Status != 1 {
		return nil, errors.New("活动未启用或已下线")
	}
	now := time.Now()
	if now.Before(act.StartTime) || now.After(act.EndTime) {
		return nil, errors.New("当前不在活动报名时间内")
	}
	if act.StoreID == nil || *act.StoreID == 0 {
		return nil, errors.New("当前活动暂不支持报名")
	}

	// 防止重复报名：同一用户对同一活动仅允许一条“已报名”记录
	var existing model.ActivityRegistration
	if err := s.db.Where("activity_id = ? AND user_id = ? AND status = 1", activityID, userID).First(&existing).Error; err == nil {
		return nil, errors.New("您已报名该活动")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	reg := &model.ActivityRegistration{
		StoreID:    *act.StoreID,
		ActivityID: activityID,
		UserID:     userID,
		UserName:   name,
		UserPhone:  phone,
		Status:     1,
		Fee:        decimal.NewFromInt(0),
	}
	if err := s.db.Create(reg).Error; err != nil {
		return nil, err
	}
	return reg, nil
}
