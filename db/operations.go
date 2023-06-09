package db

import (
	"crypto/aes"
	"crypto/cipher"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	common "reviewmakerback/common"

	"github.com/labstack/echo"
	"gorm.io/gorm"
)

// 一時セッションが生存する時間およびセッションの生死整理を行う間隔(秒)
const TempSessionAlive = 60

// セッションを削除する間隔(秒)
const SessionDelSpan = 60

// 各ID作成に失敗した際の最大試行回数
const RetryCreateCnt = 3

// 各IDの桁数
const idSize = 16

// 最小投稿間隔にの初期値(mainから上書きする)
var PostSpanMin = 10

func WriteOperationLog(id string, ipAddress string, operation string, content string) {
	// ログを記録
	log := OperationLog{
		UserId:    id,
		IpAddress: ipAddress,
		Operation: operation,
		Content:   content,
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
	fmt.Printf("id=%s\nip=%s\nerrId=%s\nope=%s\ndes=%s\n", id, ipAddress, errorId, operation, descriptions)
	// データベースに登録
	Db.Create(log)
}

func CheckSession(c echo.Context, requireUser bool, updateExpiredTime bool) (Session, error) {
	token := c.Request().Header.Get("Authorization")
	typeStr := common.Substring(token, 0, 7)

	if typeStr != "Bearer " {
		return Session{}, errors.New("認証タイプが異常です")
	}
	sessionId := common.Substring(token, 7, len(token)-7)

	var session Session
	var cnt int64
	tx := Db.Where("session_id = ?", sessionId)

	tx.Find(&session).Count(&cnt)
	if cnt != 1 {
		return Session{}, errors.New("セッションがありません")
	}

	var user User
	if requireUser {
		Db.Where("user_id = ?", session.UserId).Find(&user).Count(&cnt)
		if cnt != 1 {
			return Session{}, errors.New("ユーザーが存在しません")
		}
	}

	if updateExpiredTime {
		exTime := time.Now().Add(time.Duration(user.KeepSession) * time.Second)
		tx.Update("expired_time", exTime)
		session.ExpiredTime = exTime
	}

	return session, nil
}

// 最小投稿時間をあけているかチェック
func CheckLastPost(session Session) bool {
	return session.LastPostAt.Add(time.Second * time.Duration(PostSpanMin)).After(time.Now())
}

// 投稿時間を記録
func UpdateLastPostAt(session Session) {
	Db.Model(&session).Update("last_post_at", time.Now())
}

func WordToReg(word string) string {
	word = strings.ReplaceAll(word, "\\", "")
	word = strings.ReplaceAll(word, "'", "")
	word = strings.ReplaceAll(word, "\"", "")
	word = strings.ReplaceAll(word, "[", "")
	word = strings.ReplaceAll(word, "]", "")
	word = strings.ReplaceAll(word, "(", "")
	word = strings.ReplaceAll(word, ")", "")
	word = strings.ReplaceAll(word, "{", "")
	word = strings.ReplaceAll(word, "}", "")
	word = strings.ReplaceAll(word, "!", "")
	word = strings.ReplaceAll(word, "?", "")
	word = strings.ReplaceAll(word, "*", "")
	word = strings.ReplaceAll(word, ".", "")
	word = strings.ReplaceAll(word, "^", "")
	word = strings.ReplaceAll(word, "$", "")
	word = strings.ReplaceAll(word, "/", "")

	word = strings.ReplaceAll(word, " ", ")|(")

	return ".*[(" + word + ")].*"
}

func SearchWord(columns []string, word string) *gorm.DB {
	word = strings.ReplaceAll(word, "\\", "")
	word = strings.ReplaceAll(word, "'", "")
	word = strings.ReplaceAll(word, "\"", "")
	word = strings.ReplaceAll(word, "[", "")
	word = strings.ReplaceAll(word, "]", "")
	word = strings.ReplaceAll(word, "(", "")
	word = strings.ReplaceAll(word, ")", "")
	word = strings.ReplaceAll(word, "{", "")
	word = strings.ReplaceAll(word, "}", "")
	word = strings.ReplaceAll(word, "!", "")
	word = strings.ReplaceAll(word, "?", "")
	word = strings.ReplaceAll(word, "*", "")
	word = strings.ReplaceAll(word, "/", "")
	word = strings.ReplaceAll(word, "%", "")

	var txAnd *gorm.DB
	var txOr *gorm.DB
	txAnd = Db
	for _, like := range strings.Split(word, " ") {
		txOr = Db
		for _, column := range columns {
			txOr = txOr.Or(column+" like ?", "%"+like+"%")
		}
		txAnd = txAnd.Where(txOr)
	}
	return txAnd
}

func ArrangeSession() {
	// 一時セッションの生存期間が終了したデータを削除
	Db.Where("access_time < ?", time.Now().Add(-TempSessionAlive*time.Second)).Delete(&TempSession{})
	// セッションの生存期間が終了したデータを削除
	Db.Where("expired_time < ?", time.Now()).Delete(&Session{})
}

// 指定した項目を除外したSelect句を作成する
// ただし、項目名はキャメルケースで指定すること
func ExcludeSelect(baseStruct interface{}, columns ...string) string {
	types := reflect.TypeOf(baseStruct)
	var name string
	selectText := ""
	for i := 0; i < reflect.ValueOf(baseStruct).NumField(); i++ {
		name = types.Field(i).Name
		if !common.Contains(name, columns) {
			selectText += common.ToSnakeCase(name)
			if i != reflect.ValueOf(baseStruct).NumField()-1 {
				selectText += ", "
			}
		}
	}
	return selectText
}

// ==========================================================================================
// 9.6 データを暗号化/復号する
// https://astaxie.gitbooks.io/build-web-application-with-golang/content/ja/09.6.html

// AESの暗号化・復号化に必要な鍵
var commonIV = []byte{
	0xed,
	0xb1,
	0xfa,
	0x64,
	0x72,
	0xa4,
	0x61,
	0xe0,
	0x8c,
	0x9d,
	0x6c,
	0x82,
	0x01,
	0xa0,
	0xcc,
	0x50,
	/*
		0x49,
		0x4b,
		0xb9,
		0x9a,
		0x60,
		0x5c,
		0xa6,
		0x3e,
		0x4a,
		0x5d,
		0xcf,
		0x1c,
		0xd7,
		0x91,
		0xb9,
		0x5a,
	*/
}

type EncryptedTextData struct {
	Base64Text string `json:"b"`
	Length     int    `json:"l"`
}

func EncryptText(text string, password string) (EncryptedTextData, error) {
	plaintext := []byte(text)

	// 暗号化アルゴリズムaesを作成
	c, err := aes.NewCipher([]byte(password))
	if err != nil {
		return EncryptedTextData{}, err
	}

	// 暗号化文字列
	cfb := cipher.NewCFBEncrypter(c, commonIV)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)
	b64Text := b64.StdEncoding.EncodeToString(ciphertext)

	return EncryptedTextData{
		Base64Text: b64Text,
		Length:     len(plaintext),
	}, nil
}

