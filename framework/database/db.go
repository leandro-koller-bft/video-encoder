package database

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/leandro-koller-bft/video-encoder/domain"
	"github.com/leandro-koller-bft/video-encoder/local_constants"
	_ "github.com/lib/pq"
)

type Database struct {
	DB            *gorm.DB
	DSN           string
	DSNTest       string
	DBType        string
	DBTypeTest    string
	Debug         bool
	AutoMigrateDB bool
	Env           string
}

func NewDB() *Database {
	return &Database{}
}

func NewDBTest() *gorm.DB {
	dbInstance := NewDB()

	dbInstance.Env = local_constants.TEST_TOKEN
	dbInstance.DBTypeTest = "sqlite3"
	dbInstance.DSNTest = ":memory:"
	dbInstance.AutoMigrateDB = true
	dbInstance.Debug = true

	connection, err := dbInstance.Connect()
	if err != nil {
		log.Fatalf("Test db error: %v", err)
	}

	return connection
}

func (db *Database) Connect() (*gorm.DB, error) {
	var err error

	if db.Env != local_constants.TEST_TOKEN {
		db.DB, err = gorm.Open(db.DBType, db.DSN)
	} else {
		db.DB, err = gorm.Open(db.DBTypeTest, db.DSNTest)
	}
	if err != nil {
		return nil, err
	}

	if db.Debug {
		db.DB.LogMode(true)
	}

	if db.AutoMigrateDB {
		db.DB.AutoMigrate(&domain.Video{}, &domain.Job{})
		db.DB.Model(domain.Job{}).AddForeignKey("video_id", "videos (id)", "CASCADE", "CASCADE")
	}

	return db.DB, nil
}
