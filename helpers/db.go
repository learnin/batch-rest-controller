package helpers

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const DATABASE_CONFIG_FILE = "config/database.json"

type config struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type DataSource struct {
	db *gorm.DB
}

func (ds *DataSource) Connect() error {
	jsonHelper := Json{}
	var cfg config
	if err := jsonHelper.UnmarshalJsonFile(DATABASE_CONFIG_FILE, &cfg); err != nil {
		return err
	}
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Asia%%2FTokyo", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	db, err := gorm.Open("mysql", connStr)
	if err != nil {
		return err
	}
	if err := db.DB().Ping(); err != nil {
		return err
	}
	ds.db = &db
	return nil
}

func (ds *DataSource) LogMode(b bool) {
	ds.db.LogMode(b)
}

func (ds *DataSource) Close() error {
	return ds.db.Close()
}

func (ds *DataSource) GetDB() *gorm.DB {
	return ds.db
}

func (ds *DataSource) DoInTransaction(callback func(tx *gorm.DB) error) error {
	tx := ds.db.Begin()
	if err := callback(tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
