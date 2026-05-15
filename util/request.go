package util

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/namejlt/gozen/errors"
	"github.com/namejlt/gozen/log"
)

func RequestGetParam(c *gin.Context, key string) string {
	if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
		return c.Query(key)
	}
	return c.PostForm(key)
}

func RequestGetAllParams(c *gin.Context) (ret url.Values) {
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

func RequestQueryDataString(c *gin.Context) string {
	var query url.Values
	query = RequestGetAllParams(c)

	return query.Encode()
}

func BindParams(c *gin.Context, params any) (code int, err error) {
	err = c.ShouldBind(params)
	if err != nil {
		log.Infow(log.NameLogic, "bind param error",
			log.KNameCommonReq, c.Request.URL,
			log.KNameCommonData, params,
			log.KNameCommonErr, err,
		)
		code = errors.CodeParamsIncomplete
		return
	}
	return
}

func BindParamsWithBody(c *gin.Context, params any) (code int, err error) {
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
		log.Infow(log.NameLogic, "bind param error",
			log.KNameCommonReq, c.Request.URL,
			log.KNameCommonData, params,
			log.KNameCommonErr, err,
		)
		code = errors.CodeParamsIncomplete
		return
	}
	return
}
