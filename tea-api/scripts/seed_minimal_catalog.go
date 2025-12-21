package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	host := getenv("TEA_DATABASE_HOST", "127.0.0.1")
	port := getenv("TEA_DATABASE_PORT", "3306")
	user := getenv("TEA_DATABASE_USERNAME", "root")
	pass := getenv("TEA_DATABASE_PASSWORD", "root")
	dbname := getenv("TEA_DATABASE_DBNAME", "tea_shop")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", user, pass, host, port, dbname)
	log.Printf("Connecting MySQL: %s@%s:%s/%s", user, host, port, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil { log.Fatalf("open dsn error: %v", err) }
	defer db.Close()
	if err := db.Ping(); err != nil { log.Fatalf("ping error: %v", err) }

	sqls := []string{
		// categories
		"INSERT INTO categories (id, name, description, image, sort, status, parent_id) SELECT 1, '默认分类', '', '', 0, 1, 0 WHERE NOT EXISTS (SELECT 1 FROM categories WHERE id=1)",
		// store
		"INSERT INTO stores (id, name, address, phone, latitude, longitude, business_hours, images, status) SELECT 1, '测试门店', '上海某路1号', '12345678', 31.2304, 121.4737, '09:00-21:00', '[]', 1 WHERE NOT EXISTS (SELECT 1 FROM stores WHERE id=1)",
		// product
		"INSERT INTO products (id, category_id, name, description, images, price, original_price, stock, status) SELECT 1001, 1, '测试商品', '自动化测试用商品', '[]', 19.99, 19.99, 100, 1 WHERE NOT EXISTS (SELECT 1 FROM products WHERE id=1001)",
		// sku
		"INSERT INTO product_skus (id, product_id, sku_name, sku_code, price, stock, status) SELECT 2001, 1001, '默认SKU', 'sku-seed-2001', 19.99, 100, 1 WHERE NOT EXISTS (SELECT 1 FROM product_skus WHERE id=2001)",
	}
	for i, s := range sqls {
		if _, err := db.Exec(s); err != nil {
			log.Fatalf("seed step %d failed: %v", i+1, err)
		}
	}
	log.Println("Seed minimal catalog completed.")
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" { return def }
	return v
}
