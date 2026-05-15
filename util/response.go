package util

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/namejlt/gozen/errors"
	"github.com/namejlt/gozen/log"
)

var (
	apiOutFormat string
)

const (
	apiOutFormatDefault = "default"
	apiOutFormatCommon  = "common"
)

// SetResponseFormat sets the API output format, called by root init
func SetResponseFormat(format string) {
	apiOutFormat = format
}

func isApiOutFormatDefault() bool {
	return apiOutFormat == apiOutFormatDefault
}

func ResponseReturnJsonNoP(c *gin.Context, code int, model any, msg ...string) {
	ResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, false, true, http.StatusOK)
}

func ResponseReturnJson(c *gin.Context, code int, model any, msg ...string) {
	ResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, true, true, http.StatusOK)
}

func ResponseReturnJsonStatus(c *gin.Context, code int, model any, msg ...string) {
	var status int
	if code == 0 || code == 1001 {
		status = http.StatusOK
	} else if code == 1004 {
		status = http.StatusBadRequest
	} else {
		status = http.StatusInternalServerError
	}
	ResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, true, true, status)
}

func ResponseReturnJson400(c *gin.Context, code int, model any, msg ...string) {
	ResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, true, true, http.StatusBadRequest)
}

func ResponseReturnJson500(c *gin.Context, code int, model any, msg ...string) {
	ResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, true, true, http.StatusInternalServerError)
}

func ResponseReturnJsonNoPReal(c *gin.Context, code int, model any, msg ...string) {
	ResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, false, false, http.StatusOK)
}

func ResponseReturnJsonReal(c *gin.Context, code int, model any, msg ...string) {
	ResponseReturnJsonWithMsg(c, code, getResponseMsg(msg...), model, true, false, http.StatusOK)
}

func getResponseMsg(msg ...string) (message string) {
	if len(msg) > 0 && msg[0] != "" {
		message = msg[0]
	}
	return
}

func ResponseReturnJsonWithMsg(c *gin.Context, code int, msg string, model any,
	callbackFlag bool, unifyCode bool, status int) {
	if unifyCode && code == 0 && isApiOutFormatDefault() {
		code = 1001
	}
	if msg == "" {
		msg = errors.GetMessage(code)
	}

	var rj any
	if _, ok := errors.CodeFalseSet[code]; !ok {
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
			"traceId": "",
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

	if IsEmpty(callback) {
		c.JSON(status, rj)
	} else {
		r, err := json.Marshal(rj)
		if err != nil {
			log.Infow(log.NameLogic, "ResponseReturnJsonWithMsg json Marshal error",
				log.KNameCommonData, rj,
				log.KNameCommonErr, err,
			)
		} else {
			c.String(status, "%s(%s)", callback, r)
		}
	}
}

func ResponseReturnJsonFailed(c *gin.Context, code int) {
	ResponseReturnJson500(c, code, nil)
}

func ResponseReturnJsonSuccess(c *gin.Context, data any) {
	ResponseReturnJson(c, 0, data)
}

func ResponseOK(c *gin.Context, data any) {
	ResponseReturnJson(c, 0, data)
}

func ResponseFail(c *gin.Context, code int, msg ...string) {
	ResponseReturnJson(c, code, nil, msg...)
}

func ResponsePageOK(c *gin.Context, req PageRequest, total int64, rows any) {
	ResponseReturnJson(c, 0, NewPageResult(req, total, rows))
}

func ResponseInvalidParam(c *gin.Context, msg string) {
	ResponseReturnJson(c, errors.CodeParamsIncomplete, nil, msg)
}

func ResponseRedirect(c *gin.Context, url string) {
	c.Redirect(http.StatusMovedPermanently, url)
}

// NewSuccess creates a success response map with code 0.
func NewSuccess(data any) gin.H {
	return gin.H{"code": 0, "message": "success", "data": data}
}

// NewError creates an error response map with the given code and message.
func NewError(code int, msg string) gin.H {
	if msg == "" {
		msg = errors.GetMessage(code)
	}
	return gin.H{"code": code, "message": msg}
}

func responseJSONMarshal(t any) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}
