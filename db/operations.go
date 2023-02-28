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

// 一時セッションが生存する時間(秒)
const TempSessionAlive = 60

// 一時セッションを削除する間隔(秒)
const TempSessionDelSpan = 60

// 投稿可能な最小間隔(秒)
const PostSpanMin = 10

// 各ID作成に失敗した際の最大試行回数
const retryCreateCnt = 3

// 各IDの桁数
const idSize = 16

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

func CheckSession(c echo.Context) (Session, error) {
	token := c.Request().Header.Get("Authorization")
	typeStr := common.Substring(token, 0, 7)

	if typeStr != "Bearer " {
		return Session{}, errors.New("認証タイプが異常です")
	}
	sessionId := common.Substring(token, 7, len(token)-7)

	var session Session
	var cnt int64
	Db.Where("session_id = ?", sessionId).Find(&session).Count(&cnt)
	if cnt == 1 {
		return session, nil
	}
	return Session{}, errors.New("セッションがありません")
}

// 最小投稿時間をあけているかチェック
func CheckLastPost(session Session) bool {
	return session.LastPostAt.Add(time.Second * PostSpanMin).After(time.Now())
}

// 投稿時間を記録
func UpdateLastPostAt(session Session) {
	Db.Model(&session).Update("last_post_at", time.Now())
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
		Db.Where("session_id = ?", sessionId).Find(&session).Count(&cnt)
		if cnt == 0 {
			return sessionId, nil
		}
	}
	return "", errors.New("セッション作成に失敗")
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
	Db.Where("access_time < ?", time.Now().Add(-TempSessionDelSpan*time.Second)).Delete(&TempSession{})
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
