package model

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 设置测试 DB_TESTING
var DB_TESTING *gorm.DB

func TestMain(m *testing.M) {
	var err error
	DB_TESTING, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect to database")
	}

	// 自动迁移，创建表
	DB_TESTING.AutoMigrate(&Label{})

	m.Run()
}

func TestCreateLabels(t *testing.T) {
	labels := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	createdLabels, err := CreateLabels(labels)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(createdLabels) != len(labels) {
		t.Fatalf("expected %d labels, got %d", len(labels), len(createdLabels))
	}
}

func TestCreateLabels_Error(t *testing.T) {
	// 这里可以通过故意破坏 Label 结构来模拟错误，例如添加空 key
	labels := map[string]string{
		"": "value3",
	}

	_, err := CreateLabels(labels)
	if err == nil {
		t.Fatal("expected an error, got none")
	}
}

func TestGetLabelsWithLabels(t *testing.T) {
	// 先创建一些标签
	labels := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	_, err := CreateLabels(labels)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 现在获取这些标签
	gotLabels, err := GetLabelsWithLabels(labels)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(gotLabels) != len(labels) {
		t.Fatalf("expected %d labels, got %d", len(labels), len(gotLabels))
	}
}

func TestGetLabelsWithLabels_NotFound(t *testing.T) {
	labels := map[string]string{
		"nonexistent": "value",
	}

	gotLabels, err := GetLabelsWithLabels(labels)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(gotLabels) != 0 {
		t.Fatalf("expected 0 labels, got %d", len(gotLabels))
	}
}
