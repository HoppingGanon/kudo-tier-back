package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math/big"
	"time"
)

func GetBinSHA256(s string) []byte {
	r := sha256.Sum256([]byte(s))
	return r[:]
}

func GetSHA256(s string) string {
	return hex.EncodeToString(GetBinSHA256(s))
}

func MakeRandomChars(codeVeriferCnt int) string {
	chars := ""
	for i := 0; i < codeVeriferCnt; i++ {
		chars += MakeRandomChar()
	}
	return chars
}

func MakeRandomChar() string {
	brandNum, err := rand.Int(rand.Reader, big.NewInt(62))
	randNum := brandNum.Int64()

	if err != nil {
		randNum = big.NewInt(1).Int64()
	}

	if randNum < 26 {
		return string(65 + randNum)
	} else if randNum < 52 {
		return string(97 + randNum - 26)
	} else {
		return string(48 + randNum - 52)
	}
}

func MakeSession() (string, error) {
	// 1兆通りのランダムな数字を生成する
	max, _ := new(big.Int).SetString("1000000000000", 10)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", errors.New("乱数を生成できません")
	}

	// 1兆通りのランダムな数字と生成時間を文字列結合して、SHA256でハッシュ文字列をsession_idとする
	sessionId := base64.RawURLEncoding.EncodeToString([]byte(GetSHA256(time.Now().Format("2006-01-02-15-04-05") + ":" + n.Text(10))))

	return sessionId, nil
}
