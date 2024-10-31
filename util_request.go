package gozen

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"net/url"
)

func UtilRequestGetParam(c *gin.Context, key string) string {
	if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
		return c.Query(key)
	}
	return c.PostForm(key)
}

func UtilRequestGetAllParams(c *gin.Context) (ret url.Values) {
	switch c.Request.Method {
	case "GET":
		fallthrough
	case "DELETE":
		ret = c.Request.URL.Query()
	case "POST":
		fallthrough
	case "PATCH":
		fallthrough
	case "PUT":
		c.Request.ParseForm()
		ret = c.Request.PostForm
	}
	return ret
}

func UtilRequestQueryDataString(c *gin.Context) string {
	var query url.Values
	query = UtilRequestGetAllParams(c)

	return query.Encode()
}

func BindParams(c *gin.Context, params interface{}) (code int, err error) {
	err = c.ShouldBind(params)
	if err != nil {
		LogInfow(LogNameLogic, "bind param error",
			LogKNameCommonReq, c.Request.URL,
			LogKNameCommonData, params,
			LogKNameCommonErr, err,
		)
		code = PConstCodeParamsIncomplete
		return
	}
	return
}

func BindParamsWithBody(c *gin.Context, params interface{}) (code int, err error) {
	var contentType string
	if c.Request.Method != http.MethodGet {
		contentType = c.ContentType()
	}
	var bb binding.BindingBody
	switch contentType {
	case binding.MIMEJSON:
		bb = binding.JSON
	case binding.MIMEXML, binding.MIMEXML2:
		bb = binding.XML
	case binding.MIMEPROTOBUF:
		bb = binding.ProtoBuf
	case binding.MIMEMSGPACK, binding.MIMEMSGPACK2:
		bb = binding.MsgPack
	case binding.MIMEYAML:
		bb = binding.YAML
	default:
		err = c.ShouldBind(params)
	}
	if bb != nil {
		err = c.ShouldBindBodyWith(params, bb)
	}
	if err != nil {
		LogInfow(LogNameLogic, "bind param error",
			LogKNameCommonReq, c.Request.URL,
			LogKNameCommonData, params,
			LogKNameCommonErr, err,
		)
		code = PConstCodeParamsIncomplete
		return
	}
	return
}
