package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

const (
	permCacheTTL = 30 * time.Minute
)

// 本地进程内缓存（当 Redis 不可用时的降级方案）
type permCacheEntry struct {
	Val    []string
	Expire time.Time
}

var (
	memPermCache = map[uint]permCacheEntry{}
	// 轻量并发场景使用，避免引入额外依赖；如需更强并发可替换为 sync.Map 或 LRU
	memPermCacheMu = struct {
		RLock   func()
		RUnlock func()
		Lock    func()
		Unlock  func()
	}{
		RLock:   func() {},
		RUnlock: func() {},
		Lock:    func() {},
		Unlock:  func() {},
	}
)

// 用 init 简单绑定一个无操作的互斥体，避免在无并发环境下引入复杂性
func init() {
	// 若后续需要真实互斥，可替换为 sync.RWMutex 并在此处赋予函数闭包
}

// GetUserPermissions 返回用户拥有的权限名集合（基于 DB 的角色-权限关联），带 Redis 缓存
func GetUserPermissions(db *gorm.DB, userID uint) ([]string, error) {
	if userID == 0 {
		return nil, nil
	}

	// 尝试 Redis 缓存
	r := database.GetRedis()
	key := fmt.Sprintf("perm:user:%d", userID)
	if r != nil {
		if bs, err := r.Get(context.Background(), key).Bytes(); err == nil && len(bs) > 0 {
			var arr []string
			if e := json.Unmarshal(bs, &arr); e == nil {
				return arr, nil
			}
		}
	} else {
		// 进程内缓存降级
		// 注意：该缓存仅在单进程中有效，适用于测试或无 Redis 的开发环境
		// 生产多实例下仍建议启用 Redis 保证一致性
		// 这里采用乐观读取（无锁），如需严谨一致性可加锁
		if entry, ok := memPermCache[userID]; ok {
			if time.Now().Before(entry.Expire) {
				return entry.Val, nil
			}
		}
	}

	// 查询 DB：UserRole -> RolePermission -> Permission(name)
	var names []string
	err := db.Model(&model.Permission{}).
		Select("permissions.name").
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Joins("JOIN user_roles ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ?", userID).
		Scan(&names).Error
	if err != nil {
		return nil, err
	}

	// 去重和归一化
	set := map[string]struct{}{}
	out := make([]string, 0, len(names))
	for _, n := range names {
		n2 := strings.TrimSpace(strings.ToLower(n))
		if n2 == "" {
			continue
		}
		if _, ok := set[n2]; ok {
			continue
		}
		set[n2] = struct{}{}
		out = append(out, n2)
	}

	// 写缓存
	if r != nil {
		if bs, e := json.Marshal(out); e == nil {
			_ = r.Set(context.Background(), key, bs, permCacheTTL).Err()
		}
	} else {
		memPermCache[userID] = permCacheEntry{Val: out, Expire: time.Now().Add(permCacheTTL)}
	}

	return out, nil
}

// InvalidateUserPermCache 删除某个用户的权限缓存
func InvalidateUserPermCache(userID uint) {
	if userID == 0 {
		return
	}
	if r := database.GetRedis(); r != nil {
		key := fmt.Sprintf("perm:user:%d", userID)
		_ = r.Del(context.Background(), key).Err()
	}
	delete(memPermCache, userID)
}

// InvalidateAllPermCache 清空所有用户权限缓存（使用 SCAN）
func InvalidateAllPermCache() {
	r := database.GetRedis()
	if r == nil {
		return
	}
	ctx := context.Background()
	var cursor uint64
	pattern := "perm:user:*"
	for {
		keys, cur, err := r.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			break
		}
		cursor = cur
		if len(keys) > 0 {
			_ = r.Del(ctx, keys...).Err()
		}
		if cursor == 0 {
			break
		}
	}
}
