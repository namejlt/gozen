package gozen

import (
	"sync"
)

var (
	mongodbClusterConfigMap sync.Map
)

func ConfigMongodbClusterGetOne(dbNum uint32) (data *ConfigMongo, ok bool) {
	r, ok := mongodbClusterConfigMap.Load(dbNum)
	if !ok {
		return
	}
	data = r.(*ConfigMongo)
	return
}

func ConfigMongodbClusterGetDbCount() (n int) {
	mongodbClusterConfigMap.Range(func(k interface{}, v interface{}) bool {
		n++
		return true
	})
	return
}

func configMongodbClusterGetDefault() (lists []*ConfigMongo) {
	configMongodb := &ConfigMongo{
		DbNum:      1001,
		DbName:     "",
		Servers:    "",
		ReadOption: "primary",
		Timeout:    1000,
	}
	lists = append(lists, configMongodb)
	return
}

func configMongodbClusterInit() {
	if ConfigMongodbClusterGetDbCount() == 0 {
		configFileName := "mongo_cluster"
		if cfp.configPathExist(configFileName) {
			var mongodbClusterConfig []*ConfigMongo
			defaultMongodbClusterConfig := configMongodbClusterGetDefault()
			err := cfp.configGet(configFileName, nil, &mongodbClusterConfig, defaultMongodbClusterConfig)
			if err != nil {
				panic("configMongodbClusterInit error:" + err.Error())
			}
			for _, v := range mongodbClusterConfig {
				mongodbClusterConfigMap.Store(v.DbNum, v)
			}
		}
	}
}
