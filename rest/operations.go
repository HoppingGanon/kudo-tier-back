package rest

import (
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"

	common "reviewmakerback/common"
	db "reviewmakerback/db"
)

// セッション取得に失敗した際のリトライ回数
const SessionRetryCount = 4

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

	// ランダムな文字列を生成する
	codeVerifer := common.MakeRandomChars(codeVeriferCnt)

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

	// セッションIDの作成
	sessionId, err := makeSession(SessionRetryCount)
	// Twitterからユーザー情報の取得
	b, err := getTwitterApi("https://api.twitter.com/2/users/me", twitterToken.AccessToken)
	if err != nil {
		return c.String(http.StatusForbidden, "Twitterからアクセストークンが取得できませんでした")
	}
	var twitterUser TwitterUser
	err = json.Unmarshal(b, &twitterUser)
	if err != nil {
		return c.String(http.StatusForbidden, "Twitterから取得したアクセストークンが不正です")
	}

	// Userデータの中に該当するTwitterIdがあるかチェック
	var user db.User
	tid := twitterUser.Data.Id
	db.Db.Where("twitter_id = ?", tid).Find(&user).Count(&cnt)
	if cnt == 0 {
		// アクセスログを登録
		db.WriteAccessLog("twitter:"+tid, requestIp, accesstime, "login")

		// セッション登録
		session := db.Session{
			SessionID:    sessionId,
			UserId:       "",
			ExpiredTime:  expiredTime,
			TwitterToken: tid,
			IsNew:        true,
		}
		db.Db.Create(session)

		// レスポンスの内容を作成
		res := Session{
			SessionId:   sessionId,
			ExpiredTime: expiredTime.Format("02-Jan-2006 15:04:05-07"),
			IsNew:       true,
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
			TwitterToken: tid,
			IsNew:        false,
		}

		// レスポンスの内容を作成
		res := Session{
			SessionId:   sessionId,
			ExpiredTime: expiredTime.Format("02-Jan-2006 15:04:05-07"),
			IsNew:       false,
		}
		db.Db.Create(session)

		return c.JSON(200, res)
	}
}

func checkSession(c echo.Context) bool {
	sessionId := c.Request().Header.Get("sessionId")
	var session db.Session
	var cnt int64
	db.Db.Where("session_id = ?", sessionId).Find(&session).Count(&cnt)
	return cnt == 1
}

func getReqCheckSession(c echo.Context) error {
	if checkSession(c) {
		return c.String(200, "session exists")
	} else {
		return c.String(404, "session not exists")
	}
}

func delReqSession(c echo.Context) error {
	if checkSession(c) {
		sessionId := c.Request().Header.Get("sessionId")
		var session db.Session
		println(sessionId)
		db.Db.Where("session_id = ?", sessionId).Delete(&session)
		return c.String(200, "session deleted")
	} else {
		return c.String(404, "session not exists")
	}
}

func makeSession(retryCount int) (string, error) {
	var session db.Session
	var cnt int64
loop:
	for i := 0; i < retryCount; i++ {
		sessionId, err := common.MakeSession()
		if err != nil {
			continue loop
		}
		db.Db.Find(&session).Where("session_id = ?", sessionId).Count(&cnt)
		if cnt == 0 {
			return sessionId, nil
		}
	}
	return "", errors.New("セッション作成に失敗")
}
