package service

import (
	"testing"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

func TestProductExclusiveVisibility(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	orig := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = orig })

	if err := db.AutoMigrate(&model.Category{}, &model.Product{}, &model.StoreProduct{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	cat := &model.Category{Name: "c1", Status: 1}
	if err := db.Create(cat).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	p1 := &model.Product{CategoryID: cat.ID, Name: "normal", Description: "", Price: decimal.NewFromInt(10), Status: 1}
	p2 := &model.Product{CategoryID: cat.ID, Name: "exclusive-1", Description: "", Price: decimal.NewFromInt(20), Status: 1}
	p3 := &model.Product{CategoryID: cat.ID, Name: "exclusive-2", Description: "", Price: decimal.NewFromInt(30), Status: 1}
	if err := db.Create(p1).Error; err != nil {
		t.Fatalf("create p1: %v", err)
	}
	if err := db.Create(p2).Error; err != nil {
		t.Fatalf("create p2: %v", err)
	}
	if err := db.Create(p3).Error; err != nil {
		t.Fatalf("create p3: %v", err)
	}

	// p2 exclusive to store 1, p3 exclusive to store 2
	if err := db.Create(&model.StoreProduct{StoreID: 1, ProductID: p2.ID, Stock: 10, PriceOverride: decimal.Zero, BizType: 3}).Error; err != nil {
		t.Fatalf("create store product p2: %v", err)
	}
	if err := db.Create(&model.StoreProduct{StoreID: 2, ProductID: p3.ID, Stock: 10, PriceOverride: decimal.Zero, BizType: 3}).Error; err != nil {
		t.Fatalf("create store product p3: %v", err)
	}

	svc := NewProductService()

	// global list: only p1
	got, _, err := svc.GetProducts(1, 50, nil, "", "")
	if err != nil {
		t.Fatalf("GetProducts: %v", err)
	}
	if len(got) != 1 || got[0].ID != p1.ID {
		t.Fatalf("expected only p1 in global list, got=%v", ids(got))
	}

	// store 1 list: p1 + p2
	got1, _, err := svc.GetProductsForStore(1, 50, nil, "", "", 1)
	if err != nil {
		t.Fatalf("GetProductsForStore(1): %v", err)
	}
	want1 := map[uint]bool{p1.ID: true, p2.ID: true}
	if len(got1) != 2 {
		t.Fatalf("expected 2 products for store1, got=%v", idsWithStore(got1))
	}
	for _, it := range got1 {
		if !want1[it.ID] {
			t.Fatalf("unexpected product for store1: %d", it.ID)
		}
	}

	// store 2 list: p1 + p3
	got2, _, err := svc.GetProductsForStore(1, 50, nil, "", "", 2)
	if err != nil {
		t.Fatalf("GetProductsForStore(2): %v", err)
	}
	want2 := map[uint]bool{p1.ID: true, p3.ID: true}
	if len(got2) != 2 {
		t.Fatalf("expected 2 products for store2, got=%v", idsWithStore(got2))
	}
	for _, it := range got2 {
		if !want2[it.ID] {
			t.Fatalf("unexpected product for store2: %d", it.ID)
		}
	}

	// global detail should hide exclusive
	if _, err := svc.GetProduct(p2.ID); err != ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound for exclusive global detail, got=%v", err)
	}

	// store detail should hide exclusive when store mismatch
	if _, err := svc.GetProductForStore(p3.ID, 1); err != ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound for store mismatch, got=%v", err)
	}
}

func ids(list []*model.Product) []uint {
	out := make([]uint, 0, len(list))
	for _, p := range list {
		out = append(out, p.ID)
	}
	return out
}

func idsWithStore(list []*ProductWithStore) []uint {
	out := make([]uint, 0, len(list))
	for _, p := range list {
		out = append(out, p.ID)
	}
	return out
}
