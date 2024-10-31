package gozen

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

var (
	poolConfigMux sync.Mutex
	poolConfigMap *ConfigPoolMap
)

type ConfigPoolMap struct {
	Configs map[string]ConfigPool `yaml:"Configs"`
}

type ConfigPool struct {
	Address              []string `yaml:"Address"`
	MaxIdle              int      `yaml:"MaxIdle"`              //最大空闲
	MaxActive            int      `yaml:"MaxActive"`            //最多链接限制数 0 不限制
	MaxConcurrentStreams int      `yaml:"MaxConcurrentStreams"` //每一个http链接 并发stream数
	Reuse                bool     `yaml:"Reuse"`                //是否重用链接
}

// func GetAddressRandom get one address random
func (c *ConfigPool) GetAddressRandom() (server string, err error) {
	randomMax := len(c.Address)
	if randomMax == 0 {
		err = errors.New("addess is empty")
	} else {
		var randomValue int
		if randomMax > 1 {
			rand.Seed(time.Now().UnixNano())
			randomValue = rand.Intn(randomMax)
		} else {
			randomValue = 0
		}
		server = c.Address[randomValue]
	}
	return
}

func configPoolInit() {
	if poolConfigMap == nil {
		configFileName := "pool"
		if cfp.configPathExist(configFileName) {
			poolConfigMux.Lock()
			defer poolConfigMux.Unlock()
			if poolConfigMap == nil {
				poolConfigMap = &ConfigPoolMap{Configs: make(map[string]ConfigPool)}
			}
			defaultPoolConfig := configPoolGetDefault()
			err := cfp.configGet(configFileName, nil, poolConfigMap, defaultPoolConfig)
			if err != nil {
				panic("configPoolInit error:" + err.Error())
			}
		}
	}
}

func configPoolGetDefault() *ConfigPoolMap {
	cps := &ConfigPoolMap{Configs: make(map[string]ConfigPool)}
	cps.Configs["test"] = ConfigPool{
		Address: []string{"url"},
		MaxIdle: 20}
	return cps
}

func configPoolGet(poolName string) *ConfigPool {
	poolConfig, ok := poolConfigMap.Configs[poolName]
	if !ok {
		return nil
	}
	return &poolConfig
}
