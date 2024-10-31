package gozen

// 公共code 0 - 10000
const (
	//common

	PConstCodeOK               = 0    //成功
	PConstCodeCommonOK         = 1001 //成功
	PConstCodeAccessFail       = 1002 //无权访问
	PConstCodeServerBusy       = 1003 //服务器繁忙
	PConstCodeParamsIncomplete = 1004 //参数不全
	PConstCodeUserNoLogin      = 1005 //用户未登录
	PConstCodeUserNoLoginApp   = 1024 //APP用户未登录
	PConstCodeBusinessError    = 1010 //业务方错误

	//mysql 2000

	//mongodb 2100

	//redis 2200

	//es 2300

	//api 2400

	//ao 2500

	//gRpc 2600

	//tmq 2700

	//mq 2800
)
