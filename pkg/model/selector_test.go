package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateSelectors_Success 测试正常创建 Selector
func TestCreateSelectors_Success(t *testing.T) {
	// Arrange
	_ = SetupTestDB(t)

	selectors := map[string]string{
		"namespace": "monitoring",
		"cluster":   "prod-01",
	}

	// Act
	created, err := CreateSelectors(selectors)

	// Assert
	require.NoError(t, err)
	assert.Len(t, created, 2)
}

// TestListSelector_Empty 测试空数据库查询 Selector
func TestListSelector_Empty(t *testing.T) {
	// Arrange
	_ = SetupTestDB(t)

	// Act
	list, err := ListSelector()

	// Assert
	require.NoError(t, err)
	assert.Empty(t, list, "Empty database should return empty list")
}

// TestListSelector_WithData 测试有数据时查询 Selector
func TestListSelector_WithData(t *testing.T) {
	// Arrange
	db := SetupTestDB(t)
	db.Create(&Selector{Key: "ns", Value: "default"})
	db.Create(&Selector{Key: "ns", Value: "kube-system"})

	// Act
	list, err := ListSelector()

	// Assert
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

// TestUpdateSelectorByKeyValue_Success 测试更新 Selector
func TestUpdateSelectorByKeyValue_Success(t *testing.T) {
	// Arrange
	db := SetupTestDB(t)
	db.Create(&Selector{Key: "env", Value: "dev"})

	// Act: 将 env=dev 更新为 env=prod
	err := UpdateSelectorByKeyValue("env", "dev", "env", "prod")

	// Assert
	require.NoError(t, err)

	// 验证更新后的值
	var updated Selector
	db.Where("`key` = ? AND `value` = ?", "env", "prod").First(&updated)
	assert.Equal(t, "prod", updated.Value)
}

// TestUpdateSelectorByKeyValue_NotFound 测试更新不存在的 Selector
func TestUpdateSelectorByKeyValue_NotFound(t *testing.T) {
	// Arrange
	_ = SetupTestDB(t)

	// Act
	err := UpdateSelectorByKeyValue("nonexistent", "value", "new", "value")

	// Assert: 应该返回错误
	assert.Error(t, err, "Should return error for non-existent selector")
}
