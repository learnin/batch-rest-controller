package helpers

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

const DATABASE_CONFIG_FILE = "config/database.json"

type config struct {
	DatabaseFile string
}

type DataSource struct {
	db *gorm.DB
	tx *gorm.DB
}

func (ds *DataSource) Connect() error {
	jsonHelper := Json{}
	var cfg config
	if err := jsonHelper.UnmarshalJsonFile(DATABASE_CONFIG_FILE, &cfg); err != nil {
		return err
	}
	db, err := gorm.Open("sqlite3", "file:"+cfg.DatabaseFile+"?cache=shared&mode=rwc")
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

func (ds *DataSource) GetTx() *gorm.DB {
	return ds.tx
}

func (ds *DataSource) DoInTransaction(callback func(ds *DataSource) error) error {
	ds.tx = ds.db.Begin()
	if err := callback(ds); err != nil {
		ds.tx.Rollback()
		return err
	}
	if err := ds.tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
