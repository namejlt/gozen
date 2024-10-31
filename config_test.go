package gozen

import (
	"gopkg.in/yaml.v3"
	"testing"
	"time"
)

type ConfigTest struct {
	Configs map[string]interface{}
}

func configAppGetTest() *ConfigTest {
	return &ConfigTest{map[string]interface{}{"Env": "dev"}}
}

func Test_ConfigGet(t *testing.T) {
	config := &ConfigTest{}
	defaultConfig := configAppGetTest()
	err := cfp.configGet("app", nil, config, defaultConfig)
	if err != nil {
		t.Error(err)
	}
	if config == defaultConfig {
		t.Logf("not find config file")
	} else {
		t.Logf("%v", config)
	}
}

func Test_configProjectGet(t *testing.T) {
	t.Log(configProject)
	time.Sleep(1 * time.Minute)
}

func Test_ConfigLogInit(t *testing.T) {
	LogErrorw(LogNameLogic, "这是应该给测试日iiiiii",
		LogKNameCommonErr, "zheshi 错误的dddd")
}

func Test_ConfigYamlDecodeApp(t *testing.T) {

	in := `
Configs:
  Env: "dev"
  UrlMqServer: "ip"
  AppSecretKey: "std::string"
  AppAccessLimitTime: "0"
  sliceStr: "0,2,3,4"
  sliceInt: "0,2,3,4"
  sliceInt64: "0,2,3,4"
  sliceBool: "true,false,true,false"
  sliceFloat64: "0.2,3.4"
  sliceFloat32: "0.2,3.4"
  test: "这里是测试的值999"
  SignSwitch: "0"
`

	out := ConfigApp{}

	err := yaml.Unmarshal([]byte(in), &out)

	t.Log(out)
	t.Log(err)
}

func Test_ConfigAppGetKey(t *testing.T) {
	a := ConfigAppGetString("test1", "ni谁")
	t.Log(a)

	a = ConfigAppGetValue("test2", "默认值")
	t.Log(a)

	b := ConfigAppGetValue("num001", 10)
	t.Log(b)

	b = ConfigAppGetValue("num001", 100)
	t.Log(b)

	c := ConfigAppGetValue("arr001", []interface{}{1, 2})
	t.Log(c)

	d := ConfigAppGetValueArr("arr002", []string{"1", "ddd"})
	t.Log(d)
}
