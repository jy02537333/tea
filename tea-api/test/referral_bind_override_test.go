package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tea-api/internal/config"
	"tea-api/internal/router"
	"tea-api/pkg/database"
)

func Test_Referral_Bind_LastClickWins(t *testing.T) {
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()
	db := database.GetDB()
	if db == nil {
		t.Fatalf("db is nil")
	}

	// Some environments might not have referrals_closure table migrated.
	_ = db.Exec(`
CREATE TABLE IF NOT EXISTS referral_closures (
  ancestor_user_id BIGINT NOT NULL,
  descendant_user_id BIGINT NOT NULL,
  depth INT NOT NULL,
  PRIMARY KEY (ancestor_user_id, descendant_user_id)
)`).Error

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	login := func(openid string) string {
		b, _ := json.Marshal(map[string]string{"openid": openid})
		resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatalf("dev-login err: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("dev-login status: %d", resp.StatusCode)
		}
		var out struct {
			Code int `json:"code"`
			Data struct {
				Token string `json:"token"`
			} `json:"data"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&out)
		if out.Code != 0 || out.Data.Token == "" {
			t.Fatalf("login failed: %+v", out)
		}
		return "Bearer " + out.Data.Token
	}

	ref1OpenID := "ref_bind_ref1_openid"
	ref2OpenID := "ref_bind_ref2_openid"
	buyerOpenID := "ref_bind_buyer_openid"

	_ = login(ref1OpenID)
	_ = login(ref2OpenID)
	buyerAuth := login(buyerOpenID)

	getUserID := func(openid string) uint {
		var id uint
		if err := db.Table("users").Select("id").Where("open_id = ?", openid).Scan(&id).Error; err != nil {
			t.Fatalf("query user id err: %v", err)
		}
		if id == 0 {
			t.Fatalf("user id is 0 for openid=%s", openid)
		}
		return id
	}

	ref1ID := getUserID(ref1OpenID)
	ref2ID := getUserID(ref2OpenID)
	buyerID := getUserID(buyerOpenID)

	// Cleanup historical data
	_ = db.Exec("DELETE FROM referral_closures WHERE descendant_user_id = ? AND depth = 1", buyerID).Error

	bind := func(referrerID uint) {
		body, _ := json.Marshal(map[string]any{"referrer_id": referrerID})
		req, _ := http.NewRequest("POST", ts.URL+"/api/v1/referrals/bind", bytes.NewReader(body))
		req.Header.Set("Authorization", buyerAuth)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("bind request err: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("bind status: %d", resp.StatusCode)
		}
		var out struct {
			Code int `json:"code"`
			Data struct {
				Bound bool `json:"bound"`
			} `json:"data"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&out)
		if out.Code != 0 || !out.Data.Bound {
			t.Fatalf("bind failed: %+v", out)
		}
	}

	bind(ref1ID)
	bind(ref2ID)

	// Verify only the last referrer is kept as depth=1
	type row struct {
		AncestorUserID   uint `gorm:"column:ancestor_user_id"`
		DescendantUserID uint `gorm:"column:descendant_user_id"`
		Depth            int  `gorm:"column:depth"`
	}
	var rows []row
	if err := db.Table("referral_closures").
		Select("ancestor_user_id, descendant_user_id, depth").
		Where("descendant_user_id = ? AND depth = 1", buyerID).
		Find(&rows).Error; err != nil {
		t.Fatalf("query referral_closures err: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 depth=1 row, got %d: %+v", len(rows), rows)
	}
	if rows[0].AncestorUserID != ref2ID {
		t.Fatalf("expected last referrer=%d, got %d", ref2ID, rows[0].AncestorUserID)
	}
}
