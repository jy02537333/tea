package service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"

	"github.com/shopspring/decimal"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{db: database.GetDB()}
}

// LoginRequest 登录请求
// Support multiple login payloads:
// - {"code":"..."} for wx/code based login
// - {"openid":"..."} for dev openid login
// - {"username":"...","password":"...","captcha_id":"...","captcha_code":"..."} for dev UI username login
type LoginRequest struct {
	Code        string `json:"code" binding:"omitempty"`
	OpenID      string `json:"openid" binding:"omitempty"`
	Username    string `json:"username" binding:"omitempty"`
	Password    string `json:"password" binding:"omitempty"`
	CaptchaID   string `json:"captcha_id" binding:"omitempty"`
	CaptchaCode string `json:"captcha_code" binding:"omitempty"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string      `json:"token"`
	UserInfo interface{} `json:"user_info"`
}

// CreateAdminUserInput 管理端创建用户的入参
type CreateAdminUserInput struct {
	Username string
	Password string
	Phone    string
	Nickname string
	Role     string
	Status   int
}

// UserInfo 用户信息
type UserInfo struct {
	ID                      uint       `json:"id"`
	Username                string     `json:"username,omitempty"`
	UID                     string     `json:"uid"`
	OpenID                  string     `json:"open_id"`
	Nickname                string     `json:"nickname"`
	Avatar                  string     `json:"avatar"`
	Phone                   string     `json:"phone"`
	Gender                  int        `json:"gender"`
	Role                    string     `json:"role,omitempty"`
	StoreID                 *uint      `json:"store_id,omitempty"`
	Balance                 float64    `json:"balance"`
	Points                  int        `json:"points"`
	DefaultAddress          string     `json:"default_address"`
	DefaultAddressUpdatedAt *time.Time `json:"default_address_updated_at"`
}

type storeAdminRow struct {
	StoreID uint `gorm:"column:store_id"`
}

func (s *UserService) resolveStoreIDForStoreAdmin(userID uint) (*uint, error) {
	var rows []storeAdminRow
	if err := s.db.Table("store_admins").Select("store_id").Where("user_id = ?", userID).Order("id desc").Limit(1).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	sid := rows[0].StoreID
	if sid == 0 {
		return nil, nil
	}
	return &sid, nil
}

// Login 微信登录
func (s *UserService) Login(code string) (*LoginResponse, error) {
	// TODO: 调用微信API获取OpenID
	// 这里暂时模拟一个OpenID
	openID := "mock_open_id_" + code

	// 查找用户
	var user model.User
	err := s.db.Where("open_id = ?", openID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在，创建新用户
			genUID := utils.GenerateUID()
			// Phone 字段在 DB 中为 varchar(20)，避免把 32 字节 UID 写入导致插入失败，截断为不超过 20 字符
			phoneVal := genUID
			if len(phoneVal) > 20 {
				phoneVal = phoneVal[:20]
			}
			user = model.User{
				BaseModel: model.BaseModel{
					UID: genUID,
				},
				OpenID:   openID,
				Nickname: "微信用户",
				Status:   1,
				Balance:  decimalZero(),
				Points:   0,
				// 避免测试/开发环境 phone 唯一索引冲突
				Phone: phoneVal,
			}

			if err := s.db.Create(&user).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// 黑/白名单与停用状态拦截（白名单可豁免）
	if !user.IsWhitelisted {
		if user.Status == 2 {
			return nil, errors.New("账号已停用")
		}
		if user.IsBlacklisted {
			return nil, errors.New("账号已被加入黑名单")
		}
	}

	// 更新最后登录时间
	now := time.Now()
	s.db.Model(&user).Updates(map[string]interface{}{
		"last_login_at": &now,
	})

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.OpenID, user.Role)
	if err != nil {
		return nil, err
	}

	// 构造响应
	userInfo := UserInfo{
		ID:                      user.ID,
		Username:                "",
		UID:                     user.UID,
		OpenID:                  user.OpenID,
		Nickname:                user.Nickname,
		Avatar:                  user.Avatar,
		Phone:                   user.Phone,
		Gender:                  user.Gender,
		Role:                    user.Role,
		StoreID:                 nil,
		Balance:                 toFloat(user.Balance),
		Points:                  user.Points,
		DefaultAddress:          user.DefaultAddress,
		DefaultAddressUpdatedAt: user.DefaultAddressUpdatedAt,
	}
	if user.Username != nil {
		userInfo.Username = *user.Username
	}
	if user.Role == "store" {
		if sid, err := s.resolveStoreIDForStoreAdmin(user.ID); err == nil {
			userInfo.StoreID = sid
		}
	}

	return &LoginResponse{
		Token:    token,
		UserInfo: userInfo,
	}, nil
}

// LoginByOpenID 通过OpenID登录（本地/开发环境使用）
func (s *UserService) LoginByOpenID(openID string) (*LoginResponse, error) {
	if openID == "" {
		return nil, errors.New("openid 不能为空")
	}

	var user model.User
	err := s.db.Where("open_id = ?", openID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			genUID := utils.GenerateUID()
			phoneVal := genUID
			if len(phoneVal) > 20 {
				phoneVal = phoneVal[:20]
			}
			user = model.User{
				BaseModel: model.BaseModel{UID: genUID},
				OpenID:    openID,
				Nickname:  "微信用户",
				Status:    1,
				Balance:   decimalZero(),
				Points:    0,
				// 避免测试/开发环境 phone 唯一索引冲突
				Phone: phoneVal,
			}
			// 简单规则：特定OpenID赋予admin角色
			if openID == "admin_openid" {
				user.Role = "admin"
			} else {
				user.Role = "user"
			}
			if err := s.db.Create(&user).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// 已存在用户：若为admin_openid但角色非admin，自动提升
		if openID == "admin_openid" && user.Role != "admin" {
			user.Role = "admin"
			_ = s.db.Save(&user).Error
		}
	}

	// 黑/白名单与停用状态拦截（白名单可豁免）
	if !user.IsWhitelisted {
		if user.Status == 2 {
			return nil, errors.New("账号已停用")
		}
		if user.IsBlacklisted {
			return nil, errors.New("账号已被加入黑名单")
		}
	}

	// 更新时间
	now := time.Now()
	s.db.Model(&user).Updates(map[string]interface{}{"last_login_at": &now})

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.OpenID, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: token,
		UserInfo: UserInfo{
			ID:                      user.ID,
			Username:                func() string { if user.Username != nil { return *user.Username }; return "" }(),
			UID:                     user.UID,
			OpenID:                  user.OpenID,
			Nickname:                user.Nickname,
			Avatar:                  user.Avatar,
			Phone:                   user.Phone,
			Gender:                  user.Gender,
			Role:                    user.Role,
			StoreID:                 func() *uint { if user.Role != "store" { return nil }; sid, err := s.resolveStoreIDForStoreAdmin(user.ID); if err != nil { return nil }; return sid }(),
			Balance:                 toFloat(user.Balance),
			Points:                  user.Points,
			DefaultAddress:          user.DefaultAddress,
			DefaultAddressUpdatedAt: user.DefaultAddressUpdatedAt,
		},
	}, nil
}

// LoginByUsername 使用用户名或手机号 + 密码登录（生产/常规场景）
func (s *UserService) LoginByUsername(username, password string) (*LoginResponse, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password required")
	}

	var user model.User
	err := s.db.Where("username = ? OR phone = ?", username, username).First(&user).Error
	if err != nil {
		return nil, err
	}

	// Verify password hash
	if !utils.CheckPasswordHash(user.PasswordHash, password) {
		return nil, errors.New("invalid username or password")
	}

	// 黑/白名单与停用状态拦截（白名单可豁免）
	if !user.IsWhitelisted {
		if user.Status == 2 {
			return nil, errors.New("账号已停用")
		}
		if user.IsBlacklisted {
			return nil, errors.New("账号已被加入黑名单")
		}
	}

	// 更新最后登录时间
	now := time.Now()
	s.db.Model(&user).Updates(map[string]interface{}{"last_login_at": &now})

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.OpenID, user.Role)
	if err != nil {
		return nil, err
	}

	info := UserInfo{
		ID:                      user.ID,
		Username:                "",
		UID:                     user.UID,
		OpenID:                  user.OpenID,
		Nickname:                user.Nickname,
		Avatar:                  user.Avatar,
		Phone:                   user.Phone,
		Gender:                  user.Gender,
		Role:                    user.Role,
		StoreID:                 nil,
		Balance:                 toFloat(user.Balance),
		Points:                  user.Points,
		DefaultAddress:          user.DefaultAddress,
		DefaultAddressUpdatedAt: user.DefaultAddressUpdatedAt,
	}
	if user.Username != nil {
		info.Username = *user.Username
	}
	if user.Role == "store" {
		if sid, err := s.resolveStoreIDForStoreAdmin(user.ID); err == nil {
			info.StoreID = sid
		}
	}
	return &LoginResponse{Token: token, UserInfo: info}, nil

}

// GetUserModel returns raw user model (admin use) for consistent JSON shape.
func (s *UserService) GetUserModel(userID uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(userID uint) (*UserInfo, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	info := &UserInfo{
		ID:                      user.ID,
		Username:                "",
		UID:                     user.UID,
		OpenID:                  user.OpenID,
		Nickname:                user.Nickname,
		Avatar:                  user.Avatar,
		Phone:                   user.Phone,
		Gender:                  user.Gender,
		Role:                    user.Role,
		StoreID:                 nil,
		Balance:                 toFloat(user.Balance),
		Points:                  user.Points,
		DefaultAddress:          user.DefaultAddress,
		DefaultAddressUpdatedAt: user.DefaultAddressUpdatedAt,
	}
	if user.Username != nil {
		info.Username = *user.Username
	}
	if user.Role == "store" {
		if sid, err := s.resolveStoreIDForStoreAdmin(user.ID); err == nil {
			info.StoreID = sid
		}
	}
	return info, nil
}

// UpdateUserInfo 更新用户信息
func (s *UserService) UpdateUserInfo(userID uint, updates map[string]interface{}) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// UpdateDefaultAddress 保存用户默认地址
func (s *UserService) UpdateDefaultAddress(userID uint, address string) error {
	address = model.NormalizeJSONOrNull(address)
	updates := map[string]interface{}{
		"default_address":            address,
		"default_address_updated_at": time.Now(),
	}
	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// GetDefaultAddress 返回用户默认地址
func (s *UserService) GetDefaultAddress(userID uint) (string, *time.Time, error) {
	var user model.User
	if err := s.db.Select("default_address", "default_address_updated_at").First(&user, userID).Error; err != nil {
		return "", nil, err
	}
	return user.DefaultAddress, user.DefaultAddressUpdatedAt, nil
}

// GetUserByOpenID 根据OpenID获取用户
func (s *UserService) GetUserByOpenID(openID string) (*model.User, error) {
	var user model.User
	err := s.db.Where("open_id = ?", openID).First(&user).Error
	return &user, err
}

// helpers for decimal to preserve external JSON as float
func decimalZero() decimal.Decimal { return decimal.NewFromInt(0) }
func toFloat(d decimal.Decimal) float64 {
	f, _ := d.Float64()
	return f
}

// ListUsers 管理员获取用户列表（简化版）
func (s *UserService) ListUsers() ([]model.User, error) {
	var users []model.User
	if err := s.db.Order("id desc").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// ListUsersPaged 分页用户列表
func (s *UserService) ListUsersPaged(page, size int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	var total int64
	if err := s.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	if err := s.db.Order("id desc").Limit(size).Offset((page - 1) * size).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// ChangePassword 修改指定用户的密码（验证旧密码）
func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	if newPassword == "" {
		return errors.New("new password cannot be empty")
	}
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	// 如果已有密码，则校验旧密码
	if user.PasswordHash != "" {
		if !utils.CheckPasswordHash(user.PasswordHash, oldPassword) {
			return errors.New("old password incorrect")
		}
	}

	// Hash new password
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.db.Model(&user).Updates(map[string]interface{}{"password_hash": hashed}).Error
}

// ResetPasswordAdmin 允许管理员直接重置指定用户的密码
func (s *UserService) ResetPasswordAdmin(userID uint, newPassword string) error {
	if newPassword == "" {
		return errors.New("new password cannot be empty")
	}

	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password_hash": hashed,
	}).Error
}

// CreateAdminUser 管理端创建新用户（通常用于后台账号）
func (s *UserService) CreateAdminUser(input CreateAdminUserInput) (*model.User, error) {
	if input.Username == "" {
		return nil, errors.New("username 不能为空")
	}
	if input.Password == "" {
		return nil, errors.New("password 不能为空")
	}
	if len(input.Password) < 6 {
		return nil, errors.New("password 至少 6 位")
	}
	if input.Phone == "" {
		return nil, errors.New("phone 不能为空")
	}
	if input.Role == "" {
		input.Role = "user"
	}
	if input.Status == 0 {
		input.Status = 1
	}

	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	username := input.Username
	openID := fmt.Sprintf("manual_%s", utils.GenerateUID())
	user := model.User{
		Username:     &username,
		PasswordHash: hashed,
		Phone:        input.Phone,
		Nickname:     input.Nickname,
		Role:         input.Role,
		Status:       input.Status,
		OpenID:       openID,
	}
	if user.Nickname == "" {
		user.Nickname = input.Username
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
