package tpl

var (
	ServiceDirName   = "service"
	ServiceFilesName = []string{
		"base_user.go",
		"tests.go",
	}
	ServiceFilesContent = []string{
		ServiceBaseUserGo,
		ServiceTestsGo,
	}
)

var (
	ServiceBaseUserGo = `package service

import (
	"github.com/namejlt/gozen"
	"context"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"{{.Name}}/dao/mysql"
	"{{.Name}}/dao/redis"
	"{{.Name}}/model/mredis"
	"{{.Name}}/pconst"
)

func getUserInfo(ctx context.Context, id string) (data mredis.User, err error) {
	span, subCtx, _ := gozen.LocalSpan(ctx, gozen.SpanServiceBase, gozen.RunFuncName(), v3.SpanLayer_Unknown)
	defer gozen.SpanEnd(span)
	var b bool
	data, b, err = redis.NewUser(subCtx).GetUserInfoById(id)
	if b {
		return
	}
	if err != nil {
		return
	}
	return resetUserInfo(ctx, id)
}

func resetUserInfo(ctx context.Context, id string) (data mredis.User, err error) {
	span, subCtx, _ := gozen.LocalSpan(ctx, gozen.SpanServiceBase, gozen.RunFuncName(), v3.SpanLayer_Unknown)
	defer gozen.SpanEnd(span)
	r, err := mysql.NewUser(subCtx).GetInfo(id)
	if err != nil {
		return
	}
	data.User = r
	err = redis.NewUser(subCtx).SetUserInfoById(id, data, pconst.TIME_ONE_MINUTE)
	return
}
`
	ServiceTestsGo = `package service

import (
	"github.com/namejlt/gozen"
	"context"
	"{{.Name}}/dao/mysql"
	"{{.Name}}/pconst"

	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

func Test(ctx context.Context, param string) (int, interface{}) {
	span, subCtx, _ := gozen.LocalSpan(ctx, gozen.SpanServiceOpen, gozen.RunFuncName(), v3.SpanLayer_Unknown)
	defer gozen.SpanEnd(span)
	list, count, err := mysql.NewUser(subCtx).GetList(1, 1000)
	if err != nil {
		return pconst.CODE_COMMON_SERVER_BUSY, err
	}
	data := struct {
		Param string
		List  interface{}
		Count int64
	}{
		param,
		list,
		count,
	}
	return pconst.CODE_COMMON_OK, data
}

func UserInfo(ctx context.Context, id string) (int, interface{}) {
	span, subCtx, _ := gozen.LocalSpan(ctx, gozen.SpanServiceOpen, gozen.RunFuncName(), v3.SpanLayer_Unknown)
	defer gozen.SpanEnd(span)
	data, err := getUserInfo(subCtx, id)
	if err != nil {
		return pconst.CODE_COMMON_SERVER_BUSY, err
	}
	return pconst.CODE_COMMON_OK, data
}
`
)
