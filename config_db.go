package gozen

import (
	"math/rand"
	"sync"
	"time"
)

var (
	dbConfigMux sync.Mutex
	dbConfig    *ConfigDb
)

type ConfigDb struct {
	Mysql ConfigMysql `yaml:"Mysql"`
	Mongo ConfigMongo `yaml:"Mongo"`
}

func NewConfigDb() *ConfigDb {
	return &ConfigDb{}
}

type ConfigDbBase struct {
	Address  string `yaml:"Address"`
	Port     int    `yaml:"Port"`
	User     string `yaml:"User"`
	Password string `yaml:"Password"`
	DbName   string `yaml:"-" json:"-"`
}

type ConfigDbPool struct {
	PoolMinCap      int `yaml:"PoolMinCap"`
	PoolMaxCap      int `yaml:"PoolMaxCap"`
	PoolIdleTimeout int `yaml:"PoolIdleTimeout"` //最大空闲时间
	PoolLifeTimeout int `yaml:"PoolLifeTimeout"` //最大生命时间
}

type ConfigMysql struct {
	DbNum  uint32         `yaml:"DbNum"`  //库号
	DbName string         `yaml:"DbName"` //库名
	Pool   ConfigDbPool   `yaml:"Pool"`   //连接池配置
	Write  ConfigDbBase   `yaml:"Write"`  //写库配置
	Reads  []ConfigDbBase `yaml:"Reads"`  //读库配置
}

type ConfigMongo struct {
	DbNum                  uint32 `yaml:"DbNum"`                  //库号
	DbName                 string `yaml:"DbName"`                 //库名
	Options                string `yaml:"Options"`                //options
	User                   string `yaml:"User"`                   //用户名
	Password               string `yaml:"Password"`               //密码
	Servers                string `yaml:"Servers"`                //服务ip端口
	ReadOption             string `yaml:"ReadOption"`             //读取模式
	Timeout                int    `yaml:"Timeout"`                //具体操作超时时间 秒
	MaxPoolSize            uint64 `yaml:"MaxPoolSize"`            //最大连接数
	MinPoolSize            uint64 `yaml:"MinPoolSize"`            //最小连接数
	SocketTimeout          int    `yaml:"SocketTimeout"`          //读取写入超时
	ConnectTimeout         int    `yaml:"ConnectTimeout"`         //链接超时时间 默认30s
	MaxConnIdleTime        int    `yaml:"MaxConnIdleTime"`        //最大空闲时间
	ServerSelectionTimeout int    `yaml:"ServerSelectionTimeout"` //节点选择超时时间 默认30s
}

func configDbInit() {
	if dbConfig == nil || dbConfig.Mysql.DbName == "" {
		configFileName := "db"
		if cfp.configPathExist(configFileName) {
			dbConfigMux.Lock()
			defer dbConfigMux.Unlock()
			dbConfig = &ConfigDb{}
			defaultDbConfig := configDbGetDefault()
			err := cfp.configGet(configFileName, nil, dbConfig, defaultDbConfig)
			if err != nil {
				panic("configDbInit error:" + err.Error())
			}
		}
	}
}

func configDbClear() {
	dbConfigMux.Lock()
	defer dbConfigMux.Unlock()
	dbConfig = nil
}

func configDbGetDefault() *ConfigDb {
	return &ConfigDb{Mysql: ConfigMysql{
		DbName: "",
		Pool:   ConfigDbPool{5, 20, 60000, 60000},
		Write:  ConfigDbBase{"ip", 33062, "user", "password", ""},
		Reads: []ConfigDbBase{{"ip", 3306, "user", "password", ""},
			{"ip", 33062, "user", "password", ""}}},
		Mongo: ConfigMongo{DbName: "", Servers: "", ReadOption: "primary", Timeout: 1000}}
}

func NewConfigMysql() *ConfigMysql {
	return &ConfigMysql{}
}

func (m *ConfigMysql) GetPool() *ConfigDbPool {
	poolConfig := dbConfig.Mysql.Pool
	if &poolConfig == nil {
		poolConfig = configDbGetDefault().Mysql.Pool
	}
	return &poolConfig
}

func (m *ConfigMysql) GetWrite() *ConfigDbBase {
	writeConfig := dbConfig.Mysql.Write
	if &writeConfig == nil {
		writeConfig = configDbGetDefault().Mysql.Write
	}
	writeConfig.DbName = dbConfig.Mysql.DbName
	return &writeConfig
}

func (m *ConfigMysql) GetRead() (config *ConfigDbBase) {
	readConfigs := dbConfig.Mysql.Reads
	if &readConfigs == nil || len(readConfigs) == 0 {
		return &configDbGetDefault().Mysql.Reads[0]
	}
	count := len(readConfigs)
	if count > 1 {
		rand.Seed(time.Now().UnixNano())

		config = &readConfigs[rand.Intn(count-1)]
	}
	config = &readConfigs[0]
	config.DbName = dbConfig.Mysql.DbName
	return config
}

func (m *ConfigMongo) Get() *ConfigMongo {
	return &dbConfig.Mongo
}
