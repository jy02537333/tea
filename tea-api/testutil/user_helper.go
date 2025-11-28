package testutil

import (
    "testing"

    "github.com/shopspring/decimal"

    "tea-api/internal/model"
    "tea-api/pkg/database"
    "tea-api/pkg/utils"
)

// CreateTestUser creates and returns a unique test user inserted into DB.
// It fatals on error. Ensures `phone` and `username` conform to schema.
func CreateTestUser(t testing.TB) *model.User {
    t.Helper()
    db := database.GetDB()
    if db == nil {
        t.Fatalf("database not initialized")
    }

    username := "user_" + utils.GenerateUID()
    phone := "p_" + utils.GenerateUID()
    if len(phone) > 20 {
        phone = phone[:20]
    }

    u := &model.User{
        BaseModel: model.BaseModel{UID: utils.GenerateUID()},
        Username:  username,
        OpenID:    "u_" + utils.GenerateUID(),
        Phone:     phone,
        Nickname:  "test",
        Status:    1,
        Balance:   decimal.NewFromFloat(1000),
    }

    if err := db.Create(u).Error; err != nil {
        t.Fatalf("create test user: %v", err)
    }
    return u
}