func DecryptText(etd EncryptedTextData, password string) (string, error) {
	// 暗号化アルゴリズムaesを作成
	c, err := aes.NewCipher([]byte(password))
	if err != nil {
		return "", err
	}

	// 暗号化文字列
	cfbdec := cipher.NewCFBDecrypter(c, commonIV)

	plaintextCopy := make([]byte, etd.Length)
	ciphertext, err := b64.StdEncoding.DecodeString(etd.Base64Text)
	if err != nil {
		return "", err
	}
	cfbdec.XORKeyStream(plaintextCopy, ciphertext)

	return string(plaintextCopy), nil
}

func EncryptTextJson(text string, password string) (string, error) {
	plaintext := []byte(text)

	// 暗号化アルゴリズムaesを作成
	c, err := aes.NewCipher([]byte(password))
	if err != nil {
		return "{}", err
	}

	// 暗号化文字列
	cfb := cipher.NewCFBEncrypter(c, commonIV)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)
	b64Text := b64.StdEncoding.EncodeToString(ciphertext)

	bytes, err := json.Marshal(EncryptedTextData{
		Base64Text: b64Text,
		Length:     len(plaintext),
	})

	return string(bytes), err
}

func DecryptTextJson(jsonText string, password string) (string, error) {
	var etd EncryptedTextData
	err := json.Unmarshal([]byte(jsonText), &etd)
	if err != nil {
		return "", err
	}

	// 暗号化アルゴリズムaesを作成
	c, err := aes.NewCipher([]byte(password))
	if err != nil {
		return "", err
	}

	// 暗号化文字列
	cfbdec := cipher.NewCFBDecrypter(c, commonIV)

	plaintextCopy := make([]byte, etd.Length)
	ciphertext, err := b64.StdEncoding.DecodeString(etd.Base64Text)
	if err != nil {
		return "", err
	}
	cfbdec.XORKeyStream(plaintextCopy, ciphertext)

	return string(plaintextCopy), nil
}

// ==========================================================================================
