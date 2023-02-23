package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"gorm.io/gorm"

	"reviewmakerback/common"
	db "reviewmakerback/db"
)

const latestPostMax = 100

// ユーザー作成のためのPOSTリクエストの処理
func postReqUser(c echo.Context) error {
	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}

	var userData UserCreatingData
	err = json.Unmarshal(b, &userData)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}

	// バリデーションチェック
	if !userData.Accept {
		return c.JSON(400, MakeError("pusr-001", "利用規約への同意は必須です"))
	}
	f, er := validText("表示名", "pusr-002", userData.Name, true, 0, userValidation.nameLenMax, "", "")
	if !f {
		return c.JSON(400, er)
	}
	f, er = validText("プロフィール", "pusr-003", userData.Profile, false, 0, userValidation.profileLenMax, "", "")
	if !f {
		return c.JSON(400, er)
	}

	var twitterId string
	var googleId string

	if session.LoginService == "twitter" {
		twitterId = session.ServiceId
	} else if session.LoginService == "google" {
		googleId = session.ServiceId
	}

	// アイコンはとりあえず設定しない
	user, err := db.CreateUser(
		session.LoginService,
		userData.Name,
		userData.Profile,
		"",
		twitterId,
		session.TwitterUserName,
		googleId,
		session.GoogleEmail,
	)
	if err != nil {
		requestIp := net.ParseIP(c.RealIP()).String()
		db.WriteErrorLog(session.UserId, requestIp, "pusr-006", "ユーザーの作成に失敗しました", err.Error())
		return c.JSON(400, MakeError("pusr-006", "ユーザーの作成に失敗しました"))
	}

	session.UserId = user.UserId
	if db.Db.Save(&session).Error != nil {
		// エラー処理なし
	}

	// 画像の保存
	path, er := savePicture(user.UserId, "user", "user", "icon_", "", userData.IconBase64, "prev-009", reviewValidation.iconMaxEdge, reviewValidation.iconAspectRate, 92)
	if er != nil {
		return c.JSON(400, er)
	}

	// 後からアイコンを変更する
	db.UpdateUser(user, userData.Name, userData.Profile, path, true, false, 3600)

	requestIp := net.ParseIP(c.RealIP()).String()
	db.WriteOperationLog(user.UserId, requestIp, "pusr", "")

	return c.JSON(200, SelfUserData{
		UserId:           user.UserId,
		IsSelf:           true,
		TwitterId:        twitterId,
		Name:             userData.Name,
		Profile:          userData.Profile,
		IconUrl:          path,
		AllowTwitterLink: false,
		KeepSession:      3600,
		ReviewsCount:     0,
		TiersCount:       0,
	})
}

// ユーザーデータの更新のためのUPDATEリクエストの処理
func updateReqUser(c echo.Context) error {
	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	uid := c.Param("uid")

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}

	var userData UserEditingData
	err = json.Unmarshal(b, &userData)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}

	// 編集ユーザーとTier所有ユーザーチェック
	if session.UserId != uid {
		return c.JSON(403, commonError.userNotEqual)
	}

	// バリデーションチェック
	f, er := validText("表示名", "uusr-001", userData.Name, true, 0, userValidation.nameLenMax, "", "")
	if !f {
		return c.JSON(400, er)
	}
	f, er = validText("プロフィール", "uusr-002", userData.Profile, false, 0, userValidation.profileLenMax, "", "")
	if !f {
		return c.JSON(400, er)
	}
	f, er = validInteger("セッション保持時間", "uusr-003", userData.KeepSession, 10, 1440)
	if !f {
		return c.JSON(400, er)
	}

	var cnt int64
	user, tx := db.GetUser(uid, "*")
	if err != nil {
		return c.JSON(400, MakeError("uusr-004", "ユーザーの更新に失敗しました"))
	}
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(400, MakeError("uusr-005", "ユーザーの更新に失敗しました"))
	}

	// 画像データの名前を生成
	path := ""
	if userData.IconIsChanged {
		// 画像の保存
		path, er = savePicture(user.UserId, "user", "user", "icon_", user.IconUrl, userData.IconBase64, "uusr-006", reviewValidation.iconMaxEdge, reviewValidation.iconAspectRate, 92)
		if er != nil {
			return c.JSON(400, er)
		}
	}

	db.UpdateUser(user, userData.Name, userData.Profile, path, userData.IconIsChanged, userData.AllowTwitterLink, userData.KeepSession*60)

	requestIp := net.ParseIP(c.RealIP()).String()
	db.WriteOperationLog(user.UserId, requestIp, "uusr", "")

	return c.String(200, uid)
}

// ユーザーデータを取得するGETリクエストの処理
func getReqUserData(c echo.Context) error {
	// 送信元ユーザーと参照先ユーザーが同じかどうかチェック
	session, err := db.CheckSession(c)

	existsSession := err == nil

	user := db.User{}
	var cnt int64

	uid := c.Param("uid")
	user, tx := db.GetUser(uid, "*")
	tx.Count(&cnt)

	if cnt != 1 {
		return c.JSON(404, MakeError("gusr-001", "ユーザーが存在しません"))
	}

	if existsSession && uid == session.UserId {
		// 送信元ユーザーと参照先ユーザーが同じ場合

		selfUserData := SelfUserData{
			UserId:           user.UserId,
			IsSelf:           true,
			IconUrl:          user.IconUrl,
			TwitterId:        user.TwitterId,
			Name:             user.Name,
			Profile:          user.Profile,
			AllowTwitterLink: user.AllowTwitterLink,
			KeepSession:      user.KeepSession / 60,
			ReviewsCount:     db.GetReviewCountInUser(user.UserId),
			TiersCount:       db.GetTierCountInUser(user.UserId),
		}

		return c.JSON(200, selfUserData)
	} else {

		userData := UserData{
			UserId:           user.UserId,
			IsSelf:           false,
			IconUrl:          user.IconUrl,
			TwitterId:        "",
			Name:             user.Name,
			Profile:          user.Profile,
			AllowTwitterLink: user.AllowTwitterLink,
			ReviewsCount:     db.GetReviewCountInUser(user.UserId),
			TiersCount:       db.GetTierCountInUser(user.UserId),
		}

		// 送信元ユーザーと参照先ユーザーが異なる場合またはそもそもセッションが無い場合
		userData.IsSelf = false
		if userData.AllowTwitterLink {
			userData.TwitterId = user.TwitterId
		}
		return c.JSON(200, userData)
	}
}

