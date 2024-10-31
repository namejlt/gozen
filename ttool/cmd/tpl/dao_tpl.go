package tpl

var (
	DaoApiDirName   = "dao/api"
	DaoApiFilesName = []string{
		"init.go",
		"user.go",
	}
	DaoApiFilesContent = []string{
		DaoApiInitGo,
		DaoApiUserGo,
	}

	DaoMongoDirName   = "dao/mongo"
	DaoMongoFilesName = []string{
		"user.go",
	}
	DaoMongoFilesContent = []string{
		DaoMongoUserGo,
	}

	DaoMysqlDirName   = "dao/mysql"
	DaoMysqlFilesName = []string{
		"user.go",
	}
	DaoMysqlFilesContent = []string{
		DaoMysqlUserGo,
	}

	DaoRedisDirName   = "dao/redis"
	DaoRedisFilesName = []string{
		"user.go",
	}
	DaoRedisFilesContent = []string{
		DaoRedisUserGo,
	}
)

var (
	DaoApiInitGo = `package api

import "github.com/namejlt/gozen"

var (
	userNeibuUrl string
)

func init() {
	userNeibuUrl = gozen.ConfigAppGetString("UsercNeibu", "http://usercneibu.service.com/")
}
`
	DaoApiUserGo = `package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"{{.Name}}/model/apim"
	"{{.Name}}/pconst"

	"github.com/namejlt/gozen"
)

func GetUserListByIds(ctx context.Context, uids []int64) (users map[int64]*apim.Userc) {
	if len(uids) == 0 {
		return
	}
	uidStrs := make([]string, len(uids))
	for i, uid := range uids {
		uidStrs[i] = fmt.Sprint(uid)
	}
	uidParam := strings.Join(uidStrs, ",")
	apiUrl := userNeibuUrl + "i/v1/users/users?uids=" + uidParam
	ret, err := gozen.CurlGet(ctx, apiUrl, nil, pconst.TimeOutHttpDefault)
	if err != nil {
		gozen.LogErrorw(gozen.LogNameApi, "GetUserListByIds CurlGet",
			gozen.LogKNameCommonErr, err,
			gozen.LogKNameCommonReq, apiUrl,
			gozen.LogKNameCommonRes, string(ret),
		)
		return
	}
	resp := new(apim.UsercResp)
	err = json.Unmarshal(ret, resp)
	if err != nil {
		gozen.LogErrorw(gozen.LogNameApi, "GetUserListByIds json.Unmarshal",
			gozen.LogKNameCommonErr, err,
			gozen.LogKNameCommonReq, apiUrl,
			gozen.LogKNameCommonRes, string(ret),
		)
		return
	}
	if resp.Code == pconst.CODE_COMMON_OK {
		users = resp.Data
	}
	return
}
`
	DaoMongoUserGo = `package mongo

import (
	"github.com/namejlt/gozen"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/net/context"
	"{{.Name}}/model/mmongo"
	"{{.Name}}/pconst"
)

type User struct {
	gozen.DaoMongodbCluster
}

func NewUser(ctx context.Context) *User {
	p := &User{}
	p.Ctx = ctx
	p.CollectionName = pconst.MongoDBTableNameUser
	p.PrimaryKey = "_id"
	return p
}

func (c *User) Find(condition bson.M, limit int, skip int, sortFields bson.D) (data []mmongo.User, err error) {
	err = c.DaoMongo.Find(condition, limit, skip, &data, sortFields)
	return data, err
}

func (c *User) Insert(data *mmongo.User) error {
	data.Name = "aaa"
	return c.DaoMongo.Insert(data)
}
`
	DaoMysqlUserGo = `package mysql

import (
	"context"
	"errors"
	"{{.Name}}/model/mmysql"
	"{{.Name}}/pconst"

	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"

	"github.com/namejlt/gozen"
	"gorm.io/gorm"
)

type User struct {
	gozen.DaoMysqlCluster
}

func NewUser(ctx context.Context) (m *User) {
	m = new(User)
	m.TableName = pconst.MysqlTableNameUser
	m.Ctx = ctx
	m.DbSelector = 1001
	return
}

//获取信息
func (p *User) GetInfo(id string) (data mmysql.User, err error) {
	span, _ := gozen.ExitSpan(p.Ctx, gozen.SpanDaoMysql, gozen.RunFuncName(), v3.SpanLayer_Database)
	defer gozen.SpanEnd(span)
	orm, err := p.GetWriteOrm()
	gozen.SpanErrorFast(span, err)
	if err != nil {
		return
	}
	err = orm.Table(p.TableName).
		Where("id=?", id).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	gozen.SpanErrorFast(span, err)
	return
}

func (p *User) GetList(page int, limitNum int) (list []mmysql.User, count int64, err error) {
	span, _ := gozen.ExitSpan(p.Ctx, gozen.SpanDaoMysql, gozen.RunFuncName(), v3.SpanLayer_Database)
	defer gozen.SpanEnd(span)

	orm, err := p.GetWriteOrm()
	if err != nil {
		gozen.SpanError(span, err.Error())
		return
	}
	gozen.SpanErrorFast(span, err)
	defer orm.PutCluster(&p.DaoMysqlCluster)

	sql := orm.Table(p.TableName)
	err = sql.Count(&count).Error
	gozen.SpanErrorFast(span, err)
	if err != nil {
		gozen.LogErrorw(gozen.LogNameMysql, "mysql GetList Count error",
			gozen.LogKNameCommonErr, err,
		)
		return
	}
	offset := (page - 1) * limitNum
	err = sql.Offset(offset).Limit(limitNum).Order("created_at desc").Find(&list).Error
	gozen.SpanErrorFast(span, err)
	if err != nil {
		gozen.LogErrorw(gozen.LogNameMysql, "mysql GetList Find error",
			gozen.LogKNameCommonErr, err,
		)
	}

	return
}
`
	DaoRedisUserGo = `package redis

import (
	"context"
	"fmt"
	"{{.Name}}/model/mredis"

	"github.com/namejlt/gozen"
)

type User struct {
	gozen.DaoRedisEx
	DefaultKeyName string
}

func NewUser(ctx context.Context) *User {
	dao := new(User)
	dao.KeyName = "user"
	dao.Ctx = ctx
	dao.DefaultKeyName = dao.KeyName
	return dao
}

// 基本信息缓存
func (p *User) SetUserInfoById(id string, data mredis.User, expire int) (err error) {
	key := fmt.Sprintf("%s_id_%s", "info", id)
	err = p.SetEx(key, data, expire)
	return
}

func (p *User) GetUserInfoById(id string) (data mredis.User, b bool, err error) {
	key := fmt.Sprintf("%s_id_%s", "info", id)
	b, err = p.GetRaw(key, &data)
	return
}
`
)
