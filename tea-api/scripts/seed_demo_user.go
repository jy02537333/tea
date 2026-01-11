//go:build tools
// +build tools

package main

import (
	"fmt"
	"log"
	"time"

	"tea-api/internal/config"
	"tea-api/pkg/database"
)

// Seed a demo user and related records for E2E summary aggregation.
func main() {
	// Load config to populate database connection from config file and env overrides
	if err := config.LoadConfig("configs/config.yaml"); err != nil {
		log.Fatalf("load config: %v", err)
	}
	_, err := database.InitWithoutMigrate()
	if err != nil {
		panic(err)
	}
	db := database.GetDB()

	now := time.Now()
	// users
	db.Exec("INSERT INTO users (id, username, phone, status, created_at, updated_at) VALUES (1, 'demo_user', '13800000000', 1, ?, ?) ON DUPLICATE KEY UPDATE username='demo_user', phone='13800000000', updated_at=?", now, now, now)
	// wallets
	db.Exec("INSERT INTO wallets (user_id, balance, frozen) VALUES (1, 123450, 0) ON DUPLICATE KEY UPDATE balance=123450, frozen=0")
	// points_transactions
	db.Exec("INSERT INTO points_transactions (user_id, `change`, reason, created_at) VALUES (1, 100, 'signup_bonus', ?), (1, 50, 'order_reward', ?)", now, now)
	// coupons_templates
	db.Exec("INSERT INTO coupons_templates (id, name, type, value, min_order_amount, total_quantity, valid_from, valid_to, created_at) VALUES (100, '满100减10', 'amount', 1000, 10000, 1000, ?, DATE_ADD(?, INTERVAL 30 DAY), ?) ON DUPLICATE KEY UPDATE name='满100减10'", now, now, now)
	// coupons
	db.Exec("INSERT INTO coupons (template_id, user_id, code, status, expires_at, claimed_at) VALUES (100, 1, 'DEMO-001', 'unused', DATE_ADD(?, INTERVAL 15 DAY), ?)", now, now)
	// membership_packages
	db.Exec("INSERT INTO membership_packages (id, name, price, tea_coin_award, discount_rate, purchase_discount_rate, direct_commission_rate, team_commission_rate, upgrade_reward_rate, type) VALUES (200, '黄金会员', 9900, 0, 0.95, 1.00, 0.00, 0.00, 0.00, 'membership') ON DUPLICATE KEY UPDATE name='黄金会员'")
	// user_memberships
	db.Exec("INSERT INTO user_memberships (user_id, package_id, status, started_at, expires_at) VALUES (1, 200, 'active', ?, DATE_ADD(?, INTERVAL 365 DAY))", now, now)

	fmt.Println("✅ Demo user seeded: user_id=1")
}
