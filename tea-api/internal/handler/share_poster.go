package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/response"
	"tea-api/pkg/utils"
)

const sharePosterTemplatesConfigKey = "share_poster_templates"

type SharePosterTemplate struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"` // 1 enabled, 0 disabled
}

type SharePosterHandler struct {
	db *gorm.DB
}

func NewSharePosterHandler() *SharePosterHandler {
	return &SharePosterHandler{db: database.GetDB()}
}

// ListPublic lists enabled share poster templates.
// GET /api/v1/share/posters
func (h *SharePosterHandler) ListPublic(c *gin.Context) {
	list, err := h.loadTemplates()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]SharePosterTemplate, 0, len(list))
	for _, it := range list {
		if it.Status == 1 {
			out = append(out, it)
		}
	}
	response.Success(c, gin.H{"list": out})
}

// AdminGet lists all share poster templates (including disabled).
// GET /api/v1/admin/system/share-posters
func (h *SharePosterHandler) AdminGet(c *gin.Context) {
	list, err := h.loadTemplates()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"list": list})
}

type upsertSharePosterTemplatesReq struct {
	List []SharePosterTemplate `json:"list"`
}

// AdminPut replaces templates.
// PUT /api/v1/admin/system/share-posters
func (h *SharePosterHandler) AdminPut(c *gin.Context) {
	if h.db == nil {
		response.Error(c, http.StatusInternalServerError, "db not ready")
		return
	}

	var req upsertSharePosterTemplatesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	clean := make([]SharePosterTemplate, 0, len(req.List))
	for _, it := range req.List {
		it.ID = strings.TrimSpace(it.ID)
		it.Title = strings.TrimSpace(it.Title)
		it.ImageURL = strings.TrimSpace(it.ImageURL)
		if it.ImageURL == "" {
			response.BadRequest(c, "海报图片不能为空")
			return
		}
		if it.ID == "" {
			it.ID = utils.GenerateUID()
		}
		if it.Status != 0 {
			it.Status = 1
		}
		clean = append(clean, it)
	}

	sort.SliceStable(clean, func(i, j int) bool {
		if clean[i].Sort != clean[j].Sort {
			return clean[i].Sort < clean[j].Sort
		}
		return clean[i].ID < clean[j].ID
	})

	b, err := json.Marshal(clean)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "序列化失败")
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		var cfg model.SystemConfig
		e := tx.Where("config_key = ?", sharePosterTemplatesConfigKey).First(&cfg).Error
		if e != nil {
			if e == gorm.ErrRecordNotFound {
				cfg = model.SystemConfig{
					BaseModel:    model.BaseModel{UID: utils.GenerateUID()},
					ConfigKey:    sharePosterTemplatesConfigKey,
					ConfigValue:  string(b),
					ConfigType:   "json",
					Description:  "share poster templates",
					Status:       1,
				}
				return tx.Create(&cfg).Error
			}
			return e
		}

		return tx.Model(&model.SystemConfig{}).Where("id = ?", cfg.ID).Updates(map[string]any{
			"config_value": string(b),
			"config_type":  "json",
			"status":       1,
			"description":  "share poster templates",
		}).Error
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"ok": true, "count": len(clean)})
}

func (h *SharePosterHandler) loadTemplates() ([]SharePosterTemplate, error) {
	if h.db == nil {
		return nil, gorm.ErrInvalidDB
	}

	var cfg model.SystemConfig
	if err := h.db.Where("config_key = ? AND status = 1", sharePosterTemplatesConfigKey).First(&cfg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []SharePosterTemplate{}, nil
		}
		return nil, err
	}
	raw := strings.TrimSpace(cfg.ConfigValue)
	if raw == "" {
		return []SharePosterTemplate{}, nil
	}

	var list []SharePosterTemplate
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		// tolerate bad data without breaking the user page
		return []SharePosterTemplate{}, nil
	}
	for i := range list {
		list[i].ID = strings.TrimSpace(list[i].ID)
		list[i].Title = strings.TrimSpace(list[i].Title)
		list[i].ImageURL = strings.TrimSpace(list[i].ImageURL)
		if list[i].Status != 0 {
			list[i].Status = 1
		}
	}
	// stable sort
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Sort != list[j].Sort {
			return list[i].Sort < list[j].Sort
		}
		return list[i].ID < list[j].ID
	})
	return list, nil
}
