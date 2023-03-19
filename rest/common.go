package rest

import (
	"encoding/json"
	"io/ioutil"
	"reviewmakerback/common"
	"reviewmakerback/db"
	"strconv"

	"github.com/labstack/echo"
)

const notificationsLimit = 100

// 通知リストを送信
func getNotifications(c echo.Context) error {

	// セッションの存在チェック
	session, err := db.CheckSession(c, true, false)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	dbNotifications, tx := db.GetNotifications(session.UserId, notificationsLimit)
	if tx.Error != nil {
		return c.JSON(400, MakeError("gnts-001", "通知情報が取得できません"))
	}

	notifications := make([]Notification, len(dbNotifications))
	for i, n := range dbNotifications {
		notifications[i] = Notification{
			Id:          n.Id,
			Content:     n.Content,
			IsRead:      n.IsRead,
			IsImportant: n.IsImportant,
			Url:         n.Url,
			CreatedAt:   common.DateToString(n.CreatedAt),
		}
	}

	return c.JSON(200, notifications)
}

func getNotificationsCount(c echo.Context) error {

	// セッションの存在チェック
	session, err := db.CheckSession(c, true, false)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	cnt, tx := db.GetNotificationsCount(session.UserId, notificationsLimit)
	if tx.Error != nil {
		return c.JSON(400, MakeError("gntc-001", "通知情報の数が取得できません"))
	}

	return c.JSON(200, CountData{
		Count: cnt,
	})
}

func updateNotificationRead(c echo.Context) error {
	// セッションの存在チェック
	session, err := db.CheckSession(c, true, false)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}
	var nr NotificationReadData
	err = json.Unmarshal(b, &nr)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}

	nid, err := strconv.Atoi(c.Param("nid"))
	if err != nil {
		return c.JSON(400, MakeError("untr-001", "指定されたIDが不正です"))
	}

	err = db.UpdateNotificationRead(session.UserId, uint(nid), nr.IsRead)
	if err != nil {
		return c.JSON(400, MakeError("untr-002", "通知既読状態の更新に失敗しました"))
	}
	return c.NoContent(204)
}
