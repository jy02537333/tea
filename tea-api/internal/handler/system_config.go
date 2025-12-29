package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/response"
	"tea-api/pkg/utils"
)

type SystemConfigHandler struct {
	db *gorm.DB
}

func NewSystemConfigHandler() *SystemConfigHandler {
	return &SystemConfigHandler{db: database.GetDB()}
}

type upsertSystemConfigsReq struct {
	Items []struct {
		ConfigKey   string `json:"config_key" binding:"required"`
		ConfigValue string `json:"config_value"`
		ConfigType  string `json:"config_type"`
		Description string `json:"description"`
		Status      int    `json:"status"`
	} `json:"items" binding:"required"`
}

// List lists system configs.
// GET /api/v1/admin/system/configs?keys=a,b,c&prefix=site_
func (h *SystemConfigHandler) List(c *gin.Context) {
	if h.db == nil {
		response.Error(c, http.StatusInternalServerError, "db not ready")
		return
	}

	keysParam := strings.TrimSpace(c.Query("keys"))
	prefix := strings.TrimSpace(c.Query("prefix"))

	q := h.db.Model(&model.SystemConfig{})
	if keysParam != "" {
		raw := strings.Split(keysParam, ",")
		keys := make([]string, 0, len(raw))
		for _, k := range raw {
			k = strings.TrimSpace(k)
			if k != "" {
				keys = append(keys, k)
			}
		}
		if len(keys) > 0 {
			q = q.Where("config_key IN ?", keys)
		}
	} else if prefix != "" {
		q = q.Where("config_key LIKE ?", prefix+"%")
	}

	var list []model.SystemConfig
	if err := q.Order("id ASC").Find(&list).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"list": list})
}

// UpsertMany upserts system configs.
// PUT /api/v1/admin/system/configs
func (h *SystemConfigHandler) UpsertMany(c *gin.Context) {
	if h.db == nil {
		response.Error(c, http.StatusInternalServerError, "db not ready")
		return
	}

	var req upsertSystemConfigsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	updated := 0
	err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Items {
			key := strings.TrimSpace(item.ConfigKey)
			if key == "" {
				continue
			}

			var cfg model.SystemConfig
			err := tx.Where("config_key = ?", key).First(&cfg).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					cfg = model.SystemConfig{
						BaseModel:    model.BaseModel{UID: utils.GenerateUID()},
						ConfigKey:    key,
						ConfigValue:  item.ConfigValue,
						ConfigType:   strings.TrimSpace(item.ConfigType),
						Description:  strings.TrimSpace(item.Description),
						Status:       item.Status,
					}
					if cfg.ConfigType == "" {
						cfg.ConfigType = "string"
					}
					if cfg.Status == 0 {
						cfg.Status = 1
					}
					if e := tx.Create(&cfg).Error; e != nil {
						return e
					}
					updated++
					continue
				}
				return err
			}

			updates := map[string]any{
				"config_value": strings.TrimSpace(item.ConfigValue),
			}
			if strings.TrimSpace(item.ConfigType) != "" {
				updates["config_type"] = strings.TrimSpace(item.ConfigType)
			}
			if strings.TrimSpace(item.Description) != "" {
				updates["description"] = strings.TrimSpace(item.Description)
			}
			if item.Status != 0 {
				updates["status"] = item.Status
			}

			if e := tx.Model(&model.SystemConfig{}).Where("id = ?", cfg.ID).Updates(updates).Error; e != nil {
				return e
			}
			updated++
		}
		return nil
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"ok": true, "updated": updated})
}
