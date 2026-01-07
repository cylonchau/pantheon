package migration

import (
	"errors"
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/cylonchau/pantheon/pkg/config"
	"github.com/cylonchau/pantheon/pkg/model"
)

func Upgrade(driver string) (enconterError error) {
	if driver == "" {
		return errors.New("Unknown database driver")
	}
	var dbInterface *gorm.DB
	switch driver {
	case "mysql":
		if dbInterface, enconterError = MySQL(); enconterError == nil {
			// 检查数据库是否存在，如果不存在则创建数据库
			if enconterError = createDatabaseIfNotExists(dbInterface); enconterError != nil {
				return
			}
			// 执行迁移
			if enconterError = upgradeMigrate(dbInterface); enconterError != nil {
				return
			}
		}
		return enconterError
	case "sqlite":
		if dbInterface, enconterError = SQLite(); enconterError == nil {
			if enconterError = upgradeMigrate(dbInterface); enconterError != nil {
				return
			}
		}
		return nil
	default:
		enconterError = errors.New("UnknownDriver")
	}
	return enconterError
}

func Migrate(driver string) (enconterError error) {
	if driver == "" {
		enconterError = errors.New("Unknown database driver")
		return
	}
	var dbInterface *gorm.DB

	switch driver {
	case "mysql":
		if dbInterface, enconterError = MySQL(); enconterError == nil {
			// 检查数据库是否存在，如果不存在则创建数据库
			if enconterError = createDatabaseIfNotExists(dbInterface); enconterError != nil {
				return
			}
			// 执行迁移
			if enconterError = autoMigrate(dbInterface); enconterError != nil {
				return
			}
		}
		return enconterError
	case "sqlite":
		if dbInterface, enconterError = SQLite(); enconterError == nil {
			if enconterError = autoMigrate(dbInterface); enconterError != nil {
				return
			}
		}
		return
	}
	return nil
}

func createDatabaseIfNotExists(dbInterface *gorm.DB) error {
	database := config.CONFIG.MySQL.Database
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", database)
	return dbInterface.Exec(sql).Error
}

func upgradeMigrate(dbInterface *gorm.DB) (enconterError error) {

	if enconterError = dbInterface.AutoMigrate(&model.Target{}); enconterError != nil {
		return
	}

	if enconterError = dbInterface.AutoMigrate(&model.Selector{}); enconterError != nil {
		return
	}

	if enconterError = dbInterface.AutoMigrate(&model.Param{}); enconterError != nil {
		return
	}

	if enconterError = dbInterface.AutoMigrate(&model.Label{}); enconterError != nil {
		return
	}

	return nil
}

func autoMigrate(dbInterface *gorm.DB) (enconterError error) {
	if !dbInterface.Migrator().HasTable(&model.Target{}) {
		if enconterError = dbInterface.AutoMigrate(&model.Target{}); enconterError != nil {
			return
		}
	}
	if !dbInterface.Migrator().HasTable(&model.Selector{}) {
		if enconterError = dbInterface.AutoMigrate(&model.Selector{}); enconterError != nil {
			return
		}
	}
	if !dbInterface.Migrator().HasTable(&model.Param{}) {
		if enconterError = dbInterface.AutoMigrate(&model.Param{}); enconterError != nil {
			return
		}
	}
	if !dbInterface.Migrator().HasTable(&model.Label{}) {
		if enconterError = dbInterface.AutoMigrate(&model.Label{}); enconterError != nil {
			return
		}
	}
	return nil
}

func SQLite() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(config.CONFIG.SQLite.File+".db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
}

func MySQL() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.CONFIG.MySQL.User,
		config.CONFIG.MySQL.Password,
		config.CONFIG.MySQL.IP,
		config.CONFIG.MySQL.Port,
		config.CONFIG.MySQL.Database,
	)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
}
