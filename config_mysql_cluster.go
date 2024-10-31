package gozen

import (
	"math/rand"
	"sync"
	"time"
)

var (
	mysqlClusterConfigMap sync.Map
)

func ConfigMysqlClusterGetOne(dbNum uint32) (data *ConfigMysql, ok bool) {
	r, ok := mysqlClusterConfigMap.Load(dbNum)
	if !ok {
		return
	}
	data = r.(*ConfigMysql)
	return
}

func ConfigMysqlClusterGetDbCount() (n int) {
	mysqlClusterConfigMap.Range(func(k interface{}, v interface{}) bool {
		n++
		return true
	})
	return
}

func configMysqlClusterGetDefault() (lists []*ConfigMysql) {
	configMysql := &ConfigMysql{
		DbNum:  1001,
		DbName: "",
		Pool:   ConfigDbPool{5, 20, 60000, 60000},
		Write:  ConfigDbBase{"ip", 33062, "user", "password", ""},
		Reads: []ConfigDbBase{{"ip", 3306, "user", "password", ""},
			{"ip", 33062, "user", "password", ""}}}
	lists = append(lists, configMysql)
	return
}

func configMysqlClusterInit() {
	if ConfigMysqlClusterGetDbCount() == 0 {
		configFileName := "mysql_cluster"
		if cfp.configPathExist(configFileName) {
			var mysqlClusterConfig []*ConfigMysql
			defaultMysqlClusterConfig := configMysqlClusterGetDefault()
			err := cfp.configGet(configFileName, nil, &mysqlClusterConfig, defaultMysqlClusterConfig)
			if err != nil {
				panic("configMysqlClusterInit error:" + err.Error())
			}
			for _, v := range mysqlClusterConfig {
				mysqlClusterConfigMap.Store(v.DbNum, v)
			}
		}
	}
}

func (m *ConfigMysql) GetClusterPool() *ConfigDbPool {
	return &m.Pool
}

func (m *ConfigMysql) GetClusterWrite() (config *ConfigDbBase) {
	config = &m.Write
	config.DbName = m.DbName
	return
}

func (m *ConfigMysql) GetClusterRead() (config *ConfigDbBase) {
	readConfigs := m.Reads
	count := len(readConfigs)
	i := 0
	if count > 1 {
		rand.Seed(time.Now().UnixNano())
		i = rand.Intn(count - 1)
	}
	config = &readConfigs[i]
	config.DbName = m.DbName

	return
}
