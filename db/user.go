package db

import (
	"errors"

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

func ExistsUserTId(tid string) bool {
	var user User
	var cnt int64

	Db.Where("twitter_name = ?", tid).First(&user).Count(&cnt)
	return cnt == 1
}

func CreateUser(TwitterName string, name string, profile string, iconUrl string) (User, error) {
	var id string
	var err error
	if ExistsUserTId(TwitterName) {
		return User{}, errors.New("指定されたTwitterIDは登録済みです")
	}

	for i := 0; i < retryCreateCnt; i++ {
		// ランダムな文字列を生成して、IDにする
		id, err = common.MakeRandomChars(idSize, TwitterName)
		if err != nil {
			return User{}, err
		}
		if !ExistsUser(id) {
			user := User{
				TwitterName: TwitterName,
				UserId:      id,
				Name:        name,
				Profile:     profile,
				IconUrl:     iconUrl,
			}
			tx := Db.Create(&user)

			if err != nil {
				return User{}, tx.Error
			}

			return user, nil
		}
	}
	return User{}, errors.New("ユーザー作成の試行回数が上限に達しました")
}

func UpdateUser(user User, name string, profile string, iconUrl string) error {
	var tx *gorm.DB
	user.Name = name
	user.Profile = profile
	user.IconUrl = iconUrl
	if iconUrl != "nochange" {
		user.IconUrl = iconUrl
	}
	tx = Db.Save(&user)
	return tx.Error
}
