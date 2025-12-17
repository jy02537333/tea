package service

import (
	"errors"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

type StoreInventoryService struct{ db *gorm.DB }

func NewStoreInventoryService() *StoreInventoryService {
	return &StoreInventoryService{db: database.GetDB()}
}

func (s *StoreInventoryService) Upsert(storeID, productID uint, stock int, priceOverride string, bizType int) (*model.StoreProduct, error) {
	if storeID == 0 || productID == 0 {
		return nil, errors.New("参数错误")
	}
	if bizType <= 0 {
		bizType = 1
	}
	var d decimal.Decimal
	if priceOverride != "" {
		var err error
		d, err = decimal.NewFromString(priceOverride)
		if err != nil {
			return nil, errors.New("price_override 非法")
		}
	} else {
		d = decimal.Zero
	}
	sp := model.StoreProduct{StoreID: storeID, ProductID: productID, Stock: stock, PriceOverride: d, BizType: bizType}
	// 使用原子 Upsert，避免多步事务在 SQLite 上造成锁竞争
	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "store_id"}, {Name: "product_id"}},
		DoUpdates: clause.Assignments(map[string]any{"stock": stock, "price_override": d, "biz_type": bizType}),
	}).Create(&sp).Error; err != nil {
		return nil, err
	}
	// 返回当前记录（可能是更新后的）
	var out model.StoreProduct
	if err := s.db.Where("store_id = ? AND product_id = ?", storeID, productID).First(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *StoreInventoryService) List(storeID uint, page, limit int, bizType *int) ([]model.StoreProduct, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	q := s.db.Model(&model.StoreProduct{}).Where("store_id = ?", storeID)
	if bizType != nil && *bizType > 0 {
		q = q.Where("biz_type = ?", *bizType)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.StoreProduct
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *StoreInventoryService) Delete(storeID, productID uint) error {
	return s.db.Where("store_id = ? AND product_id = ?", storeID, productID).Delete(&model.StoreProduct{}).Error
}
