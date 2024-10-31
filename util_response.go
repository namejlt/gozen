package gozen

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

/**

公司通用输出格式

{
        "traceId": "81ed62c8d99743349c61bf1693468705",
        "code": "0",
        "msg": "success",
        "success": true,
        "data": {
                "total": 1,
                "rows": [{}]
        }
}


*/

var (
	apiOutFormat string
)

const (
	apiOutFormatDefault = "default"
	apiOutFormatCommon  = "common"
)

func init() {
	apiOutFormat = ConfigAppGetString("ApiOutFormat", apiOutFormatDefault)
}

func isApiOutFormatDefault() bool {
	return apiOutFormat == apiOutFormatDefault
}

func UtilResponseReturnJsonNoP(c *gin.Context, code int, model interface{}, msg ...string) {
	UtilResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, false, true, http.StatusOK)
}

func UtilResponseReturnJson(c *gin.Context, code int, model interface{}, msg ...string) {
	UtilResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, true, true, http.StatusOK)
}

func UtilResponseReturnJsonStatus(c *gin.Context, code int, model interface{}, msg ...string) {
	var status int
	if code == 0 || code == 1001 {
		status = http.StatusOK
	} else if code == 1004 {
		status = http.StatusBadRequest
	} else {
		status = http.StatusInternalServerError
	}
	UtilResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, true, true, status)
}

func UtilResponseReturnJson400(c *gin.Context, code int, model interface{}, msg ...string) {
	UtilResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, true, true, http.StatusBadRequest)
}

func UtilResponseReturnJson500(c *gin.Context, code int, model interface{}, msg ...string) {
	UtilResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, true, true, http.StatusInternalServerError)
}

func UtilResponseReturnJsonNoPReal(c *gin.Context, code int, model interface{}, msg ...string) {
	UtilResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, false, false, http.StatusOK)
}

func UtilResponseReturnJsonReal(c *gin.Context, code int, model interface{}, msg ...string) {
	UtilResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, true, false, http.StatusOK)
}

func getResponseMsg(msg ...string) (message string) {
	if len(msg) > 0 && msg[0] != "" {
		message = msg[0]
	}
	return
}

func UtilResponseReturnJsonWithMsg(c *gin.Context, code int, msg string, model interface{},
	callbackFlag bool, unifyCode bool, status int) {
	if unifyCode && code == 0 && isApiOutFormatDefault() {
		code = 1001
	}
	if msg == "" {
		msg = ConfigCodeGetMessage(code)
	}

	var rj interface{}
	//添加结果
	if _, ok := codeFalseMap[code]; !ok {
		c.Set("result", true)
	} else {
		c.Set("result", false)
	}

	switch apiOutFormat {
	case apiOutFormatDefault:
		rj = gin.H{
			"code":    code,
			"message": msg,
			"data":    model,
		}
	case apiOutFormatCommon:
		var success bool
		if code == 0 {
			success = true
		}
		rj = gin.H{
			"traceId": "", // 待添加，从链路跟踪获取
			"code":    strconv.Itoa(code),
			"msg":     msg,
			"success": success,
			"data":    model,
		}
	default:
		rj = gin.H{
			"code":    code,
			"message": msg,
			"data":    model,
		}
	}

	var callback string
	if callbackFlag {
		callback = c.Query("callback")
	}

	if UtilIsEmpty(callback) {
		c.JSON(status, rj)
	} else {
		r, err := json.Marshal(rj)
		if err != nil {
			LogErrorw(LogNameLogic, "UtilResponseReturnJsonWithMsg json Marshal error",
				LogKNameCommonData, rj,
				LogKNameCommonErr, err,
			)
		} else {
			c.String(status, "%s(%s)", callback, r)
		}
	}
}

func UtilResponseReturnJsonFailed(c *gin.Context, code int) {
	UtilResponseReturnJson500(c, code, nil)
}

func UtilResponseReturnJsonSuccess(c *gin.Context, data interface{}) {
	UtilResponseReturnJson(c, 0, data)
}

func UtilResponseRedirect(c *gin.Context, url string) {
	c.Redirect(http.StatusMovedPermanently, url)
}

func utilResponseJSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}
