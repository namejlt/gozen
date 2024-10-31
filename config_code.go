package gozen

import (
	"strconv"
	"sync"
	"time"
)

var (
	codeConfig        sync.Map
	codeConfigModTime time.Time
)

type ConfigCodeList struct {
	Codes map[string]string `yaml:"Codes"`
}

func configCodeInit() {
	defaultData := &ConfigCodeList{Codes: map[string]string{"0": "success"}}
	configFileName := "code"
	if cfp.configPathExist(configFileName) {
		fileCodeConfig := new(ConfigCodeList)
		err := cfp.configGet(configFileName, &codeConfigModTime, fileCodeConfig, defaultData)
		if err != nil {
			panic("configCodeInit error:" + err.Error())
		}
		for k, v := range fileCodeConfig.Codes {
			codeConfig.Store(k, v)
		}
		go cfp.configReInit("code", &codeConfigModTime, appConfig, appConfig, defaultReConfigTicker)
	}
}

func configCodeClear() {
	codeConfig = sync.Map{}
}

func ConfigCodeGetMessage(code int) string {
	msg, exists := codeConfig.Load(strconv.Itoa(code))
	if !exists {
		return "system error"
	}
	return msg.(string)
}

var (
	//监控false的code map
	codeFalseMap = map[int]struct{}{
		PConstCodeServerBusy: {},
	}
)
