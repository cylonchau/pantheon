package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateParams_Success 测试正常创建 Param
// 场景：传入一组 key-value 参数，期望全部成功创建
func TestCreateParams_Success(t *testing.T) {
	// Arrange: 准备测试环境
	_ = SetupTestDB(t)

	params := map[string]string{
		"timeout":  "30s",
		"interval": "15s",
	}

	// Act: 执行被测函数
	createdParams, err := CreateParams(params)

	// Assert: 验证结果
	require.NoError(t, err, "CreateParams should not return error")
	assert.Len(t, createdParams, 2, "Should create 2 params")

	// 验证每个 param 都被正确创建
	for _, param := range createdParams {
		assert.NotZero(t, param.ID, "Param ID should be assigned")
		assert.Contains(t, params, param.Key, "Param key should exist in input")
		assert.Equal(t, params[param.Key], param.Value, "Param value should match")
	}
}

// TestCreateParams_Duplicate 测试重复创建相同 Param (幂等性)
// 场景：相同的 key-value 参数创建两次，期望不会报错且返回同一条记录
func TestCreateParams_Duplicate(t *testing.T) {
	// Arrange
	_ = SetupTestDB(t)

	params := map[string]string{
		"max_retries": "3",
	}

	// Act: 创建两次相同的 param
	firstCreate, err1 := CreateParams(params)
	secondCreate, err2 := CreateParams(params)

	// Assert: 两次创建都应该成功，且返回相同的 ID
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, firstCreate[0].ID, secondCreate[0].ID,
		"Duplicate param should return same record (idempotent)")
}

// TestCreateParams_EmptyMap 测试传入空 map
// 场景：传入空的 params map，期望返回空数组而不是 nil
func TestCreateParams_EmptyMap(t *testing.T) {
	// Arrange
	_ = SetupTestDB(t)

	params := map[string]string{}

	// Act
	createdParams, err := CreateParams(params)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, createdParams, "Empty input should return empty slice")
}
