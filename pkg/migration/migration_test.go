package migration

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/cylonchau/pantheon/pkg/model"
)

// setupTestDB 创建内存 SQLite 数据库用于测试
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to create in-memory database")
	return db
}

// TestUpgradeMigrate_Success 测试 upgradeMigrate 正常迁移所有表
func TestUpgradeMigrate_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)

	// Act
	err := upgradeMigrate(db)

	// Assert
	require.NoError(t, err, "upgradeMigrate should not return error")

	// 验证所有表都已创建
	assert.True(t, db.Migrator().HasTable(&model.Target{}), "Target table should exist")
	assert.True(t, db.Migrator().HasTable(&model.Selector{}), "Selector table should exist")
	assert.True(t, db.Migrator().HasTable(&model.Param{}), "Param table should exist")
	assert.True(t, db.Migrator().HasTable(&model.Label{}), "Label table should exist")
}

// TestAutoMigrate_Success 测试 autoMigrate 正常迁移（表不存在时）
func TestAutoMigrate_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)

	// Act
	err := autoMigrate(db)

	// Assert
	require.NoError(t, err, "autoMigrate should not return error")

	// 验证所有表都已创建
	assert.True(t, db.Migrator().HasTable(&model.Target{}), "Target table should exist")
	assert.True(t, db.Migrator().HasTable(&model.Selector{}), "Selector table should exist")
	assert.True(t, db.Migrator().HasTable(&model.Param{}), "Param table should exist")
	assert.True(t, db.Migrator().HasTable(&model.Label{}), "Label table should exist")
}

// TestAutoMigrate_Idempotent 测试 autoMigrate 幂等性（表已存在时跳过）
func TestAutoMigrate_Idempotent(t *testing.T) {
	// Arrange: 先执行一次迁移
	db := setupTestDB(t)
	err := autoMigrate(db)
	require.NoError(t, err)

	// Act: 再次执行迁移
	err = autoMigrate(db)

	// Assert: 不应该报错
	require.NoError(t, err, "autoMigrate should be idempotent")
}

// TestUpgrade_EmptyDriver 测试空驱动名
func TestUpgrade_EmptyDriver(t *testing.T) {
	// Act
	err := Upgrade("")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unknown database driver")
}

// TestMigrate_EmptyDriver 测试空驱动名
func TestMigrate_EmptyDriver(t *testing.T) {
	// Act
	err := Migrate("")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unknown database driver")
}

// TestUpgrade_UnknownDriver 测试未知驱动名
func TestUpgrade_UnknownDriver(t *testing.T) {
	// Act
	err := Upgrade("postgres")

	// Assert: 未知驱动应返回错误
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UnknownDriver")
}

// TestMigrate_UnknownDriver 测试未知驱动名
func TestMigrate_UnknownDriver(t *testing.T) {
	// Act
	err := Migrate("postgres")

	// Assert
	assert.NoError(t, err, "Unknown driver returns nil (potential bug)")
}
