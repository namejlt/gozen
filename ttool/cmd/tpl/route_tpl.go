package tpl

var (
	RouteDirName   = "route"
	RouteFilesName = []string{
		"route.go",
	}
	RouteFilesContent = []string{
		RouteRouteGo,
	}

	RouteV1DirName   = "route/v1"
	RouteV1FilesName = []string{
		"tests.go",
		"tpl.go",
	}
	RouteV1FilesContent = []string{
		RouteV1TestsGo,
		RouteV1TplGo,
	}
)

var (
	RouteRouteGo = `package route

import (
	"{{.Name}}/controller"
	"{{.Name}}/middleware"
	"{{.Name}}/pconst"
	"{{.Name}}/route/v1"

	"github.com/gin-gonic/gin"
)

//主页
func RouteHome(parentRoute *gin.Engine) {
	parentRoute.GET("", controller.Welcome)
}

//内部接口
func RouteIApi(parentRoute *gin.Engine) {
	//内部api接口
	RouteV1 := parentRoute.Group(pconst.IAPIV1URL)
	{
		v1.IApiTest(RouteV1)
		v1.IApiTpl(RouteV1)
	}
}

//外部api接口
func RouteApi(parentRoute *gin.Engine) {
	RouteV1 := parentRoute.Group(pconst.APIAPIV1URL)
	{
		v1.ApiTest(RouteV1)
	}
}

//外部app接口
func RouteApp(parentRoute *gin.Engine) {
	RouteV1 := parentRoute.Group(pconst.APPAPIV1URL)
	RouteV1.Use(middleware.Verify())
	{
		v1.AppTest(RouteV1)
	}
}

//内部admin接口
func RouteAdmin(parentRoute *gin.Engine) {
	RouteV1 := parentRoute.Group(pconst.ADMINAPIV1URL)
	{
		v1.AdminTest(RouteV1)
	}
}
`

	RouteV1TestsGo = `package v1

import (
	"{{.Name}}/controller/v1"

	"github.com/gin-gonic/gin"
)

func IApiTest(parentRoute *gin.RouterGroup) {
	router := parentRoute.Group("/test")
	router.GET("/test", v1.IApiTestTest)
	router.GET("/user/info", v1.IApiUserInfo)
}

func ApiTest(parentRoute *gin.RouterGroup) {
	router := parentRoute.Group("/test")
	router.GET("/test", v1.ApiTestTest)
	router.GET("/user/info", v1.ApiUserInfo)
}

func AppTest(parentRoute *gin.RouterGroup) {
	router := parentRoute.Group("/test")
	router.GET("/test", v1.AppTestTest)
	router.GET("/user/info", v1.AppUserInfo)
}

func AdminTest(parentRoute *gin.RouterGroup) {
	router := parentRoute.Group("/test")
	router.GET("/test", v1.AdminTestTest)
	router.GET("/user/info", v1.AdminUserInfo)
}
`

	RouteV1TplGo = `package v1

import (
	"{{.Name}}/controller/v1"

	"github.com/gin-gonic/gin"
)

func IApiTpl(parentRoute *gin.RouterGroup) {
	router := parentRoute.Group("/tpl")
	router.GET("/test", v1.IApiTplTest)
}
`
)
