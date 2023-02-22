package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	db "reviewmakerback/db"
	"reviewmakerback/ontime"
	rest "reviewmakerback/rest"
)

func main() {
	// 環境変数の読み込み
	CheckEnvs()

	// ログ出力場所の指定
	loggingSettings("echo.log")

	e := echo.New()

	// データベース接続・マイグレート
	db.InitDb()

	// 定期処理を登録
	go ontime.DeleteTempSession()

	// ミドルウェアからCORSの使用を設定する
	// これを設定しないと、同オリジンからのアクセスが拒否される
	e.Use(middleware.CORS())

	rest.Route(e)

	// リスナーポート番号
	e.Logger.Fatal(e.Start(":" + os.Getenv("BACK_AP_PORT")))
}

// 環境変数の必須チェック
func CheckEnvs() {
	checkEnv("BACK_DB_HOST")
	checkEnv("BACK_DB_PORT")
	checkEnv("BACK_DB_NAME")
	checkEnv("BACK_DB_USER")
	checkEnv("BACK_DB_PASSWORD")
	checkEnv("BACK_DB_TIMEZONE")
	checkEnv("BACK_TW_CLIENT_ID")
	checkEnv("BACK_TW_CLIENT_SEC")
	checkEnv("BACK_TW_REDIRECT_URI")
	checkEnv("BACK_TW1_APIKEY")
	checkEnv("BACK_TW1_APISECRET")
	checkEnv("BACK_TW1_ACCESSTOKEN")
	checkEnv("BACK_TW1_ACCESSSEC")
	checkEnv("BACK_AP_FILE_PATH")
	checkEnv("BACK_AP_PORT")
}

func checkEnv(name string) {
	if os.Getenv(name) == "" {
		panic(fmt.Sprintf("環境変数'%s'がありません", name))
	}
}

func loggingSettings(filename string) {
	// ログ出力先を指定
	logfile, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	multiLogFile := io.MultiWriter(os.Stdout, logfile)
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	log.SetOutput(multiLogFile)
}
