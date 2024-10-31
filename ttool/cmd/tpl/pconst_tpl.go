package tpl

var (
	PConstDirName   = "pconst"
	PConstFilesName = []string{
		"api.go",
		"code.go",
		"db.go",
	}
	PConstFilesContent = []string{
		PConstApiGo,
		PConstCodeGo,
		PConstDbGo,
	}
)

var (
	PConstApiGo = `package pconst

import "time"

// Constants for API
const (
	IAPIRoot      = "/i"
	APPAPIRoot    = "/app"
	APIAPIRoot    = "/api"
	ADMINAPIRoot  = "/admin"
	APIV1Version  = "v1"
	APIV2Version  = "v2"
	APIV3Version  = "v3"
	APIV4Version  = "v4"
	IAPIV1URL     = IAPIRoot + "/" + APIV1Version
	IAPIV2URL     = IAPIRoot + "/" + APIV2Version
	IAPIV3URL     = IAPIRoot + "/" + APIV3Version
	IAPIV4URL     = IAPIRoot + "/" + APIV4Version
	APPAPIV1URL   = APPAPIRoot + "/" + APIV1Version
	APIAPIV1URL   = APIAPIRoot + "/" + APIV1Version
	ADMINAPIV1URL = ADMINAPIRoot + "/" + APIV1Version
	APPAPIV2URL   = APPAPIRoot + "/" + APIV2Version
	APIAPIV2URL   = APIAPIRoot + "/" + APIV2Version
	//time

	TIME_FORMAT_Y_M_D_H_I_S   = "2006-01-02 15:04:05"
	TIME_FORMAT_Y_M_D_H_I_S_2 = "2006/01/02 15:04:05"
	TIME_FORMAT_Y_M_D         = "2006.01.02"
	TIME_FORMAT_Y_M_D_        = "2006-01-02"
	TIME_FORMAT_Y_MS_D_       = "2006-January-02"
	TIME_FORMAT_M_D_H_I       = "01.02 15:04"
	TIME_FORMAT_H_I           = "15:04"

	TIME_ONE_SECOND = 1
	TIME_TWO_SECOND = 2
	TIME_ONE_MINUTE = 60
	TIME_TEN_MINUTE = 600
	TIME_ONE_HOUR   = 3600
	TIME_ONE_DAY    = 86400
	TIME_THREE_DAY  = 259200
	TIME_ONE_WEEK   = 604800
	TIME_ONE_MONTH  = 2592000
	TIME_ONE_YEAR   = 31536000

	TimeOutHttpDefault = 10 * time.Second

	//common

	COMMON_PAGE_LIMIT_NUM_10  = 10
	COMMON_PAGE_LIMIT_NUM_20  = 20
	COMMON_PAGE_LIMIT_NUM_MAX = 20

	COMMON_ERROR_RETRY_LIMIT_NUM = 5
)
`
	PConstCodeGo = `package pconst

/*

code 区间

通用code

1001 - 1999

项目通用code



内部接口


外部接口


管理后台接口


*/

const (
	//common

	CODE_ERROR_OK                 = 0
	CODE_COMMON_OK                = 1001
	CODE_COMMON_ACCESS_FAIL       = 1002
	CODE_COMMON_SERVER_BUSY       = 1003
	CODE_COMMON_PARAMS_INCOMPLETE = 1004
	CODE_COMMON_USER_NO_LOGIN     = 1005
)
`
	PConstDbGo = `package pconst

const (
	MongoDBTableNameUser = "user"

	MysqlTableNameUser = "user"
)
`
)
