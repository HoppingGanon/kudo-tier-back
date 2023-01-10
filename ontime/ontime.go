package ontime

import (
	db "reviewmakerback/db"
	"time"
)

func DeleteTempSession() {
	for range time.Tick(db.TempSessionAlive * time.Second) {
		db.ArrangeSession()
	}
}
