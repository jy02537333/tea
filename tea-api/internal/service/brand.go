package service

import (
	"fmt"

	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

type BrandService struct{ db *gorm.DB }

func NewBrandService() *BrandService { return &BrandService{db: database.GetDB()} }

// CreateBrand 创建品牌
func (s *BrandService) CreateBrand(b *model.Brand) error {
    if b.Name == "" {
        return fmt.Errorf("品牌名称必填")
    }
    return s.db.Create(b).Error
}

// GetBrands 获取品牌列表（简单分页/搜索可后续扩展）
func (s *BrandService) GetBrands(page, limit int, q string) ([]model.Brand, int64, error) {
    var list []model.Brand
    var total int64
    query := s.db.Model(&model.Brand{})
    if q != "" {
        like := "%" + q + "%"
        query = query.Where("name LIKE ? OR description LIKE ?", like, like)
    }
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    offset := (page - 1) * limit
    if err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
        return nil, 0, err
    }
    return list, total, nil
}

// GetBrand 获取单个品牌
func (s *BrandService) GetBrand(id uint) (*model.Brand, error) {
    var b model.Brand
    if err := s.db.First(&b, id).Error; err != nil {
        return nil, err
    }
    return &b, nil
}

// UpdateBrand 更新品牌
func (s *BrandService) UpdateBrand(id uint, updates map[string]any) error {
    return s.db.Model(&model.Brand{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteBrand 删除品牌（若被商品引用则拒绝）
func (s *BrandService) DeleteBrand(id uint) error {
    var cnt int64
    if err := s.db.Model(&model.Product{}).Where("brand_id = ?", id).Count(&cnt).Error; err != nil {
        return err
    }
    if cnt > 0 {
        return fmt.Errorf("该品牌已被 %d 个商品引用，无法删除", cnt)
    }
    return s.db.Delete(&model.Brand{}, id).Error
}
