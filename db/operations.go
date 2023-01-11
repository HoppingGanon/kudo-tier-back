package db

import (
	"errors"
	"reflect"
	"strings"
	"time"

	common "reviewmakerback/common"

	"github.com/labstack/echo"
	"gorm.io/gorm"
)

const TempSessionAlive = 60
const TempSessionDelSpan = 60

// 各ID作成に失敗した際の最大試行回数
const retryCreateCnt = 3

// 各IDの桁数
const idSize = 16

func WriteOperationLog(id string, ipAddress string, operation string) {
	// ログを記録
	log := OperationLog{
		UserId:    id,
		IpAddress: ipAddress,
		Operation: operation,
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

	// データベースに登録
	Db.Create(log)
}

func CheckSession(c echo.Context) (Session, error) {
	sessionId := c.Request().Header.Get("sessionId")
	var session Session
	var cnt int64
	Db.Where("session_id = ?", sessionId).Find(&session).Count(&cnt)
	if cnt == 1 {
		return session, nil
	}
	return session, errors.New("セッションがありません")
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
