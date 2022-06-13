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

		// 一時セッションの生存期間が終了したデータを削除
		db.Db.Find(&tempSession).Where("access_time < ?", time.Now().Add(-tempSessionDelSpan*time.Second)).Delete(&tempSession)
	}
}
