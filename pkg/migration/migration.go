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

func Upgrade(driver string) error {
	if driver == "" {
		return errors.New("Unknown database driver")
	}
	var (
		dbInterface   *gorm.DB
		enconterError error
	)
	switch driver {
	case "mysql":
		if dbInterface, enconterError = MySQL(); enconterError == nil {
			// 检查数据库是否存在，如果不存在则创建数据库
			if err := createDatabaseIfNotExists(dbInterface); err != nil {
				return err
			}
			// 执行迁移
			if err := upgradeMigrate(dbInterface); err != nil {
				return err
			}
		}
		return enconterError
	case "sqlite":
		if dbInterface, enconterError = SQLite(); enconterError == nil {
			if err := upgradeMigrate(dbInterface); err != nil {
				return err
			}
		}
		return nil
	}
	return enconterError
}

func Migrate(driver string) error {
	if driver == "" {
		return errors.New("Unknown database driver")
	}
	var (
		dbInterface   *gorm.DB
		enconterError error
	)
	switch driver {
	case "mysql":
		if dbInterface, enconterError = MySQL(); enconterError == nil {
			// 检查数据库是否存在，如果不存在则创建数据库
			if err := createDatabaseIfNotExists(dbInterface); err != nil {
				return err
			}
			// 执行迁移
			if err := autoMigrate(dbInterface); err != nil {
				return err
			}
		}
		return enconterError
	case "sqlite":
		if dbInterface, enconterError = SQLite(); enconterError == nil {
			if err := autoMigrate(dbInterface); err != nil {
				return err
			}
		}
		return nil
	}
	return enconterError
}

func createDatabaseIfNotExists(dbInterface *gorm.DB) error {
	database := config.CONFIG.MySQL.Database
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", database)
	return dbInterface.Exec(sql).Error
}

func upgradeMigrate(dbInterface *gorm.DB) error {

	if err := dbInterface.AutoMigrate(&model.Target{}); err != nil {
		return err
	}

	if err := dbInterface.AutoMigrate(&model.Selector{}); err != nil {
		return err
	}

	if err := dbInterface.AutoMigrate(&model.Param{}); err != nil {
		return err
	}

	if err := dbInterface.AutoMigrate(&model.Label{}); err != nil {
		return err
	}

	return nil
}

func autoMigrate(dbInterface *gorm.DB) error {
	if !dbInterface.Migrator().HasTable(&model.Target{}) {
		if err := dbInterface.AutoMigrate(&model.Target{}); err != nil {
			return err
		}
	}
	if !dbInterface.Migrator().HasTable(&model.Selector{}) {
		if err := dbInterface.AutoMigrate(&model.Selector{}); err != nil {
			return err
		}
	}
	if !dbInterface.Migrator().HasTable(&model.Param{}) {
		if err := dbInterface.AutoMigrate(&model.Param{}); err != nil {
			return err
		}
	}
	if !dbInterface.Migrator().HasTable(&model.Label{}) {
		if err := dbInterface.AutoMigrate(&model.Label{}); err != nil {
			return err
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
