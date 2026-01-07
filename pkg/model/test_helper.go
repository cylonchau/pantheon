package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 创建一个内存中的 SQLite 数据库用于测试，更接近真实场景，同时保持测试的隔离性
func SetupTestDB(t *testing.T) *gorm.DB {
	// 使用 :memory: 创建内存数据库，每次测试都是全新的
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 测试时关闭日志
	})
	require.NoError(t, err, "Failed to create in-memory database")

	// 自动迁移所有模型表结构
	// 注意：迁移顺序很重要，被引用的表需要先创建
	err = db.AutoMigrate(&Label{}, &Param{}, &Selector{}, &Target{})
	require.NoError(t, err, "Failed to migrate database schema")

	// 将全局 DB 变量指向测试数据库
	DB = db

	return db
}
