package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"gorm.io/gorm"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/pkg/database"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "tea-api/configs/config.yaml", "tea-api config file")
	flag.Parse()

	if err := config.LoadConfig(cfgPath); err != nil {
		log.Fatalf("load config: %v", err)
	}

	if _, err := database.InitWithoutMigrate(); err != nil {
		log.Fatalf("init database: %v", err)
	}

	db := database.GetDB()

	categories := []model.Category{
		{Name: "招牌鲜奶茶", Description: "店内人气鲜奶茶系列", Sort: 10, Status: 1},
		{Name: "季节限定", Description: "当季热卖限定饮品", Sort: 20, Status: 1},
		{Name: "果茶气泡", Description: "清爽果茶或气泡饮", Sort: 30, Status: 1},
	}

	var createdCats int
	for _, cat := range categories {
		var existing model.Category
		if err := db.Where("name = ?", cat.Name).First(&existing).Error; err == nil {
			fmt.Printf("category exists: %s (id=%d)\n", existing.Name, existing.ID)
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatalf("query category %s: %v", cat.Name, err)
		}
		if err := db.Create(&cat).Error; err != nil {
			log.Fatalf("create category %s: %v", cat.Name, err)
		}
		createdCats++
		fmt.Printf("created category: %s (id=%d)\n", cat.Name, cat.ID)
	}

	stores := []model.Store{
		{
			Name:          "茶心阁 · 西湖旗舰店",
			Address:       "杭州市西湖区孤山路 25 号",
			Phone:         "0571-88886666",
			Latitude:      30.24123,
			Longitude:     120.15011,
			BusinessHours: "每日 10:00-22:00",
			Images:        "https://static.example.com/stores/west-lake.jpg",
			Status:        1,
		},
		{
			Name:          "茶心阁 · 滨江万象城",
			Address:       "杭州市滨江区江南大道 228 号",
			Phone:         "0571-88997777",
			Latitude:      30.20517,
			Longitude:     120.21043,
			BusinessHours: "工作日 10:00-21:00 / 周末 10:00-22:00",
			Images:        "https://static.example.com/stores/binjiang.jpg",
			Status:        1,
		},
		{
			Name:          "茶心阁 · 武林银泰",
			Address:       "杭州市下城区延安路 530 号",
			Phone:         "0571-88225599",
			Latitude:      30.27532,
			Longitude:     120.16742,
			BusinessHours: "每日 10:00-22:00",
			Images:        "https://static.example.com/stores/wulin.jpg",
			Status:        1,
		},
	}

	var createdStores int
	for _, st := range stores {
		var existing model.Store
		if err := db.Where("name = ?", st.Name).First(&existing).Error; err == nil {
			fmt.Printf("store exists: %s (id=%d)\n", existing.Name, existing.ID)
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatalf("query store %s: %v", st.Name, err)
		}
		if err := db.Create(&st).Error; err != nil {
			log.Fatalf("create store %s: %v", st.Name, err)
		}
		createdStores++
		fmt.Printf("created store: %s (id=%d)\n", st.Name, st.ID)
	}

	fmt.Printf("seed complete: %d categories, %d stores inserted (existing entries skipped)\n", createdCats, createdStores)
}
