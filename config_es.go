package gozen

import (
	"sync"
)

var (
	esConfigMux sync.Mutex
	esConfig    *ConfigES
)

type ConfigES struct {
	Address             []string `yaml:"Address"`
	Timeout             int      `yaml:"Timeout"`
	TransportMaxIdel    int      `yaml:"TransportMaxIdel"`
	HealthcheckEnabled  bool     `yaml:"HealthcheckEnabled"`
	HealthcheckTimeout  int      `yaml:"HealthcheckTimeout"`
	HealthcheckInterval int      `yaml:"HealthcheckInterval"`
	SnifferEnabled      bool     `yaml:"SnifferEnabled"`
	Account             string   `yaml:"Account"`
	Password            string   `yaml:"Password"`
}

func configESInit() {
	if esConfig == nil || len(esConfig.Address) == 0 {
		configFileName := "es"
		if cfp.configPathExist(configFileName) {
			esConfigMux.Lock()
			defer esConfigMux.Unlock()
			if esConfig == nil || len(esConfig.Address) == 0 {
				esConfig = &ConfigES{}
				defaultESConfig := configESGetDefault()
				err := cfp.configGet(configFileName, nil, esConfig, defaultESConfig)
				if err != nil {
					panic("configESInit error:" + err.Error())
				}
				if esConfig.HealthcheckTimeout == 0 {
					esConfig.HealthcheckTimeout = 1
				}
				if esConfig.HealthcheckInterval == 0 {
					esConfig.HealthcheckInterval = 60
				}
			}
		}
	}
}

func configESClear() {
	esConfigMux.Lock()
	defer esConfigMux.Unlock()
	esConfig = nil
}

func configESGetDefault() *ConfigES {
	return &ConfigES{Address: []string{"url"},
		HealthcheckTimeout:  1,
		HealthcheckInterval: 60,
		Timeout:             3000,
		TransportMaxIdel:    10}
}

func configESGetAddress() []string {
	return esConfig.Address
}

func configESGet() *ConfigES {
	return esConfig
}
