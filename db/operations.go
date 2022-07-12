package db

import (
	"errors"
	"time"

	common "reviewmakerback/common"

	"github.com/labstack/echo"
)

// ユーザー作成に失敗した際の再試行回数
const retryCreateCnt = 3

// ユーザーIDの桁数
const userIdSize = 48

func WriteAccessLog(id string, ipAddress string, accessTime time.Time, operation string) {
	// ログを記録
	log := OperationLog{
		UserId:     id,
		IpAddress:  ipAddress,
		AccessTime: accessTime,
		Operation:  operation,
	}

	// データベースに登録
	Db.Create(log)
}

func ExistsUserId(id string) bool {
	var user User
	var cnt int64

	Db.Find(&user).Where("user_id = ?", id).Count(&cnt)
	return cnt == 1
}

func ExistsUserTId(tid string) bool {
	var user User
	var cnt int64

	Db.Find(&user).Where("twitter_name = ?", tid).Count(&cnt)
	return cnt == 1
}

func CreateUser(TwitterName string, name string, profile string, iconUrl string) (string, error) {
	var id string
	var err error
	if ExistsUserTId(TwitterName) {
		return "", errors.New("指定されたTwitterIDは登録済みです")
	}

	for i := 0; i < retryCreateCnt; i++ {
		// ランダムな文字列を生成して、IDにする
		id, err = common.MakeRandomChars(userIdSize, TwitterName)
		if err != nil {
			return "", err
		}
		if !ExistsUserId(id) {
			user := User{
				TwitterName: TwitterName,
				UserId:      id,
				Name:        name,
				Profile:     profile,
				IconUrl:     iconUrl,
			}
			Db.Create(&user)

			return id, nil
		}
	}
	return "", errors.New("ユーザー作成の試行回数が上限に達しました")
}

func CheckSession(c echo.Context) (Session, error) {
	sessionId := c.Request().Header.Get("sessionId")
	var session Session
	var cnt int64
	Db.Where("session_id = ?", sessionId).Find(&session).Count(&cnt)
	if cnt == 1 {
		return session, nil
	}
	return session, errors.New("セッションがありません")
}

func MakeSession(retryCount int, seed string) (string, error) {
	var session Session
	var cnt int64
loop:
	for i := 0; i < retryCount; i++ {
		sessionId, err := common.MakeSession(seed)
		if err != nil {
			continue loop
		}
		Db.Find(&session).Where("session_id = ?", sessionId).Count(&cnt)
		if cnt == 0 {
			return sessionId, nil
		}
	}
	return "", errors.New("セッション作成に失敗")
}
