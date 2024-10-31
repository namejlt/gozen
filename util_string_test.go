package gozen

import (
	"testing"
	"unicode/utf8"
)

func Test_StringEmoji(t *testing.T) {
	str := "%F0%9F%91%8D%F0%9F%91%8D%F0%9F%91%8D%F0%9F%91%8D"
	t.Logf(str)
}

func Test_UtilStringGenerateRandomString(t *testing.T) {
	str := UtilStringGenerateRandomString(16)
	if utf8.RuneCountInString(str) != 16 {
		t.Error("len error")
	}
	t.Log(str)
}

func Test_UtilStringGenerateRandomStringNoSymbol(t *testing.T) {
	str := UtilStringGenerateRandomStringNoSymbol(16)
	if utf8.RuneCountInString(str) != 16 {
		t.Error("len error")
	}
	t.Log(str)
}

func Test_UtilCryptoGenerateRandomToken16(t *testing.T) {
	str, _ := UtilCryptoGenerateRandomToken16()
	if utf8.RuneCountInString(str) != 16 {
		t.Error("len error")
	}
	t.Log(str)
}
