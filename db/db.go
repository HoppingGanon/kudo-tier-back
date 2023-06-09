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
		println("マイグレートを実行しました")
		ArrangeSession()
		println("最初のセッション整理を行いました")
	}
	return Db
}

// データベースに接続する関数
func connectDB() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		os.Getenv("BACK_DB_HOST"),
		os.Getenv("BACK_DB_USER"),
		os.Getenv("BACK_DB_PASSWORD"),
		os.Getenv("BACK_DB_NAME"),
		os.Getenv("BACK_DB_PORT"),
		os.Getenv("BACK_DB_TIMEZONE"))

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

// データベースのテーブルをマイグレートする関数
func migrateDB() {
	Db.AutoMigrate(
		&Session{},
		&TempSession{},
		&User{},
		&OperationLog{},
		&ErrorLog{},
		&Tier{},
		&Review{},
		&Notification{},
		&NotificationRead{},
	)
}
