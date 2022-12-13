package rest

import (
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"

	common "reviewmakerback/common"
	db "reviewmakerback/db"
	"unicode/utf8"
)

// 1つの発信元IPあたりの最大保持一時セッション数
const maxSessionPerIp = 16

// codeVeriferの文字数
const codeVeriferCnt = 64

func getReqHello(c echo.Context) error {
	return c.String(http.StatusOK, "{\"Hello\": \"World\"}")
}

/*
一時トークンを生成するGETリクエストの処理
*/
func getReqTempSession(c echo.Context) error {
	// 1兆通りのランダムな数字を生成する
	max, _ := new(big.Int).SetString("1000000000000", 10)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}

	// 1兆通りのランダムな数字と生成時間を文字列結合して、SHA256でハッシュ文字列をsession_idとする
	sessionId := b64.RawURLEncoding.EncodeToString([]byte(common.GetSHA256(time.Now().Format("2006-01-02-15-04-05") + ":" + n.Text(10))))
	var count int64
	var ipcount int64
	var tempsession db.TempSession

	// IPアドレスの特定
	requestIp := net.ParseIP(c.RealIP()).String()

	// セッションIDの重複チェック
	db.Db.Where("session_id = ?", sessionId).Find(&tempsession).Count(&count)
	if &count == nil || count > 0 {
		return c.JSON(http.StatusBadRequest, "一時接続用セッションの確立に失敗しました。しばらく時間を空けて再度実行してください。")
	}

	// 同一IPからの一時セッション上限チェック
	db.Db.Where("ip_address = ?", requestIp).Find(&tempsession).Count(&ipcount)
	if ipcount > maxSessionPerIp {
		return c.JSON(http.StatusBadRequest, "一時接続用セッションの確立に失敗しました。しばらく時間を空けて再度実行してください。")
	}

	// ランダムな文字列を生成する(IPアドレスを乱数のシードに含める)
	codeVerifer, err := common.MakeRandomChars(codeVeriferCnt, requestIp)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "一時接続用セッションの確立に失敗しました。しばらく時間を空けて再度実行してください。")
	}

	// 一時セッションのデータ作成
	tempsession = db.TempSession{
		SessionID:    sessionId,
		AccessTime:   time.Now(),
		IpAddress:    requestIp,
		CodeVerifier: codeVerifer,
	}

	// データベースに一時セッションを登録
	db.Db.Create(tempsession)

	// CodeVerifierをsha256でハッシュ化したのち、Base64変換
	codeChallenge := b64.RawURLEncoding.EncodeToString(common.GetBinSHA256(codeVerifer))

	// 一時セッションIDとcodeChallengeをクライアントに送付
	body := TempSession{
		SessionId:     sessionId,
		CodeChallenge: codeChallenge,
	}

	if err := c.Bind(&body); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, body)
}

/*
セッションを生成するGETリクエストの処理
*/
func getReqSession(c echo.Context) error {
	// クライアントから送付されたcodeと一時セッションを取り出す
	code := c.Request().Header.Get("code")
	tempSessionId := c.Request().Header.Get("tempSessionId")

	var tempsession db.TempSession
	var cnt int64

	// 一時セッションがデータベースに存在するか確認する
	db.Db.Where("session_id = ?", tempSessionId).Find(&tempsession).Count(&cnt)

	if cnt != 1 {
		return c.String(http.StatusForbidden, "一時セッションが不正です")
	}

	db.Db.Where("session_id = ?", tempSessionId).Find(&tempsession).First(&tempsession)

	if len(code) < 10 {
		return c.String(http.StatusForbidden, "クライアントが送付したコードが不正です")
	}

	// アクセス元IPと時刻を記録
	requestIp := net.ParseIP(c.RealIP()).String()
	accesstime := time.Now()

	// Twitterアクセストークンの取得
	twitterToken, err := postTwitterToken(code, os.Getenv("TW_REDIRECT_URI"), tempsession.CodeVerifier, os.Getenv("TW_CLIENT_ID"), os.Getenv("TW_CLIENT_SEC"))
	if err != nil {
		return c.String(http.StatusForbidden, "OAuth 2.0 認証に失敗しました")
	}

	if twitterToken.AccessToken == "" {
		return c.String(http.StatusForbidden, "Twitterからアクセストークンを取得できませんでした")
	}
	expiredTime := accesstime.Add(time.Duration(twitterToken.ExpiresIn) * time.Second)

	// 一時セッションの削除
	db.Db.Where("session_id = ?", tempSessionId).Delete(&tempsession)

	// セッションIDの作成(IPアドレスを乱数のシードに含める)
	sessionId, err := db.MakeSession(requestIp)

	if err != nil {
		return c.String(http.StatusForbidden, "エラーが発生しました")
	}

	// Twitterからユーザー情報の取得
	b, err := getTwitterApi("https://api.twitter.com/2/users/me?user.fields=profile_image_url", twitterToken.AccessToken)
	if err != nil {
		return c.String(http.StatusForbidden, "Twitterからユーザー情報が取得できませんでした")
	}
	var twitterUser TwitterUser
	err = json.Unmarshal(b, &twitterUser)
	if err != nil {
		return c.String(http.StatusForbidden, "Twitterから取得したユーザー情報が不正です")
	}

	// Userデータの中に該当するTwitterIdがあるかチェック
	var user db.User
	tid := twitterUser.Data.Id
	db.Db.Where("twitter_name = ?", tid).Find(&user).Count(&cnt)
	if cnt == 0 {
		// アクセスログを登録
		db.WriteAccessLog("twitter:"+tid, requestIp, accesstime, "login")

		// セッション登録
		session := db.Session{
			SessionID:    sessionId,
			UserId:       "",
			ExpiredTime:  expiredTime,
			TwitterToken: twitterToken.AccessToken,
			IsNew:        true,
		}
		db.Db.Create(session)

		// レスポンスの内容を作成
		res := Session{
			SessionId:       sessionId,
			UserId:          "",
			ExpiredTime:     expiredTime.Format("02-Jan-2006 15:04:05-07"),
			IsNew:           true,
			TwitterName:     twitterUser.Data.Name,
			TwitterUserName: twitterUser.Data.UserName,
			IconUrl:         twitterUser.Data.ProfileImageUrl,
		}

		return c.JSON(200, res)
	} else {
		// アクセスログを登録
		db.WriteAccessLog(user.UserId, requestIp, accesstime, "login")

		// セッション登録
		session := db.Session{
			SessionID:    sessionId,
			UserId:       user.UserId,
			ExpiredTime:  expiredTime,
			TwitterToken: twitterToken.AccessToken,
			IsNew:        false,
		}

		// レスポンスの内容を作成
		res := Session{
			SessionId:       sessionId,
			UserId:          user.UserId,
			ExpiredTime:     expiredTime.Format("02-Jan-2006 15:04:05-07"),
			IsNew:           false,
			TwitterName:     twitterUser.Data.Name,
			TwitterUserName: twitterUser.Data.UserName,
			IconUrl:         twitterUser.Data.ProfileImageUrl,
		}
		db.Db.Create(session)

		return c.JSON(200, res)
	}
}

