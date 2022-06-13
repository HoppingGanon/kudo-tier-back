package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
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
