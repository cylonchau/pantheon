package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateLabels_Success 测试正常创建 Label
// 场景：传入一组 key-value 标签，期望全部成功创建
func TestCreateLabels_Success(t *testing.T) {
	// Arrange: 准备测试环境
	_ = SetupTestDB(t)

	labels := map[string]string{
		"env":    "production",
		"region": "cn-east",
		"team":   "sre",
	}

	// Act: 执行被测函数
	createdLabels, err := CreateLabels(labels)

	// Assert: 验证结果
	require.NoError(t, err, "CreateLabels should not return error")
	assert.Len(t, createdLabels, 3, "Should create 3 labels")

	// 验证每个 label 都被正确创建
	for _, label := range createdLabels {
		assert.NotZero(t, label.ID, "Label ID should be assigned")
		assert.Contains(t, labels, label.Key, "Label key should exist in input")
		assert.Equal(t, labels[label.Key], label.Value, "Label value should match")
	}
}

// TestCreateLabels_Duplicate 测试重复创建相同 Label (幂等性)
// 场景：相同的 key-value 标签创建两次，期望不会报错且返回同一条记录
func TestCreateLabels_Duplicate(t *testing.T) {
	// Arrange
	_ = SetupTestDB(t)

	labels := map[string]string{
		"env": "staging",
	}

	// Act: 创建两次相同的 label
	firstCreate, err1 := CreateLabels(labels)
	secondCreate, err2 := CreateLabels(labels)

	// Assert: 两次创建都应该成功，且返回相同的 ID
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, firstCreate[0].ID, secondCreate[0].ID,
		"Duplicate label should return same record (idempotent)")
}

// TestCreateLabels_EmptyMap 测试传入空 map
// 场景：传入空的 labels map，期望返回空数组而不是 nil
func TestCreateLabels_EmptyMap(t *testing.T) {
	// Arrange
	_ = SetupTestDB(t)

	labels := map[string]string{}

	// Act
	createdLabels, err := CreateLabels(labels)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, createdLabels, "Empty input should return empty slice")
}

// TestGetLabelsWithLabels_Found 测试查询存在的 Label
func TestGetLabelsWithLabels_Found(t *testing.T) {
	// Arrange: 先创建 label
	db := SetupTestDB(t)
	db.Create(&Label{Key: "app", Value: "nginx"})

	// Act
	found, err := GetLabelsWithLabels(map[string]string{"app": "nginx"})

	// Assert
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "nginx", found[0].Value)
}

// TestGetLabelsWithLabels_NotFound 测试查询不存在的 Label
func TestGetLabelsWithLabels_NotFound(t *testing.T) {
	// Arrange
	_ = SetupTestDB(t)

	// Act: 查询不存在的 label
	found, err := GetLabelsWithLabels(map[string]string{"nonexistent": "value"})

	// Assert: 不应该报错，返回空结果或 zero-ID 的结构体
	require.NoError(t, err)
	assert.True(t, len(found) == 0 || found[0].ID == 0,
		"Should return empty or zero-ID slice for non-existent labels")
}
