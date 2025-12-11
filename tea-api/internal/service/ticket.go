package service

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

type TicketService struct{ db *gorm.DB }

func NewTicketService() *TicketService { return &TicketService{db: database.GetDB()} }

type TicketListResult struct {
	List  []model.Ticket
	Total int64
}

// ListTickets 按多条件分页查询工单
func (s *TicketService) ListTickets(page, limit int, status, ticketType, source, priority, keyword string, storeID, userID *uint) ([]model.Ticket, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}

	q := s.db.Model(&model.Ticket{})
	if status = strings.TrimSpace(status); status != "" {
		q = q.Where("status = ?", status)
	}
	if ticketType = strings.TrimSpace(ticketType); ticketType != "" {
		q = q.Where("type = ?", ticketType)
	}
	if source = strings.TrimSpace(source); source != "" {
		q = q.Where("source = ?", source)
	}
	if priority = strings.TrimSpace(priority); priority != "" {
		q = q.Where("priority = ?", priority)
	}
	if storeID != nil && *storeID > 0 {
		q = q.Where("store_id = ?", *storeID)
	}
	if userID != nil && *userID > 0 {
		q = q.Where("user_id = ?", *userID)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("title LIKE ? OR content LIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.Ticket
	if err := q.Order("id DESC").Limit(limit).Offset((page - 1) * limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetTicket 根据ID获取工单详情
func (s *TicketService) GetTicket(id uint) (*model.Ticket, error) {
	if id == 0 {
		return nil, errors.New("无效的工单ID")
	}
	var t model.Ticket
	if err := s.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateTicket 创建工单（后台人工录入或系统入口统一落库）
func (s *TicketService) CreateTicket(t *model.Ticket) error {
	if t == nil {
		return errors.New("工单不能为空")
	}
	if strings.TrimSpace(t.Type) == "" || strings.TrimSpace(t.Source) == "" || strings.TrimSpace(t.Title) == "" {
		return errors.New("type/source/title 为必填字段")
	}
	if t.Status == "" {
		t.Status = "new"
	}
	if t.Priority == "" {
		t.Priority = "normal"
	}
	return s.db.Create(t).Error
}

type UpdateTicketInput struct {
	Status       *string
	Priority     *string
	Remark       *string
	RejectReason *string
	AssigneeID   *uint
}

// UpdateTicket 更新工单状态/负责人等信息
func (s *TicketService) UpdateTicket(id uint, input UpdateTicketInput, operatorID uint) error {
	if id == 0 {
		return errors.New("无效的工单ID")
	}
	var t model.Ticket
	if err := s.db.First(&t, id).Error; err != nil {
		return err
	}

	updates := map[string]any{
		"updated_at": time.Now(),
	}
	if operatorID != 0 {
		updates["updated_by"] = operatorID
	}

	if input.Status != nil && strings.TrimSpace(*input.Status) != "" {
		status := strings.TrimSpace(*input.Status)
		updates["status"] = status
		now := time.Now()
		if status == "resolved" {
			updates["resolved_at"] = now
		} else if status == "closed" {
			updates["closed_at"] = now
		}
	}
	if input.Priority != nil && strings.TrimSpace(*input.Priority) != "" {
		updates["priority"] = strings.TrimSpace(*input.Priority)
	}
	if input.Remark != nil {
		updates["remark"] = strings.TrimSpace(*input.Remark)
	}
	if input.RejectReason != nil {
		updates["reject_reason"] = strings.TrimSpace(*input.RejectReason)
	}
	if input.AssigneeID != nil {
		updates["assignee_id"] = *input.AssigneeID
	}

	if len(updates) == 1 { // 只有 updated_at
		return nil
	}
	return s.db.Model(&model.Ticket{}).Where("id = ?", id).Updates(updates).Error
}
