package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/cylonchau/pantheon/pkg/api/query"
	"github.com/cylonchau/pantheon/pkg/api/target"
)

var DB_TEST *gorm.DB

func setupDatabase() {
	var err error
	DB_TEST, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// 自动迁移模式
	DB_TEST.AutoMigrate(&Target{}, &Label{}, &Param{}, &Selector{})
}

func TestCreateTargets(t *testing.T) {
	setupDatabase()
	target := &target.Target{
		Targets: []target.TargetItem{
			{
				Address:       "http://example.com",
				MetricPath:    "/metrics",
				ScrapeTime:    30,
				ScrapeTimeout: 10,
				Labels:        map[string]string{"env": "test"},
				Params:        map[string]string{"param1": "value1"},
			},
		},
		InstanceSelector: map[string]string{"instance": "example"},
	}

	err := CreateTargets(target)
	assert.NoError(t, err)

	// 验证创建的目标是否存在
	var createdTarget Target
	err = DB_TEST.First(&createdTarget, "address = ?", "example.com").Error
	assert.NoError(t, err)
	assert.Equal(t, "example.com", createdTarget.Address)
}

func TestDeleteTargetWithName(t *testing.T) {
	setupDatabase()
	// 首先创建一个目标以便删除
	target := &Target{
		Address: "http://delete.com",
	}
	DB_TEST.Create(target)

	err := DeleteTargetWithName("http://delete.com")
	assert.NoError(t, err)

	// 验证目标是否已被删除
	var deletedTarget Target
	err = DB_TEST.First(&deletedTarget, "address = ?", "http://delete.com").Error
	assert.Error(t, err) // 应该返回错误，因为目标已被删除
}

func TestListTargetWithCtl(t *testing.T) {
	setupDatabase()
	// 创建一个目标以便测试
	target := &Target{
		Address:       "http://list.com",
		MetricPath:    "/metrics",
		ScrapeTime:    30,
		ScrapeTimeout: 10,
	}
	DB_TEST.Create(target)

	query := &query.QueryWithLabel{
		Key:   "someKey",
		Value: "someValue",
	}

	results, err := ListTargetWithCtl(query)
	assert.NoError(t, err)
	assert.NotNil(t, results) // 验证结果不为空
}
