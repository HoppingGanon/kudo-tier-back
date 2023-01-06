package rest

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/labstack/echo"

	db "reviewmakerback/db"
	"unicode/utf8"
)

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

	var userData InitUserData
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

	uid, err := db.CreateUser(twitterUser.Data.Id, userData.Name, userData.Profile, twitterUser.Data.ProfileImageUrl)
	if err != nil {
		return c.String(400, "ユーザーの作成に失敗しました")
	}

	requestIp := net.ParseIP(c.RealIP()).String()
	db.WriteOperationLog(uid, requestIp, "login")

	return c.JSON(200, NewUserData{
		UserId:  uid,
		Name:    userData.Name,
		Profile: userData.Profile,
		IconUrl: twitterUser.Data.ProfileImageUrl,
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
	user := db.User{}
	var cnt int64

	uid := c.Param("id")
	db.Db.Where("user_id = ?", uid).Find(&user).Count(&cnt)

	if cnt != 1 {
		return c.JSON(404, MakeError("gusr-001", "ユーザーが存在しません"))
	}

	res := UserData{
		IsSelf:      false,
		IconUrl:     user.IconUrl,
		TwitterName: "",
		Name:        user.Name,
		Profile:     user.Profile,
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
