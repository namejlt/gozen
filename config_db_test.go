package gozen

import (
	"testing"
)

func Test_ConfigDbInit(t *testing.T) {
	configDbInit()
	t.Logf("config db is %v", dbConfig.Mysql.Pool.PoolIdleTimeout)
}
