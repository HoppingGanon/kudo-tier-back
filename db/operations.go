package db

import (
	"errors"
	"time"

	common "reviewmakerback/common"
)

// ユーザー作成に失敗した際の再試行回数
const retryCreateCnt = 3

// ユーザーIDの桁数
const userIdSize = 16

func WriteAccessLog(id string, ipAddress string, accessTime time.Time, operation string) {
	// ログを記録
	log := AccessLog{
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

	Db.Find(&user).Where("id = ?", id).Count(&cnt)
	return cnt == 1
}

func ExistsUserTId(tid string) bool {
	var user User
	var cnt int64

	Db.Find(&user).Where("twitter_id = ?", tid).Count(&cnt)
	return cnt == 1
}

func CreateUser(twitterId string, createTime time.Time, name string, isNew string) error {
	var id string
	if ExistsUserTId(twitterId) {
		return errors.New("指定されたTwitterIDは登録済みです")
	}

	for i := 0; i < retryCreateCnt; i++ {
		id = common.MakeRandomChars(userIdSize)
		if !ExistsUserId(id) {
			user := User{
				Id:             "",
				TwitterId:      twitterId,
				CreationTime:   createTime,
				LastAccessTime: createTime,
				Name:           name,
				IsNew:          isNew,
			}
			Db.Create(user)

			return nil
		}
	}
	return errors.New("ユーザー作成の試行回数が上限に達しました")
}
