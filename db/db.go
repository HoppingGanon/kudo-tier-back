package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// データベース
var Db *gorm.DB

// 関数についても大文字で定義しないと外部から参照できない
func InitDb() *gorm.DB {
	Db = connectDB()
	if Db != nil {
		migrateDB()
	}
	return Db
}

func connectDB() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_TIMEZONE"))

	var err error
	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		println("データベース接続エラー")
		return nil
	}

	if Db == nil {
		println("データベース接続エラー")
		return nil
	}
	println("データベース接続を確認")

	return Db
}

func migrateDB() {
	if !Db.Migrator().HasTable(&TempSession{}) {
		Db.Migrator().CreateTable(&TempSession{})
		println("テーブル")
	}

	Db.AutoMigrate(
		&Session{},
		&TempSession{},
		&User{},
		&AccessLog{},
	)
	println("マイグレート")

}