func getReqLatestPostLists(c echo.Context) error {
	uid := c.Param("uid")

	length, err := strconv.Atoi(c.QueryParam("length"))

	if err != nil {
		return c.JSON(400, MakeError("gpls-001", "ページ指定が異常です"))
	} else {
		f, er := validInteger("一度に取得できる投稿の件数上限", "gpls-001", length, 0, latestPostMax)
		if !f {
			return c.JSON(400, er)
		}
	}

	if !db.ExistsUser(uid) {
		return c.JSON(404, MakeError("gpls-004", "ユーザーが存在しません"))
	}

	var tiers []db.Tier
	db.Db.Select("tier_id, name").Where("user_id = ?", uid).Order("updated_at desc").Limit(length).Find(&tiers)

	var reviews []db.Review
	db.Db.Select("review_id, name").Where("user_id = ?", uid).Order("updated_at desc").Limit(length).Find(&reviews)

	postListData := PostListsData{
		Tiers:   make([]PostListItem, len(tiers)),
		Reviews: make([]PostListItem, len(reviews)),
	}

	for i, tier := range tiers {
		postListData.Tiers[i] = PostListItem{
			Id:   tier.TierId,
			Name: tier.Name,
		}
	}

	for i, review := range reviews {
		postListData.Reviews[i] = PostListItem{
			Id:   review.ReviewId,
			Name: review.Name,
		}
	}
	return c.JSON(200, postListData)
}

// ユーザー削除の際のステップ1
func deleteUser1(c echo.Context) error {
	uid := c.Param("uid")

	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	// 編集ユーザーとTier所有ユーザーチェック
	if session.UserId != uid {
		return c.JSON(403, commonError.userNotEqual)
	}

	session.DeleteCodeTime = time.Now()

	delcode := common.Substring(common.GetSHA256(session.SessionID+common.DateToString(session.DeleteCodeTime)), 0, 6)

	if db.Db.Save(&session).Error != nil {
		return c.JSON(400, MakeError("dus1-001", "削除コードの発行に失敗しました"))
	}

	requestIp := net.ParseIP(c.RealIP()).String()
	db.WriteOperationLog(session.UserId, requestIp, "dus1", "")

	return c.String(202, delcode)
}

// ユーザー削除の際のステップ1
func deleteUser2(c echo.Context) error {
	uid := c.Param("uid")
	delcode := c.QueryParam("delcode")
	requestIp := net.ParseIP(c.RealIP()).String()

	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	// 編集ユーザーとTier所有ユーザーチェック
	if session.UserId != uid {
		return c.JSON(403, commonError.userNotEqual)
	}

	if delcode != common.Substring(common.GetSHA256(session.SessionID+common.DateToString(session.DeleteCodeTime)), 0, 6) {
		return c.JSON(400, MakeError("dus2-001", "削除コードが一致しません"))
	} else if session.DeleteCodeTime.Add(time.Duration(60) * time.Second).Before(time.Now()) {
		return c.JSON(400, MakeError("dus2-002", "削除コードの期限が切れています"))
	}

	result := db.Db.Transaction(func(tx *gorm.DB) error {
		var user db.User
		var cnt int64
		tdb := tx.Where("user_id = ?", session.UserId).Find(&user)
		if tdb.Error != nil {
			return tdb.Error
		}
		if tdb.Count(&cnt); cnt != 1 {
			return errors.New("ユーザーが存在しません")
		}

		tdb = tx.Where("user_id = ?", session.UserId).Delete(&db.Tier{})
		if tdb.Error != nil {
			return tdb.Error
		}

		tdb = tx.Where("user_id = ?", session.UserId).Delete(&db.Review{})
		if tdb.Error != nil {
			return tdb.Error
		}

		tdb = tx.Where("user_id = ?", session.UserId).Delete(&db.Session{})
		if tdb.Error != nil {
			return tdb.Error
		}

		tdb = tx.Where("user_id = ?", session.UserId).Delete(&db.User{})
		if tdb.Error != nil {
			return tdb.Error
		}
		return nil
	})

	if result != nil {
		db.WriteErrorLog(session.UserId, requestIp, "dus2-01", "ユーザーの削除に失敗しました", result.Error())
		return c.JSON(400, MakeError("dus2-003", "ユーザーの削除に失敗しました"))
	}

	// 全ファイルを削除するが、エラーが起こっても中断せず記録のみ残す
	err = os.RemoveAll((fmt.Sprintf("%s/%s", os.Getenv("BACK_AP_FILE_PATH"), session.UserId)))
	if err != nil && !os.IsNotExist(err) {
		db.WriteErrorLog(session.UserId, requestIp, "dus2-05", "フォルダが削除できませんでした", fmt.Sprintf("'%s/%s' %s", os.Getenv("BACK_AP_FILE_PATH"), session.UserId, err.Error()))
	}

	db.WriteOperationLog(session.UserId, requestIp, "dus1", "")

	return c.NoContent(204)
}
