package gozen

import (
	"os"
	"strconv"
)

var (
	hostname string
)

// 框架初始化
func init() {
	// env
	hostname, _ = os.Hostname()

	//gozen配置
	var err error
	configProject, err = configProjectInit()
	if err != nil {
		panic("gozen config init error:" + err.Error())
	}

	//log
	loggerInit()

	//配置初始化
	if configProject.Base.Format == configFormatJson { // 修改成 json
		cfp = newConfigJsonProcessor()
	} else if configProject.Base.Format == configFormatYaml {
		cfp = newConfigYamlProcessor()
	} else {
		panic("config gozen format error" + configProject.Base.Format)
	}
	if configProject.Base.Mode == configModeNacos {
		// 启动nacos
		nacosClientInit()
		// 更新覆盖本地文件
		nacosToLocal()
	}

	// 读取本地文件
	initConfig()

	// 其他配置获取
	signSwitch = ConfigAppGetString("SignSwitch", "1")
	signAppSecretKey = ConfigAppGetString("AppSecretKey", "std::string")
	appLimitTimeStr := ConfigAppGetString("AppAccessLimitTime", "60")
	appLimitTime, err = strconv.Atoi(appLimitTimeStr)
	if err != nil {
		panic("app config AppAccessLimitTime error" + err.Error())
	}

	// mysql
	initMysql()

	// mongodb
	initMongodb()
}

// 初始化mysql连接池
func initMysql() {
	if ConfigMysqlClusterGetDbCount() > 0 {
		initMysqlCluster() //初始化mysql集群
	} else {
		initMysqlPool(true)  //初始化单个mysql
		initMysqlPool(false) //初始化单个mysql
	}
}

// 初始化mongodb连接池
func initMongodb() {
	if ConfigMongodbClusterGetDbCount() > 0 {
		initMongodbClusterSession() //初始化mongodb集群
	} else {
		initMongodbSession() //初始化单个mongodb
	}
}
