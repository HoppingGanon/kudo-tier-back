package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// 文字列配列の中に指定した文字列が存在するかどうかチェックする関数
func Contains(s string, a []string) bool {
	for _, v := range a {
		if s == v {
			return true
		}
	}
	return false
}

// SHA256のハッシュをバイナリで返す
func GetBinSHA256(s string) []byte {
	r := sha256.Sum256([]byte(s))
	return r[:]
}

// SHA256の文字列(hex)をバイナリで返す
func GetSHA256(s string) string {
	return hex.EncodeToString(GetBinSHA256(s))
}

// 指定数のランダムな文字列(hex)を返す (最大128文字)
func MakeRandomChars(codeCount int, seed string) (string, error) {
	// SHA256で対応できるかチェック
	if codeCount < 0 || codeCount > 128 {
		return "", errors.New("指定文字数で出力できません")
	}

	// 1万通りのランダムな数字を生成する
	max, _ := new(big.Int).SetString("10000", 10)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", errors.New("乱数を生成できません")
	}

	// seedと1兆通りのランダムな数字と生成時間を文字列結合して、SHA256でハッシュ文字列を指定数切り取り、それを出力する
	chars := GetSHA256(seed + time.Now().Format("2006-01-02-15-04-05") + ":" + n.Text(10))
	return chars[0:codeCount], nil
}

func MakeSession(seed string) (string, error) {
	// 512文字のセッションIDを生成する
	sessionId, err := MakeRandomChars(64, seed)

	return sessionId, err
}

func DateToString(v time.Time) string {
	return v.Format("02-Jan-2006 15:04:05-07")
}

func TestRegexp(reg, str string) bool {
	return regexp.MustCompile(reg).Match([]byte(str))
}

// 参考: Goで文字列をスネークケースに変換する
// 著者: ohnishi
// https://zenn.dev/ohnishi/articles/1c84376fe89f70888b9c
func ToSnakeCase(s string) string {
	b := &strings.Builder{}
	for i, r := range s {
		if i == 0 {
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		if unicode.IsUpper(r) {
			b.WriteRune('_')
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