func postReqUser(c echo.Context) error {
	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.String(403, "セッションがありません")
	}

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
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

	return c.JSON(200, NewUserData{
		UserId:  uid,
		Name:    userData.Name,
		Profile: userData.Profile,
		IconUrl: twitterUser.Data.ProfileImageUrl,
	})
}

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

func getReqCheckSession(c echo.Context) error {
	_, err := db.CheckSession(c)
	if err == nil {
		return c.String(200, "session exists")
	} else {
		return c.String(404, "session not exists")
	}
}

func delReqSession(c echo.Context) error {
	_, err := db.CheckSession(c)
	if err == nil {
		sessionId := c.Request().Header.Get("sessionId")
		var session db.Session
		println(sessionId)
		db.Db.Where("session_id = ?", sessionId).Delete(&session)
		return c.String(200, "session deleted")
	} else {
		return c.String(205, "session not exists")
	}
}

func getReqUserData(c echo.Context) error {
	// 送信元ユーザーと参照先ユーザーが同じかどうかチェック
	session, err := db.CheckSession(c)
	user := db.User{}
	var cnt int64

	uid := c.Param("id")
	println(uid)
	db.Db.Where("user_id = ?", uid).Find(&user).Count(&cnt)

	if cnt != 1 {
		return c.JSON(404, "ユーザーが存在しません")
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

func ConvertParags(parags []ParagData) ([]db.Parag, error) {
	size := len(parags)
	var conved = make([]db.Parag, size)
	for i := 0; i < size; i++ {
		if !db.IsParagraphType(parags[i].Type) {
			return []db.Parag{}, errors.New("Paragraphについて、Typeの値が異常です")
		}
		conved[i] = db.Parag{
			Type: parags[i].Type,
			Body: parags[i].Body,
		}
	}
	return conved, nil
}

func ConvertParams(params []ReviewParamData) []db.ReviewParam {
	size := len(params)
	var conved = make([]db.ReviewParam, size)
	for i := 0; i < size; i++ {
		conved[i] = db.ReviewParam{
			Name:    params[i].Name,
			IsPoint: params[i].IsPoint,
			Weight:  params[i].Weight,
		}
	}
	return conved
}

func postReqTier(c echo.Context) error {
	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.String(403, "セッションがありません")
	}

	// Bodyの読み取り
	b, _ := ioutil.ReadAll(c.Request().Body)
	var tierData TierPostData
	err = json.Unmarshal(b, &tierData)
	if err != nil {
		return c.String(400, "不正なTierデータです1"+err.Error())
	}

	params := ConvertParams(tierData.ReviewFactorParams)

	parags, err := ConvertParags(tierData.Parags)
	if err != nil {
		return c.String(400, "不正なTierデータです2")
	}

	db.CreateTier(session.SessionID, tierData.Name, tierData.ImageBase64, parags, tierData.PointType, params)

	return c.String(201, "")
}
