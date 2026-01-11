package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("tea.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	
	var cols []struct {
		CID       int
		Name      string
		Type      string
		Notnull   int
		DfltValue *string
		PK        int
	}
	
	db.Raw("PRAGMA table_info(orders)").Scan(&cols)
	fmt.Println("Orders table columns:")
	for _, c := range cols {
		fmt.Printf("  %s: %s\n", c.Name, c.Type)
	}
	
	// 检查是否有 table_id/table_no
	hasTableID := false
	hasTableNo := false
	for _, c := range cols {
		if c.Name == "table_id" {
			hasTableID = true
		}
		if c.Name == "table_no" {
			hasTableNo = true
		}
	}
	
	if !hasTableID || !hasTableNo {
		fmt.Println("\n警告: orders表缺少桌号字段！需要运行数据库迁移")
	}
}
