package service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"tea-api/internal/model"
)

func applyShareAttributionToOrder(tx *gorm.DB, order *model.Order, userID uint, storeID uint, sharerUID uint, shareStoreID uint) error {
	// 仅当下单接口携带分享参数时，才冻结推荐人/分享门店；否则不计分享归属。
	// 参数一致性：share_store_id 单独出现没有意义，直接拒绝。
	if shareStoreID != 0 && sharerUID == 0 {
		return errors.New("缺少分享人")
	}
	if sharerUID == 0 {
		// 未携带分享参数：不冻结推荐人/分享门店（不计分享佣金/业绩归属）
		order.ReferrerID = nil
		order.ShareStoreID = 0
		return nil
	}

	if sharerUID == userID {
		return errors.New("非法的分享人")
	}

	// 门店分享强校验：门店订单必须提供 share_store_id，且需与订单 store_id 一致。
	// 商城/会员等 store_id=0 的订单不允许携带 share_store_id。
	if storeID == 0 {
		if shareStoreID != 0 {
			return errors.New("分享门店不匹配")
		}
		order.ShareStoreID = 0
	} else {
		if shareStoreID == 0 {
			return errors.New("缺少分享门店")
		}
		if shareStoreID != storeID {
			return errors.New("分享门店不匹配")
		}
		order.ShareStoreID = shareStoreID
	}

	// 校验分享人存在
	var refUser model.User
	if err := tx.First(&refUser, sharerUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("分享人不存在")
		}
		return err
	}

	// 若已存在直推，则必须与 sharer_uid 一致；若不存在直推，则在此自动绑定一次
	var rc model.ReferralClosure
	err := tx.Where("descendant_user_id = ? AND depth = 1", userID).First(&rc).Error
	if err == nil {
		if rc.AncestorUserID != sharerUID {
			return errors.New("分享人不匹配")
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 确保自闭包 depth=0 存在（忽略并发/重复）
		var self model.ReferralClosure
		_ = tx.Where("ancestor_user_id = ? AND descendant_user_id = ? AND depth = 0", userID, userID).First(&self).Error
		if self.AncestorUserID == 0 {
			_ = tx.Create(&model.ReferralClosure{AncestorUserID: userID, DescendantUserID: userID, Depth: 0}).Error
		}
		if err := tx.Create(&model.ReferralClosure{AncestorUserID: sharerUID, DescendantUserID: userID, Depth: 1}).Error; err != nil {
			return fmt.Errorf("绑定分享人失败: %w", err)
		}
	} else {
		return err
	}

	id := sharerUID
	order.ReferrerID = &id
	return nil
}
