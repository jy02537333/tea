package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/pkg/database"

	"github.com/shopspring/decimal"
)

func main() {
	var cfgPath string
	var productID uint
	var skuName string
	var skuCode string
	var priceStr string
	var stock int

	flag.StringVar(&cfgPath, "config", "configs/config.mysql.local.yaml", "path to config.yaml")
	flag.UintVar(&productID, "product", 1, "product id to attach sku to")
	flag.StringVar(&skuName, "sku_name", "默认规格", "SKU name")
	flag.StringVar(&skuCode, "sku_code", "", "SKU code (unique). if empty will be generated")
	flag.StringVar(&priceStr, "price", "0", "price, e.g. 99.00")
	flag.IntVar(&stock, "stock", 100, "stock quantity")
	flag.Parse()

	if err := config.LoadConfig(cfgPath); err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.InitWithoutMigrate()
	if err != nil {
		log.Fatalf("init db: %v", err)
	}
	_ = db

	p, err := decimal.NewFromString(priceStr)
	if err != nil {
		if iv, err2 := strconv.Atoi(priceStr); err2 == nil {
			p = decimal.NewFromInt(int64(iv))
		} else {
			log.Fatalf("invalid price: %v", err)
		}
	}

	sku := model.ProductSku{
		ProductID: productID,
		SkuName:   skuName,
		SkuCode:   skuCode,
		Price:     p,
		Stock:     stock,
		Attrs:     "[]",
		Status:    1,
	}

	if sku.SkuCode == "" {
		sku.SkuCode = fmt.Sprintf("sku_%d_%d", productID, 1)
	}

	if err := database.GetDB().Create(&sku).Error; err != nil {
		log.Fatalf("create sku failed: %v", err)
	}

	fmt.Printf("created sku id=%d product_id=%d sku_code=%s sku_name=%s price=%s stock=%d\n", sku.ID, sku.ProductID, sku.SkuCode, sku.SkuName, sku.Price.String(), sku.Stock)
}
