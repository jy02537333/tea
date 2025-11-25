package service

import (
	"errors"
	"math"

	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

type StoreService struct{ db *gorm.DB }

func NewStoreService() *StoreService { return &StoreService{db: database.GetDB()} }

func (s *StoreService) CreateStore(st *model.Store) error {
	if st.Name == "" {
		return errors.New("门店名称必填")
	}
	return s.db.Create(st).Error
}

func (s *StoreService) UpdateStore(id uint, updates map[string]any) error {
	return s.db.Model(&model.Store{}).Where("id = ?", id).Updates(updates).Error
}

func (s *StoreService) DeleteStore(id uint) error {
	return s.db.Delete(&model.Store{}, id).Error
}

func (s *StoreService) GetStore(id uint) (*model.Store, error) {
	var st model.Store
	if err := s.db.First(&st, id).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

// ListStores 支持按经纬度计算距离并排序（在内存计算，便于 SQLite 测试环境）
func (s *StoreService) ListStores(page, limit int, status *int, lat, lng *float64) ([]map[string]any, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	q := s.db.Model(&model.Store{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.Store
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	res := make([]map[string]any, 0, len(list))
	for _, st := range list {
		item := map[string]any{
			"id":             st.ID,
			"name":           st.Name,
			"address":        st.Address,
			"phone":          st.Phone,
			"latitude":       st.Latitude,
			"longitude":      st.Longitude,
			"business_hours": st.BusinessHours,
			"images":         st.Images,
			"status":         st.Status,
		}
		if lat != nil && lng != nil && st.Latitude != 0 && st.Longitude != 0 {
			item["distance_km"] = haversine(*lat, *lng, st.Latitude, st.Longitude)
		}
		res = append(res, item)
	}

	// 简化：分页后不再排序；若需要严格距离排序，可在内存排序，但需取全量或扩大分页
	return res, total, nil
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // km
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return math.Round((R*c)*1000) / 1000 // 保留3位小数
}

func toRad(d float64) float64 { return d * math.Pi / 180.0 }
