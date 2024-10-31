package gozen

import (
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	defaultReConfigTicker = 5 * time.Minute //5分钟根据配置文件更新时间重置配置文件数据
)

const (
	configFormatJson = "json"
	configFormatYaml = "yaml"

	configModeLocal = "local"
	configModeNacos = "nacos"

	configLocalPath = "./configs"
)

type configProcessor interface {
	configGet(name string, modTime *time.Time, data interface{}, defaultData interface{}) (err error)
	configReInit(name string, modTime *time.Time, data interface{}, defaultData interface{}, tDuration time.Duration)
	getConfigPath(name string) (absPath string)
	getConfigPathName(name string) (path string)
	configPathExist(name string) bool
}

// =============== configCommonProcessor ==================
// =============== configCommonProcessor ==================
// =============== configCommonProcessor ==================

type configCommonProcessor struct {
	format string
	decode func(r io.Reader, data interface{}) (err error)
}

func newConfigJsonProcessor() *configCommonProcessor {
	p := new(configCommonProcessor)
	p.format = configFormatJson
	p.decode = jsonDecode
	return p
}

func newConfigYamlProcessor() *configCommonProcessor {
	p := new(configCommonProcessor)
	p.format = configFormatYaml
	p.decode = yamlDecode
	return p
}

// 获取配置文件优先级
func (p *configCommonProcessor) configGet(name string, modTime *time.Time, data interface{}, defaultData interface{}) (err error) {
	absPath := p.getConfigPath(name)

	return configGet(name, absPath, modTime, data, defaultData, p.decode)
}

// 定时根据modtime重置文件内容 退出忽略
func (p *configCommonProcessor) configReInit(name string, modTime *time.Time, data interface{}, defaultData interface{}, tDuration time.Duration) {
	t := time.NewTicker(tDuration)
	for {
		select {
		case <-t.C:
			err := p.configGet(name, modTime, data, defaultData)
			if err != nil {
				LogErrorw(LogNameLogic, "configReInit",
					LogKNameCommonErr, err)
			}
		}
	}
}

func (p *configCommonProcessor) getConfigPath(name string) (absPath string) {
	return getConfigPath(name, p.format)
}

func (p *configCommonProcessor) getConfigPathName(name string) (path string) {
	return getConfigPathName(name, p.format)
}

func (p *configCommonProcessor) configPathExist(name string) bool {
	return configPathExist(name, p.format)
}

// ===================  通用函数 =========================
// ===================  通用函数 =========================
// ===================  通用函数 =========================

func configGet(name string, absPath string, modTime *time.Time, data interface{}, defaultData interface{}, decode func(r io.Reader, data interface{}) (err error)) (err error) {
	var file *os.File
	file, err = os.Open(absPath)
	if err != nil {
		LogErrorw(LogNameFile, "open config file failed",
			LogKNameCommonName, name,
			LogKNameCommonErr, err,
		)
		data = defaultData
		return
	} else {
		defer file.Close()
		var fData os.FileInfo
		fData, err = file.Stat()
		if err != nil {
			LogErrorw(LogNameFile, "file config file stat failed",
				LogKNameCommonName, name,
				LogKNameCommonErr, err,
			)
			data = defaultData
			return
		}
		if modTime != nil {
			if *modTime == fData.ModTime() {
				return
			} else { //更新时间同时更新文件
				*modTime = fData.ModTime()
			}
		}
		err = decode(file, data)
		if err != nil {
			LogErrorw(LogNameFile, "decode config file failed",
				LogKNameCommonName, name,
				LogKNameCommonErr, err,
			)
			data = defaultData
			return
		}
	}
	return
}

func yamlDecode(r io.Reader, data interface{}) (err error) {
	decoder := yaml.NewDecoder(r)
	err = decoder.Decode(data)
	if err != nil {
		LogErrorw(LogNameFile, "decode config file failed",
			LogKNameCommonErr, err,
		)
	}
	return
}

func jsonDecode(r io.Reader, data interface{}) (err error) {
	decoder := json.NewDecoder(r)
	err = decoder.Decode(data)
	if err != nil {
		LogErrorw(LogNameFile, "decode config file failed",
			LogKNameCommonErr, err,
		)
	}
	return
}

func getConfigPathName(name string, format string) (path string) {
	path = fmt.Sprintf("%s/%s.%s", configProject.Base.ConfigsPath, name, format)

	return
}

func getConfigPath(name string, format string) (absPath string) {
	var (
		path string
		err  error
	)
	path = fmt.Sprintf("%s/%s.%s", configProject.Base.ConfigsPath, name, format)
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		panic(path + "file not exist")
	}
	absPath, _ = filepath.Abs(path)
	return
}

func configPathExist(name string, format string) bool {
	var (
		path string
		err  error
	)
	path = fmt.Sprintf("%s/%s.%s", configProject.Base.ConfigsPath, name, format)
	_, err = os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// ===================  nacos =========================
// ===================  nacos =========================
// ===================  nacos =========================

// 监听nacos文件 更新

type Nacos struct {
	Client config_client.IConfigClient
}

var (
	nacosObject  Nacos
	nacosContent sync.Map // 后续记录内容
)

// 设置nacos client
func nacosClientInit() {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(
			configProject.Nacos.Addr,
			configProject.Nacos.Port,
		),
	}

	cc := constant.ClientConfig{
		NamespaceId:         configProject.Nacos.NamespaceId, //namespace id
		TimeoutMs:           configProject.Nacos.TimeoutMs,
		Username:            configProject.Nacos.Username,
		Password:            configProject.Nacos.Password,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            configProject.Nacos.LogLevel,
	}

	var err error
	nacosObject.Client, err = clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)

	if err != nil {
		panic("nacos client init fail:" + err.Error())
	}
}

func nacosToLocal() {
	var content string
	var err error
	for _, key := range configProject.Nacos.DataId {
		content, err = nacosObject.Client.GetConfig(vo.ConfigParam{
			DataId: key,
			Group:  configProject.Nacos.Group,
		})

		if err != nil {
			LogErrorw(LogNameApi, "nacosToLocal GetConfig",
				LogKNameCommonErr, err,
				LogKNameCommonKey, key,
			)
			panic("nacosToLocal GetConfig" + err.Error())
		}

		// 覆盖更新本地
		err = os.WriteFile(cfp.getConfigPathName(key), []byte(content), os.ModePerm)

		if err != nil {
			LogErrorw(LogNameApi, "nacosToLocal WriteFile",
				LogKNameCommonErr, err,
				LogKNameCommonKey, key,
			)
			panic("nacosToLocal WriteFile" + err.Error())
		}

		// 异步监听
		_ = nacosListen(key, func(fileKey string, fileData string) {
			_ = os.WriteFile(cfp.getConfigPathName(fileKey), []byte(fileData), os.ModePerm)
		})
	}
}

func nacosListen(key string, f func(fileKey string, content string)) (err error) {
	err = nacosObject.Client.ListenConfig(vo.ConfigParam{
		DataId: key,
		Group:  configProject.Nacos.Group,
		OnChange: func(namespace, group, dataId, data string) {
			f(key, data)
		},
	})
	return
}
