package db

import (
	"errors"

	"gorm.io/gorm"
)

func GetNotifications(userId string, limit int) ([]NotificationJoinRead, *gorm.DB) {
	var notifications []NotificationJoinRead

	db1 := Db.Order("created_at DESC").Limit(limit).Model(&Notification{})

	tx := Db.Select("id, content, is_important, url, r.is_read as is_read, created_at").Table("(?) as t", db1)
	tx = tx.Joins("left join notification_reads as r on r.notification_id = t.id and r.user_id = ?", userId)
	tx = tx.Scan(&notifications)

	// Gormではbool型の項目でのnullはfalseになる
	return notifications, tx
}

func GetNotificationsCount(userId string, limit int) (int64, *gorm.DB) {
	var cnt int64

	db1 := Db.Order("created_at DESC").Limit(limit).Model(&Notification{})

	db2 := Db.Select("content, is_important, url, r.is_read as is_read, created_at").Table("(?) as t", db1)
	db2 = db2.Joins("left join notification_reads as r on r.notification_id = t.id and r.user_id = ?", userId)

	// Postgresのみ可能な文
	db3 := Db.Table("(?) as t2", db2).Where("COALESCE(t2.is_read, ?) = ?", false, false).Count(&cnt)

	return cnt, db3
}

// 通知の既読情報を更新する
func UpdateNotificationRead(userId string, notificationId uint, isRead bool) error {
	var cnt int64
	tx := Db.Where("id = ?", notificationId).Model(&Notification{}).Count(&cnt)
	if tx.Error != nil {
		return tx.Error
	} else if cnt != 1 {
		return errors.New("指定された通知は存在しません")
	}

	var nr NotificationRead
	tx = Db.Where("notification_id = ? and user_id = ?", notificationId, userId).Find(&nr).Count(&cnt)

	if tx.Error != nil {
		return tx.Error
	} else if cnt == 0 {
		return Db.Create(NotificationRead{
			NotificationId: notificationId,
			UserId:         userId,
			IsRead:         isRead,
		}).Error
	} else {
		return tx.Update("is_read", isRead).Error
	}

}
