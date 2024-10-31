package gozen

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	appConfigMux     sync.Mutex
	appConfig        *ConfigApp
	appConfigModTime time.Time
)

type ConfigApp struct {
	Configs map[string]interface{} `yaml:"Configs"`
}

func configAppInit() {
	if appConfig == nil || len(appConfig.Configs) == 0 {
		configFileName := "app"
		appConfigMux.Lock()
		defer appConfigMux.Unlock()
		if cfp.configPathExist(configFileName) {
			appConfig = &ConfigApp{}
			defaultConfig := configAppGetDefault()
			err := cfp.configGet(configFileName, &appConfigModTime, appConfig, defaultConfig)
			if err != nil {
				panic("configAppInit error:" + err.Error())
			}
			go cfp.configReInit(configFileName, &appConfigModTime, appConfig, appConfig, defaultReConfigTicker)
		}
	}
}

func configAppClear() {
	appConfigMux.Lock()
	defer appConfigMux.Unlock()
	appConfig = nil
}

func configAppGetDefault() *ConfigApp {
	return &ConfigApp{map[string]interface{}{"Env": "dev"}}
}

func ConfigAppGetString(key string, defaultConfig string) string {
	config := ConfigAppGet(key)
	if config == nil {
		return defaultConfig
	} else {
		configStr, ok := config.(string)
		if ok {
			if UtilIsEmpty(configStr) {
				configStr = defaultConfig
			}
		} else {
			return defaultConfig
		}

		return configStr
	}
}

func ConfigAppGetValue[T []interface{} | int | int8 | int16 | int32 | int64 | string | float32 | float64 | bool](key string, defaultData T) T {
	config := ConfigAppGet(key)
	if config == nil {
		return defaultData
	} else {
		data, ok := config.(T)

		if ok {
			return data
		}
		return defaultData
	}
}

func ConfigAppGetValueArr[T int | int8 | int16 | int32 | int64 | string | float32 | float64 | bool](key string, defaultData []T) (ret []T) {
	config := ConfigAppGet(key)
	if config == nil {
		return defaultData
	} else {
		data, ok := config.([]interface{})
		if ok {
			for _, v := range data {
				value, vOk := v.(T)
				if vOk {
					ret = append(ret, value)
				}
			}
		} else {
			return defaultData
		}
		return
	}
}

func ConfigAppGet(key string) interface{} {
	config, exists := appConfig.Configs[key]
	if !exists {
		return nil
	}
	return config
}

func ConfigAppFailOverGet(key string) (string, error) {
	return ConfigAppFailoverGet(key)
}

func ConfigAppFailoverGet(key string) (string, error) {
	var server string
	var err error

	failOverConfig := ConfigAppGet(key)
	if failOverConfig == nil {
		err = errors.New(fmt.Sprintf("config %s is null", key))
	} else {
		failOverUrl := failOverConfig.(string)
		if UtilIsEmpty(failOverUrl) {
			err = errors.New(fmt.Sprintf("config %s is empty", key))
		} else {
			failOverArray := strings.Split(failOverUrl, ",")
			randomMax := len(failOverArray)
			if randomMax == 0 {
				err = errors.New(fmt.Sprintf("config %s is empty", key))
			} else {
				var randomValue int
				if randomMax > 1 {
					randomValue = rand.Intn(randomMax)
				} else {
					randomValue = 0
				}
				server = failOverArray[randomValue]
			}
		}
	}
	return server, err
}

func ConfigEnvGet() string {
	strEnv := ConfigAppGet(appJsonEnv)
	return strEnv.(string)
}

func configDocsGet() string {
	return ConfigAppGetString(appJsonDocs, "")
}

func ConfigDocsInstanceNameGet() string {
	return ConfigAppGetString(appJsonDocsInstanceName, "swagger")
}

func configDocsInstanceNameGet(instanceName string) string {
	if instanceName != "" {
		return instanceName
	}
	return ConfigAppGetString(appJsonDocsInstanceName, "swagger")
}

func configDebugVarsGet() string {
	return ConfigAppGetString(appJsonDebugVarsPrefix, "")
}

func ConfigEnvIsDev() bool {
	env := ConfigEnvGet()
	if env == "dev" || env == "debug" {
		return true
	}
	return false
}

func ConfigEnvIsDebug() bool {
	env := ConfigEnvGet()
	if env == "debug" {
		return true
	}
	return false
}

func ConfigEnvIsBeta() bool {
	env := ConfigEnvGet()
	if env == "beta" {
		return true
	}
	return false
}

func configDocsIsExist() bool {
	docs := configDocsGet()
	if docs == "true" {
		return true
	}
	return false
}

func configDebugVarsIsExist() bool {
	if configDebugVarsGet() != "" {
		return true
	}
	return false
}

// ConfigAppGetSlice 获取slice配置，data必须是指针slice *[]，目前支持string,int,int64,bool,float64,float32
func ConfigAppGetSlice(key string, data interface{}) error {
	dataStrConfig := ConfigAppGetString(key, "")
	if UtilIsEmpty(dataStrConfig) {
		return errors.New("config is empty")
	}
	dataStrSlice := strings.Split(dataStrConfig, ",")
	dataType := reflect.ValueOf(data)
	//不是指针Slice
	if dataType.Kind() != reflect.Ptr || dataType.Elem().Kind() != reflect.Slice {
		return errors.New("reflect is not pt or slice")
	}
	dataSlice := dataType.Elem()
	//dataSlice = dataSlice.Slice(0, dataSlice.Cap())
	dataElem := dataSlice.Type().Elem()
	for _, dataStr := range dataStrSlice {
		if UtilIsEmpty(dataStr) {
			continue
		}
		var errConv error
		var item interface{}
		switch dataElem.Kind() {
		case reflect.String:
			item = dataStr
		case reflect.Int:
			item, errConv = strconv.Atoi(dataStr)
		case reflect.Int64:
			item, errConv = strconv.ParseInt(dataStr, 10, 64)
		case reflect.Bool:
			item, errConv = strconv.ParseBool(dataStr)
		case reflect.Float64:
			item, errConv = strconv.ParseFloat(dataStr, 64)
		case reflect.Float32:
			var item64 float64
			item64, errConv = strconv.ParseFloat(dataStr, 32)
			if errConv == nil {
				item = float32(item64)
			}
		default:
			return errors.New("type not support")
		}
		if errConv != nil {
			return errors.New(fmt.Sprintf("convert config failed error:%s", errConv.Error()))
		}
		dataSlice.Set(reflect.Append(dataSlice, reflect.ValueOf(item)))
	}
	return nil
}
