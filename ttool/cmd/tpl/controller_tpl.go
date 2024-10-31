package tpl

var (
	ControllerDirName   = "controller"
	ControllerFilesName = []string{
		"controller.go",
		"home.go",
	}
	ControllerFilesContent = []string{
		ControllerControllerGo,
		ControllerHomeGo,
	}

	ControllerV1DirName   = "controller/v1"
	ControllerV1FilesName = []string{
		"admin_tests.go",
		"api_tests.go",
		"app_tests.go",
		"iapi_tests.go",
		"iapi_tpl.go",
	}
	ControllerV1FilesContent = []string{
		ControllerV1AdminTestsGo,
		ControllerV1ApiTestsGo,
		ControllerV1AppTestsGo,
		ControllerV1IApiTestsGo,
		ControllerV1IApiTplGo,
	}
)

var (
	ControllerControllerGo = `package controller

import (
	"strconv"
	"strings"

	"github.com/namejlt/gozen"
	"github.com/gin-gonic/gin"
)

func ControllerGetUid(c *gin.Context) uint64 {
	uid := gozen.UtilRequestGetParam(c, "uid")
	uidInt, _ := strconv.ParseUint(uid, 10, 64)

	return uidInt
}

func ControllerGetFeedId(c *gin.Context) uint32 {
	id := gozen.UtilRequestGetParam(c, "feed_id")
	idInt, _ := strconv.ParseUint(id, 10, 32)

	return uint32(idInt)
}

func ControllerGetPsUInt8(c *gin.Context, key string) uint8 {
	id := gozen.UtilRequestGetParam(c, key)
	idInt, _ := strconv.ParseUint(id, 10, 8)

	return uint8(idInt)
}

func ControllerGetPsUInt16(c *gin.Context, key string) uint16 {
	id := gozen.UtilRequestGetParam(c, key)
	idInt, _ := strconv.ParseUint(id, 10, 16)

	return uint16(idInt)
}

func ControllerGetPsUInt32(c *gin.Context, key string) uint32 {
	id := gozen.UtilRequestGetParam(c, key)
	idInt, _ := strconv.ParseUint(id, 10, 32)

	return uint32(idInt)
}

func ControllerGetPsUInt64(c *gin.Context, key string) uint64 {
	id := gozen.UtilRequestGetParam(c, key)
	idInt, _ := strconv.ParseUint(id, 10, 64)

	return uint64(idInt)
}

func ControllerGetPsInt64(c *gin.Context, key string) int64 {
	id := gozen.UtilRequestGetParam(c, key)
	idInt, _ := strconv.ParseInt(id, 10, 64)

	return idInt
}

func ControllerGetPsFloat64(c *gin.Context, key string) float64 {
	id := gozen.UtilRequestGetParam(c, key)
	idF, _ := strconv.ParseFloat(id, 64)

	return idF
}

func ControllerGetPsInt(c *gin.Context, key string) int {
	id := gozen.UtilRequestGetParam(c, key)
	idInt, _ := strconv.ParseInt(id, 10, 64)

	return int(idInt)
}

func ControllerGetUids(c *gin.Context) []uint64 {
	uids := gozen.UtilRequestGetParam(c, "uids")
	uidsSlice := strings.Split(uids, ",")
	var uidsInt64 []uint64
	for _, v := range uidsSlice {
		uid, _ := strconv.ParseUint(v, 10, 64)
		if uid == uint64(0) {
			uidsInt64 = []uint64{}
			break
		}
		uidsInt64 = append(uidsInt64, uid)
	}

	return uidsInt64
}

func ControllerGetPsUInt32s(c *gin.Context, key string) []uint32 {
	ids := gozen.UtilRequestGetParam(c, key)
	idsSlice := strings.Split(ids, ",")
	var Ints []uint32
	for _, v := range idsSlice {
		id, _ := strconv.ParseUint(v, 10, 32)
		nid := uint32(id)
		if nid == uint32(0) {
			Ints = []uint32{}
			break
		}
		Ints = append(Ints, nid)
	}

	return Ints
}

func ControllerGetPsUInt16s(c *gin.Context, key string, space string) []uint16 {
	ids := gozen.UtilRequestGetParam(c, key)
	idsSlice := strings.Split(ids, space)
	var Ints []uint16
	for _, v := range idsSlice {
		id, _ := strconv.ParseUint(v, 10, 16)
		nid := uint16(id)
		if nid == uint16(0) {
			Ints = []uint16{}
			break
		}
		Ints = append(Ints, nid)
	}

	return Ints
}

func ControllerGetPsUInt64s(c *gin.Context, key string) []uint64 {
	ids := gozen.UtilRequestGetParam(c, key)
	idsSlice := strings.Split(ids, ",")
	var Ints []uint64
	for _, v := range idsSlice {
		id, _ := strconv.ParseUint(v, 10, 64)
		if id == uint64(0) {
			Ints = []uint64{}
			break
		}
		Ints = append(Ints, id)
	}

	return Ints
}

func ControllerGetPsStrings(c *gin.Context, key string) []string {
	params := gozen.UtilRequestGetParam(c, key)
	paramsSlice := strings.Split(params, ",")
	var ret []string
	for _, v := range paramsSlice {
		if v != "" {
			ret = append(ret, v)
		}
	}
	return ret
}
`
	ControllerHomeGo = `package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/namejlt/gozen"
	"github.com/gin-gonic/gin"
)

func Welcome(c *gin.Context) {
	now := time.Now().String()
	sysName := gozen.ConfigAppGetString("SysName", "default service")
	content := fmt.Sprintf("Welcome to %s@%s", sysName, now)
	c.String(http.StatusOK, content)
}
`
	ControllerV1AdminTestsGo = `package v1

import (
	"{{.Name}}/model/mapi"
	"{{.Name}}/model/mparam"
	"{{.Name}}/service"
	"time"

	"github.com/namejlt/gozen"
	"github.com/gin-gonic/gin"
)

// AdminTestTest 获取测试请求数据
// @Summary 获取测试请求数据接口
// @Description 根据接口请求，不进行操作，直接返回入参和当前时间
// @Tags 测试相关接口
// @Accept json,x-www-form-urlencoded
// @Produce json
// @Param param query mparam.TestParam true "入参"
// @Success 200 {object} mapi.TestRes "出参"
// @Router /admin/v1/test/test [get]
func AdminTestTest(c *gin.Context) {
	// 初始化结构体时指定初始参数
	param := mparam.TestParam{}
	code, err := gozen.BindParams(c, &param)
	if err != nil {
		gozen.UtilResponseReturnJsonNoP(c, code, err)
		return
	}
	code, data := service.Test(c.Request.Context(), gozen.UtilRequestQueryDataString(c)+param.Name)
	res := mapi.TestRes{
		Data: data,
		Time: time.Now(),
	}
	gozen.UtilResponseReturnJsonNoP(c, code, res)
}

// AdminUserInfo 获取测试用户信息
// @Summary 获取测试用户信息数据接口
// @Description 根据用户id，返回用户信息
// @Tags 用户相关接口
// @Accept json,x-www-form-urlencoded
// @Produce json
// @Param id query string true "入参"
// @Success 200 {object} mredis.User "出参"
// @Router /admin/v1/test/user/info [get]
func AdminUserInfo(c *gin.Context) {
	id := c.Param("id")
	code, data := service.UserInfo(c.Request.Context(), id)
	gozen.UtilResponseReturnJsonNoP(c, code, data)
}
`
	ControllerV1ApiTestsGo = `package v1

import (
	"{{.Name}}/model/mapi"
	"{{.Name}}/model/mparam"
	"{{.Name}}/service"
	"time"

	"github.com/namejlt/gozen"
	"github.com/gin-gonic/gin"
)

// ApiTestTest 获取测试请求数据
// @Summary 获取测试请求数据接口
// @Description 根据接口请求，不进行操作，直接返回入参和当前时间
// @Tags 测试相关接口
// @Accept json,x-www-form-urlencoded
// @Produce json
// @Param param query mparam.TestParam true "入参"
// @Success 200 {object} mapi.TestRes "出参"
// @Router /api/v1/test/test [get]
func ApiTestTest(c *gin.Context) {
	// 初始化结构体时指定初始参数
	param := mparam.TestParam{}
	code, err := gozen.BindParams(c, &param)
	if err != nil {
		gozen.UtilResponseReturnJsonNoP(c, code, err)
		return
	}
	code, data := service.Test(c.Request.Context(), gozen.UtilRequestQueryDataString(c)+param.Name)
	res := mapi.TestRes{
		Data: data,
		Time: time.Now(),
	}
	gozen.UtilResponseReturnJsonNoP(c, code, res)
}

// ApiUserInfo 获取测试用户信息
// @Summary 获取测试用户信息数据接口
// @Description 根据用户id，返回用户信息
// @Tags 用户相关接口
// @Accept json,x-www-form-urlencoded
// @Produce json
// @Param id query string true "入参"
// @Success 200 {object} mredis.User "出参"
// @Router /api/v1/test/user/info [get]
func ApiUserInfo(c *gin.Context) {
	id := c.Param("id")
	code, data := service.UserInfo(c.Request.Context(), id)
	gozen.UtilResponseReturnJsonNoP(c, code, data)
}
`
	ControllerV1AppTestsGo = `package v1

import (
	"{{.Name}}/model/mapi"
	"{{.Name}}/model/mparam"
	"{{.Name}}/service"
	"time"

	"github.com/namejlt/gozen"
	"github.com/gin-gonic/gin"
)

// AppTestTest 获取测试请求数据
// @Summary 获取测试请求数据接口
// @Description 根据接口请求，不进行操作，直接返回入参和当前时间
// @Tags 测试相关接口
// @Accept json,x-www-form-urlencoded
// @Produce json
// @Param param query mparam.TestParam true "入参"
// @Success 200 {object} mapi.TestRes "出参"
// @Router /app/v1/test/test [get]
func AppTestTest(c *gin.Context) {
	// 初始化结构体时指定初始参数
	param := mparam.TestParam{}
	code, err := gozen.BindParams(c, &param)
	if err != nil {
		gozen.UtilResponseReturnJsonNoP(c, code, err)
		return
	}
	code, data := service.Test(c.Request.Context(), gozen.UtilRequestQueryDataString(c)+param.Name)
	res := mapi.TestRes{
		Data: data,
		Time: time.Now(),
	}
	gozen.UtilResponseReturnJsonNoP(c, code, res)
}

// AppUserInfo 获取测试用户信息
// @Summary 获取测试用户信息数据接口
// @Description 根据用户id，返回用户信息
// @Tags 用户相关接口
// @Accept json,x-www-form-urlencoded
// @Produce json
// @Param id query string true "入参"
// @Success 200 {object} mredis.User "出参"
// @Router /app/v1/test/user/info [get]
func AppUserInfo(c *gin.Context) {
	id := c.Param("id")
	code, data := service.UserInfo(c.Request.Context(), id)
	gozen.UtilResponseReturnJsonNoP(c, code, data)
}
`
	ControllerV1IApiTestsGo = `package v1

import (
	"{{.Name}}/model/mapi"
	"{{.Name}}/model/mparam"
	"{{.Name}}/service"
	"time"

	"github.com/namejlt/gozen"
	"github.com/gin-gonic/gin"
)

// IApiTestTest 获取测试请求数据
// @Summary 获取测试请求数据接口
// @Description 根据接口请求，不进行操作，直接返回入参和当前时间
// @Tags 测试相关接口
// @Accept json,x-www-form-urlencoded
// @Produce json
// @Param param query mparam.TestParam true "入参"
// @Success 200 {object} mapi.TestRes "出参"
// @Router /i/v1/test/test [get]
func IApiTestTest(c *gin.Context) {
	// 初始化结构体时指定初始参数
	param := mparam.TestParam{}
	code, err := gozen.BindParams(c, &param)
	if err != nil {
		gozen.UtilResponseReturnJsonNoP(c, code, err)
		return
	}
	code, data := service.Test(c.Request.Context(), gozen.UtilRequestQueryDataString(c)+param.Name)
	res := mapi.TestRes{
		Data: data,
		Time: time.Now(),
	}
	gozen.UtilResponseReturnJsonNoP(c, code, res)
}

// IApiUserInfo 获取测试用户信息
// @Summary 获取测试用户信息数据接口
// @Description 根据用户id，返回用户信息
// @Tags 用户相关接口
// @Accept json,x-www-form-urlencoded
// @Produce json
// @Param id query string true "入参"
// @Success 200 {object} mredis.User "出参"
// @Router /i/v1/test/user/info [get]
func IApiUserInfo(c *gin.Context) {
	id := c.Param("id")
	code, data := service.UserInfo(c.Request.Context(), id)
	gozen.UtilResponseReturnJsonNoP(c, code, data)
}
`
	ControllerV1IApiTplGo = `package v1

import (
	"{{.Name}}/service"

	"github.com/namejlt/gozen"
	"github.com/gin-gonic/gin"
)

// IApiTplTest 获取模板测试数据
// @Summary 获取模板测试数据接口
// @Description 根据接口请求，返回所有传参
// @Tags 测试相关接口
// @Accept json,x-www-form-urlencoded
// @Produce json
// @Param param query string true "入参"
// @Success 200 {object} string "出参"
// @Router /i/v1/tpl/test [get]
func IApiTplTest(c *gin.Context) {
	code, data := service.Test(c, gozen.UtilRequestQueryDataString(c))
	gozen.UtilResponseReturnJsonNoP(c, code, data)
}
`
)
