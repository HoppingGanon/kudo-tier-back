package tests

import (
	"fmt"
	"reviewmakerback/db"
	"testing"
)

func TestEncrypting(t *testing.T) {
	const plaintext = "plaintext"
	const password = "passwordpassword"
	fmt.Println("================ EN ================")
	fmt.Printf("plaintext = '%s'\n", plaintext)
	fmt.Printf("password = '%s'\n", password)
	etd, err := db.EncryptText(plaintext, password)
	fmt.Printf("etd.Base64Text = '%s'\n", etd.Base64Text)
	fmt.Printf("etd.Length = %d\n", etd.Length)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println("================ DE ================")
	result, err := db.DecryptText(etd, password)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Printf("result = '%s'\n", result)

	if result != plaintext {
		t.Error("miss")
	}
}
func TestEncryptingNone(t *testing.T) {
	const plaintext = ""
	const password = "passwordpassword"
	fmt.Println("================ EN ================")
	fmt.Printf("plaintext = '%s'\n", plaintext)
	fmt.Printf("password = '%s'\n", password)
	etd, err := db.EncryptText(plaintext, password)
	fmt.Printf("etd.Base64Text = '%s'\n", etd.Base64Text)
	fmt.Printf("etd.Length = %d\n", etd.Length)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println("================ DE ================")
	result, err := db.DecryptText(etd, password)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Printf("result = '%s'\n", result)

	if result != plaintext {
		t.Error("miss")
	}
}
