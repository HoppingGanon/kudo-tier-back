package db

import (
	"errors"
	"time"

	common "reviewmakerback/common"

	"github.com/labstack/echo"
	"gorm.io/gorm"
)

// ユーザー作成に失敗した際の再試行回数
const retryCreateCnt = 3

// ユーザーIDの桁数
const idSize = 16

func WriteOperationLog(id string, ipAddress string, operation string) {
	// ログを記録
	log := OperationLog{
		UserId:    id,
		IpAddress: ipAddress,
		Operation: operation,
		CreatedAt: time.Now(),
	}

	// データベースに登録
	Db.Create(log)
}

func WriteErrorLog(id string, ipAddress string, errorId string, operation string, descriptions string) {
	// ログを記録
	log := ErrorLog{
		UserId:       id,
		IpAddress:    ipAddress,
		ErrorId:      errorId,
		Operation:    operation,
		Descriptions: descriptions,
		CreatedAt:    time.Now(),
	}

	// データベースに登録
	Db.Create(log)
}

func GetUser(id string) (User, *gorm.DB) {
	var user User

	tx := Db.Find(&user).Where("user_id = ?", id)
	return user, tx
}

func ExistsUser(id string) bool {
	var cnt int64
	_, tx := GetUser(id)
	tx.Count(&cnt)
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
		id, err = common.MakeRandomChars(idSize, TwitterName)
		if err != nil {
			return "", err
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
				return "", tx.Error
			}

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

func MakeSession(seed string) (string, error) {
	var session Session
	var cnt int64
loop:
	for i := 0; i < retryCreateCnt; i++ {
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

func GetTier(tid string, uid string) (Tier, *gorm.DB) {
	var tier Tier

	tx := Db.Where("tier_id = ? and user_id = ?", tid, uid).Find(&tier)
	return tier, tx
}

func ExistsTier(tid string, uid string) bool {
	var cnt int64

	_, tx := GetTier(tid, uid)

	tx.Count(&cnt)
	return cnt == 1
}

func CreateTierId(userId string) (string, error) {
	var id string
	var err error
	for i := 0; i < retryCreateCnt; i++ {
		// ランダムな文字列を生成して、IDにする
		id, err = common.MakeRandomChars(idSize, userId)
		if err != nil {
			return "", err
		}
		if !ExistsTier(id, userId) {
			return id, err
		}
	}
	return "", err
}

func CreateTier(
	userId string,
	tierId string,
	name string,
	// 画像の保存パス、NULLなら変更しない
	path string,
	parags string,
	pointType string,
	reviewFactorParams string,
) error {
	var tier Tier
	if path == "nochange" {
		tier = Tier{
			TierId:       tierId,
			UserId:       userId,
			Name:         name,
			ImageUrl:     "",
			Parags:       parags,
			PointType:    pointType,
			FactorParams: reviewFactorParams,
		}
	} else {
		tier = Tier{
			TierId:       tierId,
			UserId:       userId,
			Name:         name,
			ImageUrl:     path,
			Parags:       parags,
			PointType:    pointType,
			FactorParams: reviewFactorParams,
		}
	}
	tx := Db.Create(&tier)
	return tx.Error
}

func UpdateTier(
	tier Tier,
	userId string,
	tierId string,
	name string,
	// 画像の保存パス、"nochange"なら変更しない
	imageUrl string,
	parags string,
	pointType string,
	factorParams string,
) error {
	var tx *gorm.DB
	tier.TierId = tierId
	tier.Name = name
	tier.Parags = parags
	tier.PointType = pointType
	tier.FactorParams = factorParams
	tier.UpdatedAt = time.Now()
	if imageUrl != "nochange" {
		tier.ImageUrl = imageUrl
	}
	tx = Db.Save(&tier)
	return tx.Error
}
