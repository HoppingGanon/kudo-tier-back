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
	db.Db.Find(&tempsession).Where("session_id = ?", sessionId).Count(&count)
	if &count == nil || count > 0 {
		return c.JSON(http.StatusBadRequest, "一時接続用セッションの確立に失敗しました。しばらく時間を空けて再度実行してください。")
	}

	// 同一IPからの一時セッション上限チェック
	db.Db.Find(&tempsession).Where("ip_address = ?", requestIp).Count(&ipcount)
	if ipcount > maxSessionPerIp {
		return c.JSON(http.StatusBadRequest, "一時接続用セッションの確立に失敗しました。しばらく時間を空けて再度実行してください。")
	}

	// ランダムな文字列を生成する
	codeVerifer := common.MakeRandomChars(codeVeriferCnt)

	println("ver: " + codeVerifer)

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
アクセストークン(JWT)を生成するGETリクエストの処理
*/
func getReqToken(c echo.Context) error {
	// クライアントから送付されたcodeと一時セッションを取り出す
	code := c.QueryParam("code")
	tempSessionId := c.QueryParam("temp_session")

	println("---")
	println(tempSessionId)
	println("---")

	var tempsession db.TempSession
	var cnt int64

	// 一時セッションがデータベースに存在するか確認する
	records := db.Db.Find(&tempsession).Where("session_id = ?", tempSessionId)
	records.Count(&cnt)

	if cnt != 1 {
		return c.String(http.StatusForbidden, "一時セッションが不正です")
	}

	records.First(&tempsession)

	if len(code) < 10 {
		return c.String(http.StatusForbidden, "クライアントが送付したコードが不正です")
	}

	// アクセス元IPと時刻を記録
	requestIp := net.ParseIP(c.RealIP()).String()
	accesstime := time.Now()

	// Twitterアクセストークンの取得
	println("ver: " + tempsession.CodeVerifier)
	twitterToken, err := postTwitterToken(code, os.Getenv("TW_REDIRECT_URI"), tempsession.CodeVerifier, os.Getenv("TW_CLIENT_ID"), os.Getenv("TW_CLIENT_SEC"))
	if err != nil {
		return c.String(http.StatusForbidden, "OAuth 2.0 認証に失敗しました")
	}

	if twitterToken.AccessToken == "" {
		return c.String(http.StatusForbidden, "Twitterからアクセストークンを取得できませんでした")
	}

	println(twitterToken.AccessToken)
	println("---")

	// セッションIDの作成
	sessionId, err := makeSession(SessionRetryCount)
	println(sessionId)
	println("---")
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
	print(string(b))

	// アクセスログを登録
	db.WriteAccessLog("aaaa", requestIp, accesstime, "login")

	// レスポンスの内容を作成
	session := Session{
		ExpiredTime: accesstime.Format("2006-01-02-15-04-05"),
		SessionId:   sessionId + " . " + twitterUser.Data.Id,
	}

	return c.JSON(200, session)
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
