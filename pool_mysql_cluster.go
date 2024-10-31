package gozen

import (
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"time"
)

const (
	// 无db时 多租户连接池配置 为防止服务db超过连接数 设置默认值小
	noDbMaxIdleConns = 1
	noDbMaxOpenConns = 10
)

func makeMysqlConnection(isRead bool, selector uint32, dbExtName string, hasDb bool) (conn MysqlConnection, err error) {
	var (
		dbConf *ConfigDbBase
	)
	mysqlConfig, ok := ConfigMysqlClusterGetOne(selector)
	if !ok {
		panic(fmt.Sprintf("makeMysqlConnection error dbnum:%d", selector))
	}
	if isRead {
		dbConf = mysqlConfig.GetClusterRead()
	} else {
		dbConf = mysqlConfig.GetClusterWrite()
	}
	// 判断配置可用性
	if dbConf.Address == "" || dbConf.DbName == "" {
		err = errors.New("dbConf is null")
		return
	}
	var fullDbName string
	if dbExtName == "" {
		fullDbName = dbConf.DbName
	} else {
		fullDbName = dbConf.DbName + "_" + dbExtName
	}
	var dsn string
	if hasDb {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4,utf8&parseTime=True&loc=Local", dbConf.User,
			dbConf.Password, dbConf.Address, dbConf.Port, fullDbName) //指定数据库
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4,utf8&parseTime=True&loc=Local", dbConf.User,
			dbConf.Password, dbConf.Address, dbConf.Port) //不指定数据库
	}
	resultDb, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		LogErrorw(LogNameNet, "connect mysql error",
			LogKNameCommonAddress, dsn,
			LogKNameCommonErr, err,
		)
		return MysqlConnection{}, err
	}

	if ConfigEnvIsDev() {
		resultDb.Logger = resultDb.Logger.LogMode(glogger.Info)
	}

	configPool := mysqlConfig.GetClusterPool()
	db, err := resultDb.DB()
	if err != nil {
		LogErrorw(LogNameNet, "connect mysql error",
			LogKNameCommonAddress, dsn,
			LogKNameCommonErr, err,
		)
		return MysqlConnection{}, err
	}
	if hasDb {
		db.SetMaxIdleConns(configPool.PoolMinCap)
		db.SetMaxOpenConns(configPool.PoolMaxCap)
	} else {
		db.SetMaxIdleConns(noDbMaxIdleConns)
		db.SetMaxOpenConns(noDbMaxOpenConns)
	}
	db.SetConnMaxIdleTime(time.Duration(configPool.PoolIdleTimeout) * time.Millisecond)
	db.SetConnMaxLifetime(time.Duration(configPool.PoolLifeTimeout) * time.Millisecond)
	return MysqlConnection{resultDb, isRead}, err
}
