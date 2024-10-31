package gozen

import (
	"testing"
)

func TestConfigPool_configPoolGet(t *testing.T) {
	configPoolInit()
	poolConfig := configPoolGet("grpc-test")
	if poolConfig == nil {
		t.Error("config pool is null")
	}
}

func TestConfigPool_GetAddressRandom(t *testing.T) {
	configPoolInit()
	poolConfig := configPoolGet("grpc-test")
	_, err := poolConfig.GetAddressRandom()
	if err != nil {
		t.Errorf("get error:%v", err.Error())
	}
}
