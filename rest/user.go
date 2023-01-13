package rest

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"

	"github.com/labstack/echo"

	common "reviewmakerback/common"
	db "reviewmakerback/db"
	"unicode/utf8"
)

const latestPostMax = 100

// ユーザー作成のためのPOSTリクエストの処理
func postReqUser(c echo.Context) error {
	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.String(403, "セッションがありません")
	}

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)

	if err != nil {
		return c.String(400, "JSONデータが不正です")
	}

	var userData UserEdittingData
	err = json.Unmarshal(b, &userData)
	if err != nil {
		return c.String(400, "不正なユーザーデータです")
	}

	// バリデーションチェック
	if !userData.Accept {
		return c.String(400, "利用規約への同意は必須です")
	}
	cnt := utf8.RuneCountInString(userData.Name)
	if cnt < 1 || cnt > 64 {
		return c.String(400, "名前が不正です")
	}
	cnt = utf8.RuneCountInString(userData.Profile)
	if cnt > 200 {
		return c.String(400, "プロフィールが不正です")
	}

	// Twitterからユーザー情報の取得
	b, err = getTwitterApi("https://api.twitter.com/2/users/me?user.fields=profile_image_url", session.TwitterToken)
	if err != nil {
		return c.String(403, "Twitterからユーザー情報が取得できませんでした")
	}
	var twitterUser TwitterUser
	err = json.Unmarshal(b, &twitterUser)
	if err != nil {
		return c.String(403, "Twitterから取得したユーザー情報が不正です")
	}

	user, err := db.CreateUser(twitterUser.Data.Id, userData.Name, userData.Profile, "")
	if err != nil {
		return c.String(400, "ユーザーの作成に失敗しました")
	}

	// 画像データの名前を生成
	code, err := common.MakeRandomChars(16, user.UserId)
	if err != nil {
		return c.JSON(400, MakeError("prev-008", "レビューアイコンの保存に失敗しました しばらく時間を開けて実行してください"))
	}
	fname := "icon_" + code + ".jpg"

	// 画像の保存
	path, er := savePicture(user.UserId, "user", "user", fname, "", userData.IconBase64, "prev-009", reviewValidation.iconMaxEdge, reviewValidation.iconAspectRate, 92)
	if er != nil {
		return c.JSON(400, er)
	}

	println("userid:" + user.UserId)
	println("path:" + path)

	db.UpdateUser(user, userData.Name, userData.Profile, path)

	requestIp := net.ParseIP(c.RealIP()).String()
	db.WriteOperationLog(user.UserId, requestIp, "login")

	return c.JSON(200, UserData{
		UserId:      user.UserId,
		IsSelf:      true,
		TwitterName: twitterUser.Data.Id,
		Name:        userData.Name,
		Profile:     userData.Profile,
		IconUrl:     path,
		ReviewCount: 0,
		TierCount:   0,
	})
}

// ユーザーデータの更新のためのUPDATEリクエストの処理
func updateReqUser(c echo.Context) error {

	// リクエストのURIからIDを取得
	requestId := c.Param("id")

	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.String(404, "session not exists")
	}

	// リクエストのIDとセッションのIDを比較して、一致してなければエラー
	if requestId != session.SessionID {
		return c.String(403, "Unauthorized operation")
	}

	// Twitterからユーザー情報の取得
	b, err := getTwitterApi("https://api.twitter.com/2/users/me?user.fields=profile_image_url", session.TwitterToken)
	if err != nil {
		return c.String(http.StatusForbidden, "Twitterからユーザー情報が取得できませんでした")
	}
	var twitterUser TwitterUser
	err = json.Unmarshal(b, &twitterUser)
	if err != nil {
		return c.String(http.StatusForbidden, "Twitterから取得したユーザー情報が不正です")
	}

	// db.Db.Update()

	return c.String(200, "")
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

	res := UserData{
		UserId:      user.UserId,
		IsSelf:      existsSession && session.UserId == user.UserId,
		IconUrl:     user.IconUrl,
		TwitterName: "",
		Name:        user.Name,
		Profile:     user.Profile,
		ReviewCount: db.GetReviewCountInUser(user.UserId),
		TierCount:   db.GetTierCountInUser(user.UserId),
	}

	if err == nil && uid == session.UserId {
		// 送信元ユーザーと参照先ユーザーが同じ場合
		res.IsSelf = true
		res.TwitterName = user.TwitterName
	} else {
		// 送信元ユーザーと参照先ユーザーが異なる場合またはそもそもセッションが無い場合
		res.IsSelf = false
	}
	return c.JSON(200, res)
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
