package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/response"
)

// ContentHandler serves public content pages such as privacy/about/terms
type ContentHandler struct {
    db *gorm.DB
}

func NewContentHandler() *ContentHandler {
    return &ContentHandler{db: database.GetDB()}
}

// GET /api/v1/content/pages?keys=content_about,content_privacy
// or   /api/v1/content/pages?slugs=about,privacy,terms
// Returns: { items: [{ key, value }] }
func (h *ContentHandler) GetPages(c *gin.Context) {
    if h.db == nil {
        response.Error(c, http.StatusInternalServerError, "db not ready")
        return
    }

    keysParam := strings.TrimSpace(c.Query("keys"))
    slugsParam := strings.TrimSpace(c.Query("slugs"))

    var keys []string
    if keysParam != "" {
        for _, k := range strings.Split(keysParam, ",") {
            k = strings.TrimSpace(k)
            if k != "" {
                keys = append(keys, k)
            }
        }
    } else if slugsParam != "" {
        for _, s := range strings.Split(slugsParam, ",") {
            s = strings.TrimSpace(s)
            if s == "" {
                continue
            }
            switch s {
            case "about":
                keys = append(keys, "content_about")
            case "help":
                keys = append(keys, "content_help")
            case "privacy":
                keys = append(keys, "content_privacy")
            case "terms", "agreement":
                keys = append(keys, "content_terms")
            default:
                // ignore unknown slug
            }
        }
    }

    if len(keys) == 0 {
        response.BadRequest(c, "missing keys or slugs")
        return
    }

    var rows []model.SystemConfig
    if err := h.db.Model(&model.SystemConfig{}).
        Where("config_key IN ?", keys).
        Order("id ASC").
        Find(&rows).Error; err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    items := make([]gin.H, 0, len(rows))
    for _, r := range rows {
        items = append(items, gin.H{"key": r.ConfigKey, "value": r.ConfigValue})
    }
    response.Success(c, gin.H{"items": items})
}
