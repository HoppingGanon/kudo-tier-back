package rest

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"strconv"

	"github.com/labstack/echo"

	common "reviewmakerback/common"
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

	// Twitterからユーザー情報の取得
	b, err = getTwitterApi("https://api.twitter.com/2/users/me?user.fields=profile_image_url", session.TwitterToken)
	if err != nil {
		return c.JSON(403, MakeError("pusr-004", "Twitterからユーザー情報が取得できませんでした"))
	}
	var twitterUser TwitterUser
	err = json.Unmarshal(b, &twitterUser)
	if err != nil {
		return c.JSON(403, MakeError("pusr-005", "Twitterから取得したユーザー情報が不正です"))
	}

	user, err := db.CreateUser(twitterUser.Data.Id, userData.Name, userData.Profile, "")
	if err != nil {
		return c.JSON(400, MakeError("pusr-006", "ユーザーの作成に失敗しました"))
	}

	// 画像データの名前を生成
	code, err := common.MakeRandomChars(16, user.UserId)
	if err != nil {
		return c.JSON(400, MakeError("pusr-007", "レビューアイコンの保存に失敗しました しばらく時間を開けて実行してください"))
	}
	fname := "icon_" + code + ".jpg"

	// 画像の保存
	path, er := savePicture(user.UserId, "user", "user", fname, "", userData.IconBase64, "prev-009", reviewValidation.iconMaxEdge, reviewValidation.iconAspectRate, 92)
	if er != nil {
		return c.JSON(400, er)
	}

	db.UpdateUser(user, userData.Name, userData.Profile, path, false)

	requestIp := net.ParseIP(c.RealIP()).String()
	db.WriteOperationLog(user.UserId, requestIp, "pusr", "")

	return c.JSON(200, UserData{
		UserId:           user.UserId,
		IsSelf:           true,
		TwitterName:      twitterUser.Data.Id,
		Name:             userData.Name,
		Profile:          userData.Profile,
		IconUrl:          path,
		AllowTwitterLink: false,
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

	var cnt int64
	user, tx := db.GetUser(uid, "*")
	if err != nil {
		return c.JSON(400, MakeError("uusr-003", "ユーザーの更新に失敗しました"))
	}
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(400, MakeError("uusr-004", "ユーザーの更新に失敗しました"))
	}

	// 画像データの名前を生成
	code, err := common.MakeRandomChars(16, user.UserId)
	if err != nil {
		return c.JSON(400, MakeError("uusr-005", "レビューアイコンの保存に失敗しました しばらく時間を開けて実行してください"))
	}
	fname := "icon_" + code + ".jpg"

	// 画像の保存
	path, er := savePicture(user.UserId, "user", "user", fname, user.IconUrl, userData.IconBase64, "uusr-006", reviewValidation.iconMaxEdge, reviewValidation.iconAspectRate, 92)
	if er != nil {
		return c.JSON(400, er)
	}

	db.UpdateUser(user, userData.Name, userData.Profile, path, userData.AllowTwitterLink)

	requestIp := net.ParseIP(c.RealIP()).String()
	db.WriteOperationLog(user.UserId, requestIp, "uusr", "")

	return c.String(200, uid)
}

// ユーザーデータを取得するGETリクエストの処理
func getReqUserData(c echo.Context) error {
	// 送信元ユーザーと参照先ユーザーが同じかどうかチェック
	session, err := db.CheckSession(c)

	var existsSession bool
	if err != nil {
		// セッションが存在しない
		existsSession = false
	}

	user := db.User{}
	var cnt int64

	uid := c.Param("uid")
	user, tx := db.GetUser(uid, "*")
	tx.Count(&cnt)

	if cnt != 1 {
		return c.JSON(404, MakeError("gusr-001", "ユーザーが存在しません"))
	}

	userData := UserData{
		UserId:           user.UserId,
		IsSelf:           existsSession && session.UserId == user.UserId,
		IconUrl:          user.IconUrl,
		TwitterName:      "",
		Name:             user.Name,
		Profile:          user.Profile,
		AllowTwitterLink: user.AllowTwitterLink,
		ReviewsCount:     db.GetReviewCountInUser(user.UserId),
		TiersCount:       db.GetTierCountInUser(user.UserId),
	}

	if err == nil && uid == session.UserId {
		// 送信元ユーザーと参照先ユーザーが同じ場合
		userData.IsSelf = true
		userData.TwitterName = user.TwitterName
	} else {
		// 送信元ユーザーと参照先ユーザーが異なる場合またはそもそもセッションが無い場合
		userData.IsSelf = false
		if userData.AllowTwitterLink {
			userData.TwitterName = user.TwitterName
		}
	}
	return c.JSON(200, userData)
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
