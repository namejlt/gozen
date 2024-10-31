package gozen

import (
	"context"
	"fmt"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"sync"
)

var (
	mysqlClusterReadPool     map[string]MysqlConnection // DbSelector + DbName + DbExtName 保证一个唯一的pool
	mysqlClusterReadPoolMux  sync.Mutex
	mysqlClusterWritePool    map[string]MysqlConnection
	mysqlClusterWritePoolMux sync.Mutex
)

type DaoMysqlCluster struct {
	DbExtName    string //库扩展
	TableExtName string //表扩展
	TableName    string //表名
	DbSelector   uint32 //库号
	Ctx          context.Context
}

func NewDaoMysqlCluster() *DaoMysqlCluster {
	return &DaoMysqlCluster{}
}

func getMysqlClusterPoolKey(selector uint32, dbName string, dbExtName string) string {
	return fmt.Sprintf("%d%s%s", selector, dbName, dbExtName)
}

func initMysqlCluster() {
	mysqlClusterReadPool = make(map[string]MysqlConnection)
	mysqlClusterWritePool = make(map[string]MysqlConnection)
}

func initMysqlClusterPool(isRead bool, selector uint32, dbExtName string, hasDb bool) (conn MysqlConnection, err error) {
	mysqlConfig, ok := ConfigMysqlClusterGetOne(selector)
	if !ok {
		panic(fmt.Sprintf("initMysqlClusterPool error dbnum:%d", selector))
	}
	var selectorKey string
	if hasDb {
		selectorKey = getMysqlClusterPoolKey(selector, mysqlConfig.DbName, dbExtName)
	} else {
		selectorKey = getMysqlClusterPoolKey(selector, "", "")
	}
	if isRead {
		if conn, ok = mysqlClusterReadPool[selectorKey]; !ok {
			mysqlClusterReadPoolMux.Lock()
			defer mysqlClusterReadPoolMux.Unlock()
			conn, err = makeMysqlConnection(isRead, selector, dbExtName, hasDb)
			if err != nil {
				LogErrorw(LogNameMysql, "initMysqlClusterPool makeMysqlConnection mysqlClusterReadPool err:"+err.Error())
				return
			}
			mysqlClusterReadPool[selectorKey] = conn
		}
	} else {
		if conn, ok = mysqlClusterWritePool[selectorKey]; !ok {
			mysqlClusterWritePoolMux.Lock()
			defer mysqlClusterWritePoolMux.Unlock()
			conn, err = makeMysqlConnection(isRead, selector, dbExtName, hasDb)
			if err != nil {
				LogErrorw(LogNameMysql, "initMysqlClusterPool makeMysqlConnection mysqlClusterWritePool err:"+err.Error())
				return
			}
			mysqlClusterWritePool[selectorKey] = conn
		}
	}
	return
}

func (p *DaoMysqlCluster) GetReadOrm() (MysqlConnection, error) {
	return p.getOrm(true)
}

func (p *DaoMysqlCluster) GetWriteOrm() (MysqlConnection, error) {
	return p.getOrm(false)
}

func (p *DaoMysqlCluster) GetWriteOrmNoDb() (conn MysqlConnection, err error) {
	conn, err = p.getOrmNoDb(false)
	return
}

func (p *DaoMysqlCluster) getOrm(isRead bool) (conn MysqlConnection, err error) {
	conn, err = initMysqlClusterPool(isRead, p.DbSelector, p.DbExtName, true)
	return
}

func (p *DaoMysqlCluster) getOrmNoDb(isRead bool) (conn MysqlConnection, err error) {
	conn, err = initMysqlClusterPool(isRead, p.DbSelector, p.DbExtName, false)
	return
}

func (p MysqlConnection) PutCluster(d *DaoMysqlCluster) {
	return
}

func (p *DaoMysqlCluster) Insert(model interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMysql, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	orm, err := p.GetWriteOrm()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer orm.PutCluster(p)
	err = orm.Table(p.TableName).Create(model).Error
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameMysql, "Insert",
			LogKNameCommonErr, err,
			LogKNameCommonName, p.TableName,
			LogKNameCommonData, model,
		)
	}
	return err
}

func (p *DaoMysqlCluster) Update(condition string, sets map[string]interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMysql, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	orm, err := p.GetWriteOrm()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer orm.PutCluster(p)
	err = orm.Table(p.TableName).Where(condition).Updates(sets).Error
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameMysql, "Update",
			LogKNameCommonErr, err,
			LogKNameCommonName, p.TableName,
			LogKNameCommonCondition, condition,
			LogKNameCommonData, sets,
		)
	}
	return err
}

func (p *DaoMysqlCluster) Remove(condition string) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMysql, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	orm, err := p.GetWriteOrm()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer orm.PutCluster(p)
	err = orm.Table(p.TableName).Where(condition).Delete(nil).Error
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameMysql, "Remove",
			LogKNameCommonErr, err,
			LogKNameCommonName, p.TableName,
			LogKNameCommonCondition, condition,
		)
	}
	return err
}

func (p *DaoMysqlCluster) Select(condition string, data interface{}, skip int, limit int, fields []string, sort string) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMysql, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	orm, err := p.GetReadOrm()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer orm.PutCluster(p)
	db := orm.Table(p.TableName).Where(condition)

	if len(fields) > 0 {
		db = db.Select(fields)
	}
	if skip > 0 {
		db = db.Offset(skip)
	}
	if limit > 0 {
		db = db.Limit(limit)
	}
	if sort != "" {
		db = db.Order(sort)
	}
	errFind := db.Find(data).Error
	SpanErrorFast(span, errFind)
	return errFind
}

func (p *DaoMysqlCluster) First(condition string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMysql, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	orm, err := p.GetReadOrm()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer orm.PutCluster(p)
	db := orm.Table(p.TableName).Where(condition)
	errFind := db.First(data).Error
	SpanErrorFast(span, errFind)
	return errFind
}

func (p *DaoMysqlCluster) Count(condition string, data int64) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMysql, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	orm, err := p.GetReadOrm()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer orm.PutCluster(p)
	err = orm.Table(p.TableName).Where(condition).Count(&data).Error
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameMysql, "Remove",
			LogKNameCommonErr, err,
			LogKNameCommonName, p.TableName,
			LogKNameCommonCondition, condition,
		)
	}
	return err
}
