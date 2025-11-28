package service_test

import (
	"testing"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/database"
	"tea-api/pkg/logx"

	"github.com/stretchr/testify/require"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDeleteProductImage_DeletesOSSAndDB(t *testing.T) {
	// 初始化日志
	_ = logx.Init()

	// 使用 in-memory SQLite 替代外部 MySQL，便于 CI 与本地测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	// 注入到 package variable，服务代码通过 database.GetDB() 获取
	database.DB = db
	// 自动创建需要的表结构并初始化测试数据
	require.NoError(t, db.AutoMigrate(&model.Product{}, &model.ProductImage{}))

	p := &model.Product{Name: "test-product"}
	require.NoError(t, db.Create(p).Error)

	img := &model.ProductImage{ProductID: p.ID, ImageURL: "https://bucket/test1.jpg", Sort: 0, IsMain: false}
	require.NoError(t, db.Create(img).Error)

	called := false
	// mock OSS delete
	service.DeleteFilesFunc = func(urls []string) error {
		called = true
		// assert the URL is as expected
		if len(urls) != 1 || urls[0] != img.ImageURL {
			t.Fatalf("unexpected urls: %v", urls)
		}
		return nil
	}

	svc := service.NewProductService()
	require.NoError(t, svc.DeleteProductImage(img.ID))

	// ensure DB record deleted
	var count int64
	require.NoError(t, db.Model(&model.ProductImage{}).Where("id = ?", img.ID).Count(&count).Error)
	require.Equal(t, int64(0), count)
	require.True(t, called, "oss delete should be called")
}
