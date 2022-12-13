package db

import (
	"errors"
	"image"
	"os"
	"time"

	bytes "bytes"
	b64 "encoding/base64"
	"image/jpeg"
	common "reviewmakerback/common"

	"github.com/labstack/echo"
	"github.com/nfnt/resize"
)

// ユーザー作成に失敗した際の再試行回数
const retryCreateCnt = 3

// ユーザーIDの桁数
const idSize = 16

// tierの画像サイズの最大
const tierImgMaxEdge = 1080

// tierの画像サイズの最大(KB)
const tierImgMaxBytes = 5000

// tierの画像サイズの最大

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
		id, err = common.MakeRandomChars(idSize, TwitterName)
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

func ExistsTier(tid string, uid string) bool {
	var tier Tier
	var cnt int64

	Db.Find(&tier).Where("tier_id = ? && user_id", tid, uid).Count(&cnt)
	return cnt == 1
}

func getTierId(userId string) (string, error) {
	for i := 0; i < retryCreateCnt; i++ {
		// ランダムな文字列を生成して、IDにする
		id, err := common.MakeRandomChars(idSize, userId)
		if err != nil {
			return "", err
		}
		if !ExistsTier(id, userId) {
			return id, nil
		}
	}
	return "", errors.New("再試行の上限に達しました")
}

func CreateTier(
	userId string,
	name string,
	imageBase64 string,
	parags []Parag,
	pointType string,
	reviewFactorParams []ReviewParam,
) (string, error) {
	var id string
	var err error

	// PointTypeのチェック
	if !IsPointType(pointType) {
		return "", errors.New("ポイント表示方法が異常です")
	}

	// 画像が既定のサイズ以下であることを確認する
	if len(imageBase64) < tierImgMaxBytes*1024*8/6 {
		return "", errors.New("画像のサイズが大きすぎます")
	}

	for i := 0; i < retryCreateCnt; i++ {
		// ランダムな文字列を生成して、IDにする
		id, err = common.MakeRandomChars(idSize, userId)
		if err != nil {
			return "", err
		}
		if !ExistsTier(id, userId) {
			// Base64文字列をバイト列に変換する
			byteAry, err := b64.StdEncoding.DecodeString(imageBase64)
			if err != nil {
				return "", err
			}

			// バイト列をReaderに変換
			r := bytes.NewReader(byteAry)
			img, _, err := image.Decode(r)

			// 画像サイズを取得
			w := img.Bounds().Dx()
			h := img.Bounds().Dy()

			if h < w {
				// 幅の方が大きい
				if tierImgMaxEdge < w {
					w = tierImgMaxEdge
					h = h * tierImgMaxEdge / w
				}
			} else {
				// 高さの方が大きい
				if tierImgMaxEdge < h {
					w = w * tierImgMaxEdge / h
					h = tierImgMaxEdge
				}
			}

			resizedImg := resize.Resize(uint(w), uint(h), img, resize.NearestNeighbor)

			err = os.MkdirAll(os.Getenv("AP_FILE_PATH")+"/"+userId+"/tier/"+id, os.ModePerm)
			if err != nil {
				return "", err
			}

			path := os.Getenv("AP_FILE_PATH") + "/" + userId + "/tier/" + id + "/icon.jpg"

			out, err := os.Create(path)
			if err != nil {
				return "", err
			}

			opts := &jpeg.Options{
				Quality: 92,
			}

			jpeg.Encode(out, resizedImg, opts)

			if err != nil {
				return "", err
			}

			tier := Tier{
				TierId:       id,
				UserId:       userId,
				Name:         name,
				ImageUrl:     path,
				Prags:        parags,
				PointType:    pointType,
				FactorParams: reviewFactorParams,
			}
			Db.Create(&tier)

			return id, nil
		}
	}
	return "", errors.New("Tier作成の試行回数が上限に達しました")
}
