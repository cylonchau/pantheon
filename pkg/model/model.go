package model

import (
	"context"
	"database/sql"
	"net/url"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"k8s.io/klog/v2"

	"github.com/cylonchau/pantheon/pkg/config"
)

var dbConn *sql.DB
var DB *gorm.DB

// KlogLogger 实现 gorm.Logger 接口
type KlogLogger struct {
	logLevel logger.LogLevel
}

// LogMode 设置日志级别
func (l KlogLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := l
	newLogger.logLevel = level
	return newLogger
}

// Info 记录 info 级别的日志
func (l KlogLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info && klog.V(4).Enabled() {
		klog.Infof(msg, data...)
	}
}

// Warn 记录 warn 级别的日志
func (l KlogLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn && klog.V(1).Enabled() {
		klog.Warningf(msg, data...)
	}
}

// Error 记录 error 级别的日志
func (l KlogLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error && klog.V(1).Enabled() {
		klog.Errorf(msg, data...)
	}
}

// Trace 记录 trace 级别的日志
func (l KlogLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		// 错误日志输出在 v1 级别
		if klog.V(1).Enabled() {
			klog.Errorf("Trace Error: %v | SQL: %s | Rows affected: %d | Time: %s", err, sql, rows, elapsed)
		}
	} else {
		// 正常 SQL 日志输出在 v4 级别
		if klog.V(4).Enabled() {
			klog.Infof("Trace Success | SQL: %s | Rows affected: %d | Time: %s", sql, rows, elapsed)
		}
	}
}

func InitDB(driver string) error {
	var enconterError error
	switch driver {
	case "mysql":
		dsn := config.CONFIG.MySQL.User + ":" + config.CONFIG.MySQL.Password + "@tcp(" + config.CONFIG.MySQL.IP + ":" + config.CONFIG.MySQL.Port + ")/" + config.CONFIG.MySQL.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
		newLogger := KlogLogger{logLevel: logger.Info}

		if DB, enconterError = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger}); enconterError == nil {
			if dbConn, enconterError = DB.DB(); enconterError == nil {
				dbConn.SetMaxOpenConns(config.CONFIG.MySQL.MaxOpenConnection)
				dbConn.SetMaxIdleConns(config.CONFIG.MySQL.MaxIdleConnection)

				klog.V(4).Infof("Databases stats is %+v", dbConn.Stats())
				return nil
			}
		}
	case "sqlite":
		newLogger := KlogLogger{logLevel: logger.Info}
		if DB, enconterError = gorm.Open(sqlite.Open(config.CONFIG.SQLite.File+".db"), &gorm.Config{Logger: newLogger}); enconterError == nil {
			if dbConn, enconterError = DB.DB(); enconterError == nil {
				dbConn.SetMaxOpenConns(config.CONFIG.SQLite.MaxOpenConnection)
				dbConn.SetMaxIdleConns(config.CONFIG.SQLite.MaxIdleConnection)

				klog.V(4).Infof("Databases stats is %+v", dbConn.Stats())
				return nil
			}
		}
	}
	return enconterError
}

func mapToURLParams(params map[string]string) string {
	// 使用 url.Values 来处理参数
	urlParams := url.Values{}
	for key, value := range params {
		urlParams.Add(key, value)
	}
	return urlParams.Encode() // 返回编码后的字符串
}

func parseConfigURL(raw string) *url.URL {
	if raw == "" {
		raw = "localhost"
	}
	// 检查是否包含 schema
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "http://" + raw // 默认使用 http
	}

	// 解析 URL
	parsedURL, err := url.Parse(raw)
	if err != nil {
		return nil // 返回 nil 代表解析失败
	}

	// 如果没有 port，检查 schema 并赋默认值
	if parsedURL.Port() == "" {
		if parsedURL.Scheme == "http" {
			parsedURL.Host += ":80"
		} else if parsedURL.Scheme == "https" {
			parsedURL.Host += ":443"
		}
	}

	// 处理路径，默认为 /
	if parsedURL.Path == "" {
		parsedURL.Path = "/"
	}

	return parsedURL
}
