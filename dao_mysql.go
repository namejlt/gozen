package gozen

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"log"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"time"
)

type DaoMysql struct {
	TableName string
	Ctx       context.Context
}

func NewDaoMysql() *DaoMysql {
	return &DaoMysql{}
}

type MysqlConnection struct {
	*gorm.DB
	IsRead bool
}

func (p MysqlConnection) Close() {
	if p.DB != nil {
		db, err := p.DB.DB()
		if err == nil {
			_ = db.Close()
		}
	}
}

func (p MysqlConnection) Put() {
	//db database sql inner put
}

var (
	mysqlReadPool  MysqlConnection
	mysqlWritePool MysqlConnection
)

func initMysqlPool(isRead bool) {
	var err error
	config := NewConfigDb()
	configPool := config.Mysql.GetPool()
	if isRead {
		mysqlReadPool.DB, err = initDb(isRead)
		mysqlReadPool.IsRead = isRead
	} else {
		mysqlWritePool.DB, err = initDb(isRead)
		mysqlWritePool.IsRead = isRead
	}
	if err != nil {
		log.Println(fmt.Sprintf("initMysqlPool isread:%v ,error: %v", isRead, err))
		return
	}
	if isRead {
		db, dbErr := mysqlReadPool.DB.DB()
		if dbErr != nil {
			log.Println(fmt.Sprintf("initMysqlPool isread:%v ,error: %v", isRead, dbErr))
			return
		}
		db.SetMaxIdleConns(configPool.PoolMinCap)                                           // 空闲链接
		db.SetMaxOpenConns(configPool.PoolMaxCap)                                           // 最大链接
		db.SetConnMaxLifetime(time.Duration(configPool.PoolLifeTimeout) * time.Millisecond) // 链接最大生命时间
		db.SetConnMaxIdleTime(time.Duration(configPool.PoolIdleTimeout) * time.Millisecond) // 链接最大空闲时间
	} else {
		db, err := mysqlWritePool.DB.DB()
		if err != nil {
			log.Println(fmt.Sprintf("initMysqlPool isread:%v ,error: %v", isRead, err))
			return
		}
		db.SetMaxIdleConns(configPool.PoolMinCap)                                           // 空闲链接
		db.SetMaxOpenConns(configPool.PoolMaxCap)                                           // 最大链接
		db.SetConnMaxLifetime(time.Duration(configPool.PoolLifeTimeout) * time.Millisecond) // 链接最大生命时间
		db.SetConnMaxIdleTime(time.Duration(configPool.PoolIdleTimeout) * time.Millisecond) // 链接最大空闲时间
	}
}

func initDb(isRead bool) (resultDb *gorm.DB, err error) {
	dbConfigMux.Lock()
	defer dbConfigMux.Unlock()
	config := NewConfigDb()
	var dbConf *ConfigDbBase
	if isRead {
		dbConf = config.Mysql.GetRead()
	} else {
		dbConf = config.Mysql.GetWrite()
	}
	// 判断配置可用性
	if dbConf.Address == "" || dbConf.DbName == "" {
		err = errors.New("dbConf is null")
		return
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4,utf8&parseTime=True&loc=Local", dbConf.User,
		dbConf.Password, dbConf.Address, dbConf.Port, dbConf.DbName)
	resultDb, err = gorm.Open(mysql.Open(dsn))
	if err != nil {
		LogErrorw(LogNameMysql, "connect mysql error",
			LogKNameCommonAddress, dsn,
			LogKNameCommonErr, err,
		)
		return resultDb, err
	}
	if ConfigEnvIsDev() {
		resultDb.Logger = resultDb.Logger.LogMode(glogger.Info)
	}
	return resultDb, err
}

func initMysqlPoolConnection(isRead bool) (conn MysqlConnection, err error) {
	if isRead {
		conn = mysqlReadPool
	} else {
		conn = mysqlWritePool
	}
	return
}

func (p *DaoMysql) GetReadOrm() (MysqlConnection, error) {
	return p.getOrm(true)
}

func (p *DaoMysql) GetWriteOrm() (MysqlConnection, error) {
	return p.getOrm(false)
}

func (p *DaoMysql) getOrm(isRead bool) (MysqlConnection, error) {
	return initMysqlPoolConnection(isRead)
}

func (p *DaoMysql) Insert(model interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMysql, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	orm, err := p.GetWriteOrm()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer orm.Put()
	errInsert := orm.Table(p.TableName).Create(model).Error
	SpanErrorFast(span, errInsert)
	if errInsert != nil {
		//记录
		UtilLogError(fmt.Sprintf("insert data error:%s", errInsert.Error()))
	}

	return errInsert
}

func (p *DaoMysql) Select(condition string, data interface{}, field ...[]string) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMysql, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	orm, err := p.GetReadOrm()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer orm.Put()

	err = p.SelectWithConn(&orm, condition, data, field...)
	SpanErrorFast(span, err)
	return err
}

// SelectWithConn SelectWithConn 事务的时候使用
func (p *DaoMysql) SelectWithConn(orm *MysqlConnection, condition string, data interface{}, field ...[]string) error {
	var errFind error
	if len(field) == 0 {
		errFind = orm.Table(p.TableName).Where(condition).Find(data).Error
	} else {
		errFind = orm.Table(p.TableName).Where(condition).Select(field[0]).Find(data).Error
	}
	return errFind
}

func (p *DaoMysql) Update(condition string, sets map[string]interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMysql, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	orm, err := p.GetWriteOrm()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer orm.Put()

	err = orm.Table(p.TableName).Where(condition).Updates(sets).Error
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogError(fmt.Sprintf("update table:%s error:%s, condition:%s, set:%+v", p.TableName, err.Error(), condition, sets))
	}
	return err
}

func (p *DaoMysql) Remove(condition string) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMysql, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	orm, err := p.GetWriteOrm()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer orm.Put()

	err = orm.Table(p.TableName).Where(condition).Delete(nil).Error
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogError(fmt.Sprintf("remove from table:%s error:%s, condition:%s", p.TableName, err.Error(), condition))
	}
	return err
}

func (p *DaoMysql) Find(condition string, data interface{}, skip int, limit int, fields []string, sort string) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMysql, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	orm, err := p.GetReadOrm()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer orm.Put()
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
	if errors.Is(errFind, gorm.ErrRecordNotFound) {
		err = nil
	}
	SpanErrorFast(span, errFind)

	return errFind
}

func (p *DaoMysql) First(condition string, data interface{}, sort string) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMysql, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	orm, err := p.GetReadOrm()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer orm.Put()

	db := orm.Table(p.TableName).Where(condition)
	if !UtilIsEmpty(sort) {
		db = db.Order(sort)
	}

	err = db.First(data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogError(fmt.Sprintf("findone from table:%s error:%s, condition:%s", p.TableName, err.Error(), condition))
	}

	return err
}
