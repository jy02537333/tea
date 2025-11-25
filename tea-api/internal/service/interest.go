package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"strings"
	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"

	"github.com/shopspring/decimal"
)

type AccrualService struct {
	db *gorm.DB
}

func NewAccrualService() *AccrualService {
	return &AccrualService{db: database.GetDB()}
}

// Run 在指定日期按给定日利率为所有（或指定）用户计息
// rate 例如 0.001 表示 0.1%/日
func (s *AccrualService) Run(date time.Time, rate float64, userIDOptional *uint) (int, error) {
	if rate <= 0 {
		return 0, errors.New("rate must be positive")
	}
	d := normalizeDate(date)
	var users []model.User
	q := s.db.Model(&model.User{}).Where("status = ?", 1)
	if userIDOptional != nil {
		q = q.Where("id = ?", *userIDOptional)
	}
	if err := q.Find(&users).Error; err != nil {
		return 0, err
	}
	updated := 0
	for i := range users {
		if users[i].Balance.Cmp(decimal.Zero) <= 0 {
			continue
		}
		// 按用户定制利率覆盖
		r := rate
		if users[i].InterestRate.Cmp(decimal.Zero) > 0 {
			if f, ok := users[i].InterestRate.Float64(); ok {
				r = f
			} else {
				r = f
			}
		}
		if err := s.accrueUser(&users[i], d, r); err != nil {
			// 若重复或失败，继续下一个
			continue
		}
		updated++
	}
	return updated, nil
}

func (s *AccrualService) accrueUser(u *model.User, date time.Time, rate float64) error {
	// 幂等：同一天只记一次
	var exist int64
	if err := s.db.Model(&model.InterestRecord{}).
		Where("user_id = ? AND date(date) = date(?)", u.ID, date).
		Count(&exist).Error; err != nil {
		return err
	}
	if exist > 0 {
		return errors.New("already accrued")
	}

	principalBefore := u.Balance
	rateDec := decimal.NewFromFloat(rate)
	interest := principalBefore.Mul(rateDec).Round(2)
	if interest.Cmp(decimal.Zero) <= 0 {
		return errors.New("zero interest")
	}
	principalAfter := principalBefore.Add(interest).Round(2)

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 更新余额
		if err := tx.Model(&model.User{}).Where("id = ?", u.ID).Update("balance", principalAfter).Error; err != nil {
			return err
		}
		// 写入记录
		rec := &model.InterestRecord{
			BaseModel:       model.BaseModel{UID: utils.GenerateUID()},
			UserID:          u.ID,
			Date:            date,
			PrincipalBefore: principalBefore,
			Rate:            rateDec,
			InterestAmount:  interest,
			PrincipalAfter:  principalAfter,
			Method:          "daily",
		}
		if err := tx.Create(rec).Error; err != nil {
			// 若命中唯一约束，视为已计提（并发幂等）
			if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "constraint") {
				return errors.New("already accrued")
			}
			return err
		}
		return nil
	})
}

func normalizeDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// 保留接口占位（不再使用，逻辑迁移到 decimal.Round）
