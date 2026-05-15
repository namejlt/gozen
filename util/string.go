package util

import (
	"bytes"
	"math/rand"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	Letters         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!#$%&*+=?@^_|-"
	LettersNoSymbol = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

func IsEmpty(data string) bool {
	return strings.Trim(data, " ") == ""
}

func GetStringFromIntArray(data []int, sep string) string {

	dataStr := GetStringArrayFromIntArray(data)

	return strings.Join(dataStr, sep)

}

func GetStringFromInt64Array(data []int64, sep string) string {

	dataStr := GetStringArrayFromInt64Array(data)

	return strings.Join(dataStr, sep)

}

func GetStringArrayFromIntArray(data []int) []string {

	model := []string{}

	for _, item := range data {

		m := strconv.Itoa(item)

		model = append(model, m)

	}
	return model
}
func GetStringArrayFromInt64Array(data []int64) []string {

	model := []string{}

	for _, item := range data {

		m := strconv.FormatInt(item, 10)

		model = append(model, m)

	}
	return model
}

func SplitToIntArray(data string, sep string) []int {
	var model []int

	dataArray := strings.Split(data, sep)

	for _, item := range dataArray {
		m, err := strconv.Atoi(item)

		if err != nil {
			continue
		}

		model = append(model, m)
	}
	return model
}

func SplitToInt64Array(data string, sep string) []int64 {
	var model []int64

	dataArray := strings.Split(data, sep)

	for _, item := range dataArray {
		m, err := strconv.ParseInt(item, 10, 64)

		if err != nil {
			continue
		}

		model = append(model, m)
	}
	return model
}

func StringGenerateRandomString(n int) string {
	letters := []rune(Letters)
	rand.Seed(time.Now().UTC().UnixNano())
	randomString := make([]rune, n)
	for i := range randomString {
		randomString[i] = letters[rand.Intn(len(letters))]
	}
	return string(randomString)
}

func StringGenerateRandomStringNoSymbol(n int) string {
	letters := []rune(LettersNoSymbol)
	rand.Seed(time.Now().UTC().UnixNano())
	randomString := make([]rune, n)
	for i := range randomString {
		randomString[i] = letters[rand.Intn(len(letters))]
	}
	return string(randomString)
}

func StringCheckStringExisted(strs []string, str string) bool {
	for _, v := range strs {
		if v == str {
			return true
		}
	}

	return false
}

func StringContains(obj any, target any) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	default:
		return false
	}
	return false
}

func StringConcat(buffer *bytes.Buffer, str string) {
	buffer.WriteString(str)
}

func StringConcatExist(strs []string, str string) []string {
	return append(strs, str)
}

func GetUrlHost(urlPath string) (host string) {
	u, err := url.Parse(urlPath)
	if err != nil {
		host = urlPath
	} else {
		host = u.Host
	}
	return
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// length：截取长度，负数表示截取到末尾
func SubStr(str string, start int, length int) (result string) {
	s := []rune(str)
	total := len(s)
	if total == 0 {
		return
	}
	// 允许从尾部开始计算
	if start < 0 {
		start = total + start
		if start < 0 {
			return
		}
	}
	if start > total {
		return
	}
	// 到末尾
	if length < 0 {
		length = total
	}

	end := start + length
	if end > total {
		result = string(s[start:])
	} else {
		result = string(s[start:end])
	}

	return
}

// JSONDecodeUseNumber 解析json，舍弃科学计数法
func JSONDecodeUseNumber(body []byte, obj any) (err error) {
	gin.EnableJsonDecoderUseNumber()
	err = binding.JSON.BindBody(body, obj)
	return
}
