package main

import (
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	db "reviewmakerback/db"
	"reviewmakerback/ontime"
	rest "reviewmakerback/rest"
)

func main() {
	// 環境変数の読み込み
	loadEnv()

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
	e.Logger.Fatal(e.Start(":" + os.Getenv("AP_PORT")))
}

// envLoad 環境変数のロード
func loadEnv() {
	// 開発環境のファイルを読み込む
	err := godotenv.Load(".env.local")
	if err != nil {
		// もしファイルがなければ、ローカル環境ファイルを読み込む
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal(".env.localおよび.envが見つかりませんでした")
		}
	}
}

func loggingSettings(filename string) {
	// ログ出力先を指定
	logfile, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	multiLogFile := io.MultiWriter(os.Stdout, logfile)
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	log.SetOutput(multiLogFile)
}
