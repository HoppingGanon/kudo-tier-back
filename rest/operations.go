package rest

import (
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo"

	common "reviewmakerback/common"
	db "reviewmakerback/db"
)

// 1つの発信元IPあたりの最大保持セッション数
const maxSessionPerIp = 16

// codeVeriferの文字数
const codeVeriferCnt = 64

func getReqHello(c echo.Context) error {
	return c.String(http.StatusOK, "{\"Hello\": \"World\"}")
}

func getReqTempSession(c echo.Context) error {
	max, _ := new(big.Int).SetString("1000000000000", 10)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}

	session := b64.RawURLEncoding.EncodeToString([]byte(common.GetSHA256(time.Now().Format("2006-01-02-15-04-05") + ":" + n.Text(10))))
	var count int64
	var ipcount int64
	var tempsession db.TempSession

	// IPアドレスの特定
	requestIp := net.ParseIP(c.RealIP()).String()

	db.Db.Find(&tempsession).Where("session_id = ?", session).Count(&count)
	db.Db.Find(&tempsession).Where("ip_address = ?", requestIp).Count(&ipcount)

	if &count == nil || count > 0 || ipcount > maxSessionPerIp {
		return c.JSON(http.StatusBadRequest, "一時接続用セッションの確立に失敗しました。しばらく時間を空けて再度実行してください。")
	}

	// ランダムな文字列を生成する
	codeVerifer := common.MakeRandomChars(codeVeriferCnt)

	tempsession = db.TempSession{
		SessionID:    session,
		AccessTime:   time.Now(),
		IpAddress:    requestIp,
		CodeVerifier: codeVerifer,
	}

	// データベースに登録
	db.Db.Create(tempsession)

	// CodeVerifierをsha256でハッシュ化したのち、Base64変換
	codeChallenge := b64.RawURLEncoding.EncodeToString([]byte(common.GetBinSHA256(codeVerifer)))

	// セッションIDとcodeChallengeを送付
	body := TempSession{
		SessionId:     session,
		CodeChallenge: codeChallenge,
	}

	if err := c.Bind(&body); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, body)
}

func getReqSession(c echo.Context) error {
	code := c.QueryParam("code")
	tempSessionId := c.QueryParam("temp_session")
	var tempsession db.TempSession
	var cnt int64
	println(tempSessionId)
	records := db.Db.Find(&tempsession).Where("session_id = ?", tempSessionId)
	records.Count(&cnt)

	if cnt != 1 {
		return c.String(http.StatusForbidden, "一時セッションが不正です")
	}

	records.First(&tempsession)

	if len(code) < 10 {
		return c.String(http.StatusForbidden, "クライアントが送付したコードが不正です")
	}

	endpoint := "https://api.twitter.com/2/oauth2/token"

	values := url.Values{}
	values.Set("code", code)
	values.Add("grant_type", "authorization_code")
	values.Add("redirect_uri", os.Getenv("TW_REDIRECT_URI"))
	values.Add("code_verifier", tempsession.CodeVerifier)
	values.Add("client_id", os.Getenv("TW_CLIENT_ID"))

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return c.String(http.StatusForbidden, "OAuth 2.0 認証に失敗しました")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(os.Getenv("TW_CLIENT_ID"), os.Getenv("TW_CLIENT_SEC"))

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusForbidden, "OAuth 2.0 認証に失敗しました")
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.String(http.StatusForbidden, "OAuth 2.0 認証に失敗しました")
	}
	println(string(b))

	twitterToken := TwitterToken{}
	err = json.Unmarshal(b, &twitterToken)

	return c.String(200, string(b))
}
