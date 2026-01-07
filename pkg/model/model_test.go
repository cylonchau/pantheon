package model

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm/logger"
)

// 假设你的 config 包中的 CONFIG 结构如下
type MySQLConfig struct {
	User              string
	Password          string
	IP                string
	Port              string
	Database          string
	MaxOpenConnection int
	MaxIdleConnection int
}

type SQLiteConfig struct {
	File              string
	MaxOpenConnection int
	MaxIdleConnection int
}

type Config struct {
	MySQL  MySQLConfig
	SQLite SQLiteConfig
}

var CONFIG Config

func TestInitDB_MySQL(t *testing.T) {
	// 设置模拟数据库
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer mockDB.Close()

	// 配置 MySQL
	CONFIG.MySQL = MySQLConfig{
		User:              "testuser",
		Password:          "testpass",
		IP:                "localhost",
		Port:              "3306",
		Database:          "testdb",
		MaxOpenConnection: 10,
		MaxIdleConnection: 5,
	}

	// 模拟数据库查询行为
	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	// 初始化数据库
	err = InitDB("mysql")
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	// 断言 DB 和 dbConn 已正确初始化
	if DB == nil {
		t.Fatalf("DB is nil after initialization")
	}
	if dbConn == nil {
		t.Fatalf("dbConn is nil after initialization")
	}

	// 校验模拟的 SQL 操作
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInitDB_SQLite(t *testing.T) {
	// 设置 SQLite 配置
	CONFIG.SQLite = SQLiteConfig{
		File:              "testdb",
		MaxOpenConnection: 10,
		MaxIdleConnection: 5,
	}

	// 初始化数据库
	err := InitDB("sqlite")
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	// 断言 DB 和 dbConn 已正确初始化
	if DB == nil {
		t.Fatalf("DB is nil after initialization")
	}
	if dbConn == nil {
		t.Fatalf("dbConn is nil after initialization")
	}
}

func TestKlogLogger(t *testing.T) {
	logger := KlogLogger{logLevel: logger.Info}

	// 测试日志记录的行为
	logger.Info(context.Background(), "This is an info message")
	logger.Warn(context.Background(), "This is a warning message")
	logger.Error(context.Background(), "This is an error message")

	// 你可以通过设置 klog 的 Verbosity 和验证输出是否正确来进一步扩展测试
}
