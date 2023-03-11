package db

import (
	"errors"
	"fmt"

	common "reviewmakerback/common"

	"gorm.io/gorm"
)

func GetUser(id string, selectText string) (User, *gorm.DB) {
	var user User

	tx := Db.Select(selectText).Where("user_id = ?", id).Find(&user)
	return user, tx
}

func ExistsUser(id string) bool {
	var cnt int64
	_, tx := GetUser(id, "user_id")
	tx.Count(&cnt)
	return cnt == 1
}

func ExistsUserTId(tid string) (bool, User) {
	var user User
	var cnt int64

	Db.Where("twitter_id = ?", tid).Find(&user).Count(&cnt)
	return cnt > 0, user
}

func ExistsUserGId(gid string) (bool, User) {
	var user User
	var cnt int64

	Db.Where("google_id = ?", gid).Find(&user).Count(&cnt)
	return cnt > 0, user
}

func CreateUser(service string, name string, profile string, iconUrl string, twitterId string, twitterUserName string, googleId string, googleEmail string, requestIp string) (User, error) {
	var id string
	var err error
	if service == "twitter" {
		if f, u := ExistsUserTId(twitterId); f {
			return u, errors.New("指定されたTwitterIDは登録済みです")
		}
	} else if service == "google" {
		if f, u := ExistsUserGId(googleId); f {
			return u, errors.New("指定されたGoogleIDは登録済みです")
		}
	}

	for i := 0; i < retryCreateCnt; i++ {
		// ランダムな文字列を生成して、IDにする
		id, err = common.MakeRandomChars(idSize, twitterId)
		if err != nil {
			return User{}, err
		}
		if !ExistsUser(id) {
			user := User{
				UserId:           id,
				Name:             name,
				Profile:          profile,
				IconUrl:          iconUrl,
				AllowTwitterLink: false,
				KeepSession:      7200,
				TwitterId:        twitterId,
				TwitterUserName:  twitterUserName,
				GoogleId:         googleId,
				GoogleEmail:      googleEmail,
			}
			tx := Db.Create(&user)

			if tx.Error != nil {
				WriteErrorLog("", requestIp, "pusr-005", "ユーザーの作成に失敗しました", fmt.Sprintf("使用したID(%s) %s", id, tx.Error.Error()))
				return User{}, errors.New("ユーザー作成に失敗しました")
			}
			return user, nil
		}
	}
	return User{}, errors.New("ユーザー作成の試行回数が上限に達しました")
}

func UpdateUser(user User, name string, profile string, iconUrl string, iconIsChanged bool, allowTwitterLink bool, keepSession int) error {
	var tx *gorm.DB
	user.Name = name
	user.Profile = profile
	if iconIsChanged {
		user.IconUrl = iconUrl
	}
	user.AllowTwitterLink = allowTwitterLink
	user.KeepSession = keepSession
	tx = Db.Save(&user)
	return tx.Error
}
