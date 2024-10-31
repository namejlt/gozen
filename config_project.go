package gozen

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

/**

配置入口

- 初始化project.yaml
- nacos
- local

*/

var (
	cfp           configProcessor
	configProject ConfigProject
)

// 框架配置初始化
func initConfig() {
	configAppInit()
	configCodeInit()
	configCacheInit()
	configDbInit()
	configESInit()
	configMysqlClusterInit()
	configMongodbClusterInit()
	configPoolInit()
	configTracerInit()
}

type ConfigProject struct {
	Base struct {
		Mode        string `yaml:"mode"`
		Format      string `yaml:"format"`
		ConfigsPath string `yaml:"configs_path"`
	} `yaml:"base"`
	Nacos struct {
		Addr        string   `yaml:"addr"`
		Port        uint64   `yaml:"port"`
		NamespaceId string   `yaml:"namespace_id"`
		Group       string   `yaml:"group"`
		DataId      []string `yaml:"data_id"`
		TimeoutMs   uint64   `yaml:"timeout_ms"`
		LogLevel    string   `yaml:"log_level"`
		Username    string   `yaml:"username"`
		Password    string   `yaml:"password"`
		Interval    int      `yaml:"interval"`
	} `yaml:"nacos"`
	Log struct {
		Name       string `yaml:"name"`
		Path       string `yaml:"path"`
		Debug      bool   `yaml:"debug"`
		MaxSize    int    `yaml:"max_size"`
		MaxAge     int    `yaml:"max_age"`
		MaxBackups int    `yaml:"max_backups"`
		Compress   bool   `yaml:"compress"`
	} `yaml:"log"`
}

func configProjectInit() (data ConfigProject, err error) {
	// 固定获取配置
	absPath, _ := filepath.Abs(configLocalPath + "/project.yaml")
	var file *os.File
	file, err = os.Open(absPath)
	if err != nil {
		panic("gozen init open path err" + err.Error())
	} else {
		defer file.Close()
		decoder := yaml.NewDecoder(file)
		err = decoder.Decode(&data)
		if err != nil {
			panic("gozen init file yaml decode err" + err.Error())
		}
		// 初始化path
		check, _ := PathExists(data.Base.ConfigsPath)
		if !check {
			err = os.Mkdir(data.Base.ConfigsPath, os.ModePerm)
			if err != nil {
				panic("gozen init file mkdir config path" + err.Error())
			}
		}
	}

	return
}

func GetConfigProjectConfigsPath() (path string) {
	return configProject.Base.ConfigsPath
}
