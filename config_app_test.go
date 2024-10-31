package gozen

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func Test_ConfigAppReInit(t *testing.T) {
	filePath := "./configs/app.json"

	appFileContent := `{
  "Configs": {
    "Env": "dev",
    "UrlMqServer": "ip",
    "AppSecretKey": "std::string",
    "AppAccessLimitTime": "0",
    "sliceStr": "0,2,3,4",
    "sliceInt": "0,2,3,4",
    "sliceInt64": "0,2,3,4",
    "sliceBool": "true,false,true,false",
    "sliceFloat64": "0.2,3.4",
    "sliceFloat32": "0.2,3.4",
    "SignSwitch": "0"
  }
}`
	err := ioutil.WriteFile(filePath, []byte(appFileContent), os.ModePerm)
	if err != nil {
		t.Errorf("WriteFile:%v", err)
	}
	key := "SignSwitch"
	value := ConfigAppGetString(key, "default")
	if value != "0" {
		t.Errorf("data:%v", value)
	} else {
		t.Logf("data:%v", value)
	}

	appFileToContent := `{
  "Configs": {
    "Env": "dev",
    "UrlMqServer": "ip",
    "AppSecretKey": "std::string",
    "AppAccessLimitTime": "0",
    "sliceStr": "0,2,3,4",
    "sliceInt": "0,2,3,4",
    "sliceInt64": "0,2,3,4",
    "sliceBool": "true,false,true,false",
    "sliceFloat64": "0.2,3.4",
    "sliceFloat32": "0.2,3.4",
    "SignSwitch": "1"
  }
}`

	err = ioutil.WriteFile(filePath, []byte(appFileToContent), os.ModePerm)
	if err != nil {
		t.Errorf("WriteFile:%v", err)
	}
	go cfp.configReInit("app", &appConfigModTime, appConfig, appConfig, 2*time.Second)

	time.Sleep(5 * time.Second)

	value = ConfigAppGetString(key, "default")
	if value != "1" {
		t.Errorf("data:%v", value)
	} else {
		t.Logf("data:%v", value)
	}

	time.Sleep(2 * time.Second)

	value = ConfigAppGetString(key, "default")
	if value != "1" {
		t.Errorf("data:%v", value)
	} else {
		t.Logf("data:%v", value)
	}

	err = ioutil.WriteFile(filePath, []byte(appFileContent), os.ModePerm)
	if err != nil {
		t.Errorf("WriteFile:%v", err)
	}
}

func Test_ConfigAppGetSliceString(t *testing.T) {
	key := "sliceStr"
	var slice []string
	err := ConfigAppGetSlice(key, &slice)
	if err != nil {
		t.Errorf("failed:%s", err.Error())
	} else if len(slice) == 0 {
		t.Error("len is 0")
	} else {
		t.Logf("data:%v", slice)
	}
}

func Test_ConfigAppGetSliceInt(t *testing.T) {
	key := "sliceInt"
	var slice []int
	err := ConfigAppGetSlice(key, &slice)
	if err != nil {
		t.Errorf("failed:%s", err.Error())
	} else if len(slice) == 0 {
		t.Error("len is 0")
	} else {
		t.Logf("data:%v", slice)
	}
}

func Test_ConfigAppGetSliceInt64(t *testing.T) {
	key := "sliceInt64"
	var slice []int64
	err := ConfigAppGetSlice(key, &slice)
	if err != nil {
		t.Errorf("failed:%s", err.Error())
	} else if len(slice) == 0 {
		t.Error("len is 0")
	} else {
		t.Logf("data:%v", slice)
	}
}

func Test_ConfigAppGetSliceBool(t *testing.T) {
	key := "sliceBool"
	var slice []bool
	err := ConfigAppGetSlice(key, &slice)
	if err != nil {
		t.Errorf("failed:%s", err.Error())
	} else if len(slice) == 0 {
		t.Error("len is 0")
	} else {
		t.Logf("data:%v", slice)
	}
}

func Test_ConfigAppGetSliceFloat64(t *testing.T) {
	key := "sliceFloat64"
	var slice []float64
	err := ConfigAppGetSlice(key, &slice)
	if err != nil {
		t.Errorf("failed:%s", err.Error())
	} else if len(slice) == 0 {
		t.Error("len is 0")
	} else {
		t.Logf("data:%v", slice)
	}
}

func Test_ConfigAppGetSliceFloat32(t *testing.T) {
	key := "sliceFloat32"
	var slice []float32
	err := ConfigAppGetSlice(key, &slice)
	if err != nil {
		t.Errorf("failed:%s", err.Error())
	} else if len(slice) == 0 {
		t.Error("len is 0")
	} else {
		t.Logf("data:%v", slice)
	}
}
