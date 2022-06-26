package ontime

import (
	db "reviewmakerback/db"
	"time"
)

const tempSessionAlive = 60
const tempSessionDelSpan = 60

func DeleteTempSession() {
	for range time.Tick(tempSessionAlive * time.Second) {
		var tempSession db.TempSession
		var session db.Session

		// 一時セッションの生存期間が終了したデータを削除
		db.Db.Where("access_time < ?", time.Now().Add(-tempSessionDelSpan*time.Second)).Delete(&tempSession)
		// セッションの生存期間が終了したデータを削除
		db.Db.Where("expired_time < ?", time.Now()).Delete(&session)
	}
}
