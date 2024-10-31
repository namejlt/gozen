package gozen

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/atomic"
	"log"
	"math/rand"
	"net"
	"reflect"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var (
	dynamicAddress                RedisDynamicAddress
	cachePoolDefaultActive        = 1   // 默认池子连接数
	cachePoolDefaultCycleTime     = 300 // 默认地址更新时间
	cachePoolDefaultCycleFlagTime = 5   // 默认地址立即更新 间隔时间

	redisExecCount  atomic.Int64 // 统计执行次数
	redisPExecCount atomic.Int64 // 统计执行次数
)

type RedisDynamicAddress struct {
	Address      []string      // 地址列表
	AddressLen   int32         // 地址总数
	AddressIndex atomic.Int32  // 地址获取索引 从0到AddressLen-1
	ErrorConnM   sync.Map      // 地址连接错误 k 是索引 v 是空
	ErrorConnNum atomic.Int32  // 地址连接错误数
	FlagCount    atomic.Int32  // 更新标记统计
	Flag         chan struct{} // 是否主动重置
	L            sync.RWMutex
}

type DaoRedis struct {
	KeyName    string
	Persistent bool            // 持久化key
	Ctx        context.Context // 必须定义 否则报错
}

// 统计数据

func getRedisExecCount(persistent bool) (num int64) {
	if persistent {
		num = redisPExecCount.Load()
		redisPExecCount.Sub(num)
	} else {
		num = redisExecCount.Load()
		redisExecCount.Sub(num)
	}
	return
}

func incrRedisExecCount(persistent bool) {
	if persistent {
		redisPExecCount.Inc()
	} else {
		redisExecCount.Inc()
	}
}

/*type redisPool struct {
	redisPool     *pools.ResourcePool
	redisPoolMux  sync.RWMutex
	redisPPool    *pools.ResourcePool // 持久化Pool
	redisPPoolMux sync.RWMutex
}*/

func initDynamicRedisAddress() {
	dynamic := ConfigCacheGetRedisDynamic()
	if dynamic.IsDynamic {
		dynamicAddress.Flag = make(chan struct{})
		if dynamic.CycleTime <= 0 { // 默认5分钟 更新一次
			dynamic.CycleTime = cachePoolDefaultCycleTime
		}
		if dynamic.CycleFlagTime <= 0 { // 默认10s 最多一次
			dynamic.CycleFlagTime = cachePoolDefaultCycleFlagTime
		}
		//更新地址标记逻辑
		go GoFuncOne(func() error {
			t := time.NewTicker(time.Duration(dynamic.CycleFlagTime) * time.Second)
			for {
				select {
				case <-t.C:
					if dynamicAddress.FlagCount.Load() > 0 {
						dynamicAddress.Flag <- struct{}{}
					}
					dynamicAddress.FlagCount.Store(0)
				}
			}
		})
		//更新地址逻辑
		go GoFuncOne(func() error {
			t := time.NewTicker(time.Duration(dynamic.CycleTime) * time.Second)
			for {
				//更新地址
				dynamicAddress.L.Lock()
				err, address := getDynamicRedisAddress(dynamic.DynamicAddress)
				if err == nil && len(address) > 0 {
					dynamicAddress.Address = address
					dynamicAddress.AddressLen = int32(len(address)) //获取这么多地址，后续地址随机顺序获取
					dynamicAddress.AddressIndex.Store(0)
					dynamicAddress.FlagCount.Store(0)
				} else {
					//获取失败则 本地不变
					LogErrorw(LogNameApi, "InitDynamicRedisAddress getDynamicRedisAddress ret null",
						LogKNameCommonErr, err,
					)
				}
				dynamicAddress.L.Unlock()
				select {
				case <-t.C:
					//定时重置
				case <-dynamicAddress.Flag:
					//立即重置
				}
			}
		})
	}
}

func (p *RedisDynamicAddress) GetAddress() (addr string, index int32) {
	addr = p.Address[p.AddressIndex.Inc()%p.AddressLen] //保证顺序获取地址
	return
}

func (p *RedisDynamicAddress) dealConnError(index int32) {
	_, b := p.ErrorConnM.LoadOrStore(index, struct{}{})
	if !b { //保存成功 记录新错误数量
		p.ErrorConnNum.Inc()
	}
	if p.ErrorConnNum.Load() >= p.AddressLen {
		p.FlagCount.Inc()
	}
	return
}

// no proxy

var noProxyLocalInfo sync.Map //proxy_key_persistent => noProxyLocalClusterInfo

func setNoProxyLocalInfo(persistent bool, data noProxyLocalClusterInfo) {
	noProxyLocalInfo.Store(getNoProxyLocalInfoKey(persistent), data)
}

func getNoProxyLocalInfo(persistent bool) (data noProxyLocalClusterInfo) {
	ret, _ := noProxyLocalInfo.Load(getNoProxyLocalInfoKey(persistent))
	data = ret.(noProxyLocalClusterInfo)
	return
}

func getNoProxyLocalInfoKey(persistent bool) string {
	return fmt.Sprintf("proxy_key_%v", persistent)
}

// initNoProxyRedisAddressAll 平台去代理模式更新地址 OR 同步更新redis客户端
func initNoProxyRedisAddressAll() {
	initNoProxyRedisAddress(true)
	initNoProxyRedisAddress(false)
}

func initNoProxyRedisAddress(persistent bool) {
	var redisConfig ConfigCacheRedis
	if !persistent {
		redisConfig = cacheConfig.Redis
	} else {
		redisConfig = cacheConfig.RedisP
	}
	if cacheConfig.NoProxy.IsNoProxy {

		// 初始化更新地址 失败panic
		clusterInfo, err := getNoProxyRedisAddress(cacheConfig.NoProxy.ServiceAddress, cacheConfig.NoProxy.AppCode, redisConfig.Bid, cacheConfig.NoProxy.Iv)
		if err != nil {
			panic("initNoProxyRedisAddress getNoProxyRedisAddress fail" + err.Error())
		}

		if ConfigEnvIsDebug() {
			log.Printf("首次获取 getNoProxyRedisAddress %v", clusterInfo)
		}

		setNoProxyLocalInfo(persistent, clusterInfo)
		initRedisPoll(persistent)

		// 定期监控上报，可能更新地址和redis客户端
		// 更新地址标记逻辑
		go GoFuncOne(func() error {
			t := time.NewTicker(time.Duration(cacheConfig.NoProxy.MonitorTime) * time.Second)
			for {
				select {
				case <-t.C:
					// 心跳
					heartInfo := getSendHeartInfo(persistent)
					isNew, newClusterInfo, heartErr := sendNoProxyRedisHeart(cacheConfig.NoProxy.MonitorAddress, cacheConfig.NoProxy.AppCode, redisConfig.Bid, cacheConfig.NoProxy.Iv, heartInfo)
					if heartErr != nil {
						LogErrorw(LogNameApi, "initNoProxyRedisAddress sendNoProxyRedisHeart error",
							LogKNameCommonReq, heartInfo,
							LogKNameCommonErr, heartErr,
						)
					} else {
						if isNew { //需进行更新本地信息 并更新redis client
							setNoProxyLocalInfo(persistent, newClusterInfo)
							initRedisPoll(persistent)
						}
					}
				}
			}
		})
	}
}

// getSendHeartInfo 获取心跳数据
func getSendHeartInfo(persistent bool) (info noProxyHeartInfo) {
	info.Lang = cacheConfig.NoProxy.Lang
	info.AppCode = cacheConfig.NoProxy.AppCode
	info.ClientIp = SubStr(hostname, 0, 49)

	item := noProxyHeartInfoClusterInfo{}
	localInfo := getNoProxyLocalInfo(persistent)
	item.Bid = localInfo.Bid
	item.ClusterInfo = localInfo.ClusterInfo
	item.ExecCount = getRedisExecCount(persistent)

	info.ClusterInfos = append(info.ClusterInfos, item)

	return
}

/*
func (p *redisPool) Get(persistent bool) *pools.ResourcePool {
	if persistent {
		p.redisPPoolMux.RLock()
		defer p.redisPPoolMux.RUnlock()
		return p.redisPPool
	} else {
		p.redisPoolMux.RLock()
		defer p.redisPoolMux.RUnlock()
		return p.redisPool
	}
}

func (p *redisPool) Set(pool *pools.ResourcePool, persistent bool) {
	if persistent {
		p.redisPPoolMux.Lock()
		defer p.redisPPoolMux.Unlock()
		p.redisPPool = pool
	} else {
		p.redisPoolMux.Lock()
		defer p.redisPoolMux.Unlock()
		p.redisPool = pool
	}
}

func (p *redisPool) Put(resource redis.Cmdable, persistent bool) {
	return
}

var daoPool redisPool

type ResourceConn struct {
	redis.Conn
}

func (r ResourceConn) Close() {
	r.Conn.Close()
}*/

func RedisGetAddress(conf *ConfigCacheRedis) (addr string, index int32, err error) {
	dynamic := ConfigCacheGetRedisDynamic()
	if dynamic.IsDynamic { //这里根据主动重置标记 进行主动更新本地IP
		dynamicAddress.L.RLock()
		defer dynamicAddress.L.RUnlock()
		if len(dynamicAddress.Address) > 0 {
			addr, index = dynamicAddress.GetAddress()
			return
		}
		resetRedisGetAddress()
		var address []string
		err, address = getDynamicRedisAddress(dynamic.DynamicAddress)
		if err != nil {
			return
		}
		if len(address) > 0 { //初始默认0
			addr = address[0]
			return
		}
	} else { //静态 忽略index 每次随机
		addr = conf.Address[rand.Intn(conf.addressLen)]
	}
	return
}

// 重置redis 地址标记
func resetRedisGetAddress() {
	dynamicAddress.FlagCount.Inc()
}

/*func dial(persistent bool) (conn redis.Conn, err error) {
	cacheConfig := ConfigCacheGetRedisWithConn(persistent)
	var addr string
	var index int32
	addr, index, err = RedisGetAddress(cacheConfig)
	if err == nil && addr != "" {
		var opt []redis.DialOption
		opt = append(opt, redis.DialConnectTimeout(time.Duration(cacheConfig.ConnectTimeout)*time.Millisecond),
			redis.DialReadTimeout(time.Duration(cacheConfig.ReadTimeout)*time.Millisecond),
			redis.DialWriteTimeout(time.Duration(cacheConfig.WriteTimeout)*time.Millisecond))
		if cacheConfig.Password != "" {
			opt = append(opt, redis.DialPassword(cacheConfig.Password))
		}
		conn, err = redis.Dial("tcp", addr, opt...)
		if err != nil {
			LogErrorw(LogNameNet, "dial redis pool error", LogKNameCommonErr, err)
			if isCacheGetRedisDynamic() { //记录动态地址获取错误
				dynamicAddress.dealConnError(index)
			}
		}
		return conn, err
	} else {
		return nil, errors.New("redis address length is 0")
	}
}*/

type CacheClient struct {
	l      sync.RWMutex
	client map[string]redis.UniversalClient
}

func (p *CacheClient) getClient(key string) (c redis.UniversalClient, ok bool) {
	p.l.RLock()
	defer p.l.RUnlock()
	c, ok = p.client[key]
	return
}

func (p *CacheClient) setClient(key string, c redis.UniversalClient) {
	p.l.Lock()
	defer p.l.Unlock()
	p.client[key] = c
}

func getClientKey(persistent bool) string {
	return fmt.Sprintf("client_%v", persistent)
}

var (
	cacheClient = CacheClient{
		client: make(map[string]redis.UniversalClient),
	} //通用client 根据不同场景保存不同client
)

// 初始化redis连接池
func initRedisPoll(persistent bool) {
	redisConfig := ConfigCacheGetRedisWithConn(persistent)
	if redisConfig.PoolMinActive == 0 { //默认连接数1
		redisConfig.PoolMinActive = cachePoolDefaultActive
		redisConfig.PoolMaxActive = cachePoolDefaultActive
	}
	// 因redis代理ip是变动的，所以连接池不能使用固定ip

	key := getClientKey(persistent)

	if redisConfig.Cluster {
		c := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        redisConfig.Address,
			Password:     redisConfig.Password,
			DialTimeout:  time.Duration(redisConfig.ConnectTimeout) * time.Millisecond,
			ReadTimeout:  time.Duration(redisConfig.ReadTimeout) * time.Millisecond,
			WriteTimeout: time.Duration(redisConfig.WriteTimeout) * time.Millisecond,
			PoolSize:     redisConfig.PoolMaxActive,
			MinIdleConns: redisConfig.PoolMinActive,
			IdleTimeout:  time.Duration(redisConfig.PoolIdleTimeout) * time.Millisecond,
		})
		cacheClient.setClient(key, c)
	} else {
		c := redis.NewClient(&redis.Options{
			Dialer: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
				//根据addr 里面配置来获取client 每次随机 addr
				addr, _, err = RedisGetAddress(redisConfig)
				if err == nil && addr != "" {
					var d net.Dialer
					return d.DialContext(ctx, network, addr)
				} else {
					return nil, errors.New("redis address length is 0")
				}
			},
			Password:     redisConfig.Password,
			DialTimeout:  time.Duration(redisConfig.ConnectTimeout) * time.Millisecond,
			ReadTimeout:  time.Duration(redisConfig.ReadTimeout) * time.Millisecond,
			WriteTimeout: time.Duration(redisConfig.WriteTimeout) * time.Millisecond,
			PoolSize:     redisConfig.PoolMaxActive,
			MinIdleConns: redisConfig.PoolMinActive,
			IdleTimeout:  time.Duration(redisConfig.PoolIdleTimeout) * time.Millisecond,
		})
		cacheClient.setClient(key, c)
	}
}

// 获取redis连接
func (p *DaoRedis) getRedisConn() (c redis.UniversalClient, err error) {
	key := getClientKey(p.Persistent)
	c, ok := cacheClient.getClient(key)
	if !ok {
		err = errors.New("cacheClient null")
		return
	}
	incrRedisExecCount(p.Persistent) //执行统计
	return
}

// GetRedisConn 获取redis连接
func (p *DaoRedis) GetRedisConn() (c redis.UniversalClient, err error) {
	key := getClientKey(p.Persistent)
	c, ok := cacheClient.getClient(key)
	if !ok {
		err = errors.New("cacheClient null")
		return
	}
	incrRedisExecCount(p.Persistent) //执行统计
	return
}

func (p *DaoRedis) checkCtx() {
	if p.Ctx == nil {
		p.Ctx = context.Background()
	}
}

func (p *DaoRedis) getKey(key string) string {
	redisConfig := ConfigCacheGetRedisBaseWithConn(p.Persistent)
	prefixRedis := redisConfig.Prefix
	if strings.Trim(key, " ") == "" {
		return fmt.Sprintf("%s:%s", prefixRedis, p.KeyName)
	}
	return fmt.Sprintf("%s:%s:%s", prefixRedis, p.KeyName, key)
}

func (p *DaoRedis) doSet(cmd string, key string, value interface{}, expire int, fields ...string) (interface{}, error) {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return nil, err
	}
	key = p.getKey(key)
	data, err := json.Marshal(value)
	if err != nil {
		LogErrorw(LogNameLogic, "redis marshal data to json error", LogKNameCommonData, cmd, LogKNameCommonErr, err)
		return nil, err
	}
	if expire == 0 {
		redisConfig := ConfigCacheGetRedisBaseWithConn(p.Persistent)
		expire = redisConfig.Expire
	}
	var reply interface{}
	var errDo error
	p.checkCtx()
	if len(fields) == 0 {
		if expire > 0 && strings.ToUpper(cmd) == "SET" {
			reply, errDo = redisClient.Do(p.Ctx, cmd, key, data, "ex", expire).Result()
		} else {
			reply, errDo = redisClient.Do(p.Ctx, cmd, key, data).Result()
		}
	} else {
		field := fields[0]
		reply, errDo = redisClient.Do(p.Ctx, cmd, key, field, data).Result()
	}
	if errDo != nil {
		LogErrorw(LogNameRedis, "run redis command error", LogKNameCommonData, cmd, LogKNameCommonErr, errDo, LogKNameCommonKey, key, LogKNameCommonFields, fields, LogKNameCommonValue, value)
		return nil, errDo
	}
	//set expire
	if expire > 0 && strings.ToUpper(cmd) != "SET" {
		_, errExpire := redisClient.Do(p.Ctx, "EXPIRE", key, expire).Result()
		if errExpire != nil {
			LogErrorw(LogNameRedis, "run redis EXPIRE command error",
				"err", errExpire,
				"key", key,
				"expire", expire,
			)
		}
	}
	return reply, errDo
}

func (p *DaoRedis) doSetNX(cmd string, key string, value interface{}, expire int, field ...string) (int64, bool) {
	reply, err := p.doSet(cmd, key, value, expire, field...)
	if err != nil {
		return 0, false
	}
	replyInt, ok := reply.(int64)
	if !ok {
		LogErrorw(LogNameRedis, "HSetNX reply to int error",
			LogKNameCommonKey, key,
			LogKNameCommonTime, expire,
		)
		return 0, false
	}
	return replyInt, true
}
func (p *DaoRedis) doMSet(cmd string, key string, value map[string]interface{}) (interface{}, error) {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return nil, err
	}
	var args []interface{}
	args = append(args, cmd)
	if key != "" {
		key = p.getKey(key)
		args = append(args, key)
	}
	for k, v := range value {
		data, errJson := json.Marshal(v)
		if errJson != nil {
			LogErrorw(LogNameLogic, "redis marshal data error",
				LogKNameCommonData, cmd,
				LogKNameCommonErr, errJson,
				LogKNameCommonValue, v,
			)
			return nil, errJson
		}
		if key == "" {
			args = append(args, p.getKey(k), data)
		} else {
			args = append(args, k, data)
		}
	}
	var reply interface{}
	var errDo error
	p.checkCtx()
	reply, errDo = redisClient.Do(p.Ctx, args...).Result()
	if errDo != nil {
		LogErrorw(LogNameRedis, "run redis command error",
			LogKNameCommonData, cmd,
			LogKNameCommonErr, errDo,
			LogKNameCommonKey, key,
			LogKNameCommonValue, value,
		)
		return nil, errDo
	}
	return reply, errDo
}

func (p *DaoRedis) doGet(cmd string, key string, value interface{}, fields ...string) (bool, error) {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return false, err
	}
	key = p.getKey(key)
	var result interface{}
	var errDo error
	var args []interface{}
	args = append(args, cmd)
	args = append(args, key)
	for _, f := range fields {
		args = append(args, f)
	}
	p.checkCtx()
	result, errDo = redisClient.Do(p.Ctx, args...).Result()
	if errors.Is(errDo, redis.Nil) {
		value = nil
		return false, nil
	}
	if errDo != nil {
		LogErrorw(LogNameRedis, "run redis command error",
			LogKNameCommonData, cmd,
			LogKNameCommonErr, errDo,
			LogKNameCommonKey, key,
			LogKNameCommonFields, fields,
		)
		return false, errDo
	}
	if reflect.TypeOf(result).Kind() == reflect.Slice {
		byteResult := result.([]byte)
		strResult := string(byteResult)
		if strResult == "[]" {
			return true, nil
		}
	}
	errorJson := JSONDecodeUseNumber([]byte(result.(string)), value)
	if errorJson != nil {
		if reflect.TypeOf(value).Kind() == reflect.Ptr && reflect.TypeOf(value).Elem().Kind() == reflect.String {
			var strValue string
			strValue = string(result.([]byte))
			v := value.(*string)
			*v = strValue
			value = v
			return true, nil
		}
		LogErrorw(LogNameRedis, "get redis command result error",
			LogKNameCommonData, cmd,
			LogKNameCommonErr, errorJson,
		)
		return false, errorJson
	}
	return true, nil
}

func (p *DaoRedis) doMGet(cmd string, args []interface{}, value interface{}) error {
	refValue := reflect.ValueOf(value)
	if refValue.Kind() != reflect.Ptr || refValue.Elem().Kind() != reflect.Slice {
		return errors.New(fmt.Sprintf("value is not *[]*object:  %v", refValue.Elem().Type().Elem().Kind()))
	}
	refSlice := refValue.Elem()
	refItem := refSlice.Type().Elem()
	redisClient, err := p.getRedisConn()
	if err != nil {
		return err
	}
	argsM := make([]interface{}, len(args)+1)
	argsM[0] = cmd
	copy(argsM[1:], args)
	p.checkCtx()
	result, errDo := redisClient.Do(p.Ctx, argsM...).Slice()
	if errDo != nil {
		LogErrorw(LogNameRedis, "run redis command error",
			LogKNameCommonData, cmd,
			LogKNameCommonErr, errDo,
			LogKNameCommonFields, args,
		)
		return errDo
	}
	if result == nil {
		return nil
	}
	if len(result) > 0 {
		for i := 0; i < len(result); i++ {
			r := result[i]
			if r != nil {
				item := reflect.New(refItem)
				errorJson := JSONDecodeUseNumber([]byte(r.(string)), item.Interface())
				if errorJson != nil {
					LogErrorw(LogNameRedis, "redis command json Unmarshal error",
						LogKNameCommonData, cmd,
						LogKNameCommonErr, errorJson,
					)
					return errorJson
				}
				refSlice.Set(reflect.Append(refSlice, item.Elem()))
			} else {
				refSlice.Set(reflect.Append(refSlice, reflect.Zero(refItem)))
			}
		}
	}
	return nil
}

func (p *DaoRedis) doMGetGo(keys []string, value interface{}) error {
	var (
		args     []interface{}
		keysMap  sync.Map
		keysLen  int
		resultDo bool
		wg       sync.WaitGroup
	)
	keysLen = len(keys)
	if keysLen == 0 {
		return nil
	}
	refValue := reflect.ValueOf(value)
	if refValue.Kind() != reflect.Ptr || refValue.Elem().Kind() != reflect.Slice {
		return errors.New(fmt.Sprintf("value is not *[]*object:  %v", refValue.Elem().Type().Elem().Kind()))
	}
	refSlice := refValue.Elem()
	refItem := refSlice.Type().Elem()
	resultDo = true
	for _, v := range keys {
		args = append(args, p.getKey(v))
	}
	p.checkCtx()
	wg.Add(keysLen)
	for _, v := range args {
		getK := v
		go GoFuncOne(func() error {
			redisClient, err := p.getRedisConn()
			if err != nil {
				resultDo = false
			} else {
				rDo, errDo := redisClient.Do(p.Ctx, "GET", getK).Result()
				//获取 nil 也报错 忽略该错误
				if errors.Is(errDo, redis.Nil) {
					errDo = nil
				}
				keysMap.Store(getK, rDo)
				if errDo != nil {
					LogErrorw(LogNameRedis, "doMGetGo run redis command error",
						LogKNameCommonData, "GET",
						LogKNameCommonErr, errDo,
						LogKNameCommonFields, getK,
					)
					resultDo = false
				}
			}
			wg.Done()
			return nil
		})
	}
	wg.Wait()
	if !resultDo {
		return errors.New("doMGetGo one get error")
	}
	//整合结果
	for _, v := range args {
		r, ok := keysMap.Load(v)
		if ok && r != nil {
			item := reflect.New(refItem)
			errorJson := JSONDecodeUseNumber([]byte(r.(string)), item.Interface())
			if errorJson != nil {
				LogErrorw(LogNameRedis, "doMGetGo GET command result error",
					LogKNameCommonErr, errorJson,
				)
				return errorJson
			}
			refSlice.Set(reflect.Append(refSlice, item.Elem()))
		} else {
			refSlice.Set(reflect.Append(refSlice, reflect.Zero(refItem)))
		}
	}

	return nil
}

func (p *DaoRedis) doMGetStringMap(cmd string, args ...interface{}) (err error, data map[string]string) {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return err, nil
	}
	p.checkCtx()
	argsM := make([]interface{}, len(args)+1)
	argsM[0] = cmd
	copy(argsM[1:], args)
	doCmd := redis.NewStringStringMapCmd(p.Ctx, argsM...)
	_ = redisClient.Process(p.Ctx, doCmd)
	data, err = doCmd.Result()
	if err != nil {
		LogErrorw(LogNameRedis, "doMGetStringMap run redis command error",
			LogKNameCommonData, cmd,
			LogKNameCommonErr, err,
			LogKNameCommonFields, args,
		)
		return err, nil
	}
	return
}

func (p *DaoRedis) doMGetIntMap(cmd string, args ...interface{}) (err error, data map[string]int64) {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return err, nil
	}

	p.checkCtx()
	argsM := make([]interface{}, len(args)+1)
	argsM[0] = cmd
	copy(argsM[1:], args)
	doCmd := redis.NewStringIntMapCmd(p.Ctx, argsM...)
	_ = redisClient.Process(p.Ctx, doCmd)
	data, err = doCmd.Result()
	if err != nil {
		LogErrorw(LogNameRedis, "doMGetIntMap run redis command error",
			LogKNameCommonData, cmd,
			LogKNameCommonErr, err,
			LogKNameCommonFields, args,
		)
		return err, nil
	}
	return
}

func (p *DaoRedis) doIncr(cmd string, key string, value int, expire int, fields ...string) (int, bool) {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return 0, false
	}
	key = p.getKey(key)
	var data interface{}
	var errDo error
	p.checkCtx()
	if len(fields) == 0 {
		data, errDo = redisClient.Do(p.Ctx, cmd, key, value).Result()
	} else {
		field := fields[0]
		data, errDo = redisClient.Do(p.Ctx, cmd, key, field, value).Result()
	}
	if errDo != nil {
		LogErrorw(LogNameRedis, "doIncr run redis command error",
			LogKNameCommonData, cmd,
			LogKNameCommonErr, errDo,
			LogKNameCommonKey, key,
			LogKNameCommonFields, fields,
			LogKNameCommonValue, value,
		)
		return 0, false
	}
	count, result := data.(int64)
	if !result {
		LogErrorw(LogNameRedis, "doIncr get command result error",
			LogKNameCommonCmd, cmd,
			LogKNameCommonData, data,
			LogKNameCommonDataType, reflect.TypeOf(data),
		)
		return 0, false
	}
	if expire == 0 {
		redisConfig := ConfigCacheGetRedisBaseWithConn(p.Persistent)
		expire = redisConfig.Expire
	}
	//set expire
	if expire > 0 {
		_, errExpire := redisClient.Do(p.Ctx, "EXPIRE", key, expire).Result()
		if errExpire != nil {
			LogErrorw(LogNameRedis, "run redis EXPIRE command error",
				LogKNameCommonErr, errExpire,
				LogKNameCommonKey, key,
				LogKNameCommonTime, expire,
			)
		}
	}
	return int(count), true
}

func (p *DaoRedis) doDel(cmd string, args ...interface{}) error {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return err
	}
	argsM := make([]interface{}, len(args)+1)
	argsM[0] = cmd
	copy(argsM[1:], args)
	p.checkCtx()
	errDo := redisClient.Do(p.Ctx, argsM...).Err()
	if errDo != nil {
		LogErrorw(LogNameRedis, "doDel run redis command error",
			LogKNameCommonCmd, cmd,
			LogKNameCommonErr, errDo,
			LogKNameCommonData, args,
		)
	}
	return errDo
}

/*基础结束*/

func (p *DaoRedis) Set(key string, value interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doSet("SET", key, value, 0)
	SpanErrorFast(span, err)
	if err != nil {
		return false
	}
	return true
}

func (p *DaoRedis) SetE(key string, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doSet("SET", key, value, 0)
	SpanErrorFast(span, err)
	return err
}

// MSet mset
func (p *DaoRedis) MSet(datas map[string]interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doMSet("MSET", "", datas)
	SpanErrorFast(span, err)
	if err != nil {
		return false
	}
	return true
}

// SetEx setex
func (p *DaoRedis) SetEx(key string, value interface{}, expire int) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doSet("SET", key, value, expire)
	SpanErrorFast(span, err)
	if err != nil {
		return false
	}
	return true
}

// SetEx setex
func (p *DaoRedis) SetExE(key string, value interface{}, expire int) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doSet("SET", key, value, expire)
	SpanErrorFast(span, err)
	return err
}

// Expire expire
func (p *DaoRedis) Expire(key string, expire int) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return false
	}
	key = p.getKey(key)
	p.checkCtx()
	err = redisClient.Do(p.Ctx, "EXPIRE", key, expire).Err()
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameRedis, "Expire run redis EXPIRE command error",
			"key", key,
			"err", err,
			"expire", expire,
		)
		return false
	}
	return true
}

func (p *DaoRedis) Get(key string, data interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	result, err := p.doGet("GET", key, data)
	SpanErrorFast(span, err)
	if err == nil && result {
		return true
	}
	return false
}
func (p *DaoRedis) GetE(key string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doGet("GET", key, data)
	SpanErrorFast(span, err)
	return err
}

// 返回 1. key是否存在 2. error
func (p *DaoRedis) GetRaw(key string, data interface{}) (b bool, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	b, err = p.doGet("GET", key, data)
	SpanErrorFast(span, err)
	return
}

func (p *DaoRedis) MGet(keys []string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var args []interface{}
	for _, v := range keys {
		args = append(args, p.getKey(v))
	}
	err := p.doMGet("MGET", args, data)
	SpanErrorFast(span, err)
	return err
}

// 封装mget通过go并发get
func (p *DaoRedis) MGetGo(keys []string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	err := p.doMGetGo(keys, data)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedis) Incr(key string) (int, bool) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	return p.doIncr("INCRBY", key, 1, 0)
}

func (p *DaoRedis) IncrBy(key string, value int) (int, bool) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	return p.doIncr("INCRBY", key, value, 0)
}
func (p *DaoRedis) SetNX(key string, value interface{}) (int64, bool) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	return p.doSetNX("SETNX", key, value, 0)
}

func (p *DaoRedis) SetNXNoExpire(key string, value interface{}) (int64, bool) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	return p.doSetNX("SETNX", key, value, -1)
}

func (p *DaoRedis) Del(key string) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key = p.getKey(key)
	err := p.doDel("DEL", key)
	SpanErrorFast(span, err)
	if err != nil {
		return false
	}
	return true
}

func (p *DaoRedis) MDel(key ...string) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var keys []interface{}
	for _, v := range key {
		keys = append(keys, p.getKey(v))
	}
	err := p.doDel("DEL", keys...)
	SpanErrorFast(span, err)
	if err != nil {
		return false
	}
	return true
}

func (p *DaoRedis) Exists(key string) (bool, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return false, err
	}
	key = p.getKey(key)
	p.checkCtx()
	data, err := redisClient.Do(p.Ctx, "EXISTS", key).Result()
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameRedis, "Exists run redis EXISTS command error",
			"key", key,
			"err", err,
		)
		return false, err
	}
	count, result := data.(int64)
	if !result {
		err := errors.New(fmt.Sprintf("get EXISTS command result failed:%v ,is %v", data, reflect.TypeOf(data)))
		LogErrorw(LogNameRedis, "get EXISTS command result error",
			"data", data,
			"data_type", reflect.TypeOf(data),
		)
		return false, err
	}
	if count == 1 {
		return true, nil
	}
	return false, nil
}

// hash start
func (p *DaoRedis) HIncrby(key string, field string, value int) (int, bool) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	return p.doIncr("HINCRBY", key, value, 0, field)
}

func (p *DaoRedis) HGet(key string, field string, value interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	result, err := p.doGet("HGET", key, value, field)
	SpanErrorFast(span, err)
	if err == nil && result {
		return true
	}
	return false
}

// HGetE 返回error
func (p *DaoRedis) HGetE(key string, field string, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doGet("HGET", key, value, field)
	SpanErrorFast(span, err)
	return err
}

// HGetRaw 返回 1. key是否存在 2. error
func (p *DaoRedis) HGetRaw(key string, field string, value interface{}) (b bool, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	b, err = p.doGet("HGET", key, value, field)
	SpanErrorFast(span, err)
	return
}

func (p *DaoRedis) HMGet(key string, fields []interface{}, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var args []interface{}
	args = append(args, p.getKey(key))
	for _, v := range fields {
		args = append(args, v)
	}
	err := p.doMGet("HMGET", args, data)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedis) HSet(key string, field string, value interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doSet("HSET", key, value, 0, field)
	SpanErrorFast(span, err)
	if err != nil {
		return false
	}
	return true
}
func (p *DaoRedis) HSetNX(key string, field string, value interface{}) (int64, bool) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	return p.doSetNX("HSETNX", key, value, 0, field)
}

// HMSet value是filed:data
func (p *DaoRedis) HMSet(key string, value map[string]interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doMSet("HMSet", key, value)
	SpanErrorFast(span, err)
	if err != nil {
		return false
	}
	return true
}

// HMSetE value是filed:data
func (p *DaoRedis) HMSetE(key string, value map[string]interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doMSet("HMSet", key, value)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedis) HLen(key string, data *int) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return false
	}
	key = p.getKey(key)
	p.checkCtx()
	resultData, errDo := redisClient.Do(p.Ctx, "HLEN", key).Result()
	SpanErrorFast(span, errDo)
	if errDo != nil {
		LogErrorw(LogNameRedis, "HLen run redis HLEN command error",
			"key", key,
			"err", errDo,
		)
		return false
	}
	length, b := resultData.(int64)
	if !b {
		LogErrorw(LogNameRedis, "HLen redis data convert to int64 error",
			"ret", resultData,
		)
	}
	*data = int(length)
	return b
}

func (p *DaoRedis) HDel(key string, data ...interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var args []interface{}
	key = p.getKey(key)
	args = append(args, key)
	for _, item := range data {
		args = append(args, item)
	}
	err := p.doDel("HDEL", args...)
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameRedis, "HDel run redis HDEL command error",
			"key", key,
			"err", err,
			"data", data,
		)
		return false
	}
	return true
}

func (p *DaoRedis) HExists(key string, field string) (bool, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return false, err
	}
	key = p.getKey(key)
	p.checkCtx()
	data, err := redisClient.Do(p.Ctx, "HEXISTS", key, field).Result()
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameRedis, "HExists run redis HEXISTS command error",
			"key", key,
			"err", err,
		)
		return false, err
	}
	count, result := data.(int64)
	if !result {
		err := errors.New(fmt.Sprintf("get HEXISTS command result failed:%v ,is %v", data, reflect.TypeOf(data)))
		LogErrorw(LogNameRedis, "HExists get HEXISTS command result error",
			"data", data,
			"data_type", reflect.TypeOf(data),
		)
		return false, err
	}
	if count == 1 {
		return true, nil
	}
	return false, nil
}

// hash end

// sorted set start
func (p *DaoRedis) ZAdd(key string, score interface{}, data interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return false
	}
	key = p.getKey(key)
	p.checkCtx()
	errDo := redisClient.Do(p.Ctx, "ZADD", key, score, data).Err()
	SpanErrorFast(span, errDo)
	if errDo != nil {
		LogErrorw(LogNameRedis, "ZAdd run redis ZAdd command error",
			"key", key,
			"err", errDo,
			"score", score,
			"data", data,
		)
		return false
	}
	return true
}

func (p *DaoRedis) ZCard(key string) (data int, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return
	}
	key = p.getKey(key)
	var reply interface{}
	p.checkCtx()
	reply, err = redisClient.Do(p.Ctx, "ZCARD", key).Result()
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameRedis, "ZCard run redis ZCard command error",
			"key", key,
			"err", err,
		)
		return
	}
	if v, ok := reply.(int64); ok {
		data = int(v)
		return
	} else {
		LogErrorw(LogNameRedis, "ZCard get replay is not int64 error",
			"key", key,
			"reply", reply,
		)
		err = errors.New(fmt.Sprintf("ZCard get replay is not int64:%v", reply))
		return
	}
}

func (p *DaoRedis) ZCount(key string, min, max int) (data int, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return
	}
	key = p.getKey(key)
	var reply interface{}
	p.checkCtx()
	reply, err = redisClient.Do(p.Ctx, "ZCOUNT", key, min, max).Result()
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameRedis, "ZCount run redis ZCOUNT command error",
			"key", key,
			"min", min,
			"max", max,
		)
		return
	}
	if v, ok := reply.(int64); ok {
		data = int(v)
		return
	} else {
		LogErrorw(LogNameRedis, "ZCount get replay is not int64 error",
			"key", key,
			"reply", reply,
		)
		err = errors.New(fmt.Sprintf("ZCount get replay is not int64:%v", reply))
		return
	}
}

func (p *DaoRedis) ZIncrBy(key string, increment int, member interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return false
	}
	key = p.getKey(key)
	p.checkCtx()
	errDo := redisClient.Do(p.Ctx, "ZINCRBY", key, increment, member).Err()
	SpanErrorFast(span, errDo)
	if errDo != nil {
		LogErrorw(LogNameRedis, "ZIncrBy run redis ZINCRBY command error",
			"key", key,
			"err", errDo,
			"increment", increment,
			"member", member,
		)
		return false
	}
	return true
}

// sorted set start
func (p *DaoRedis) ZAddM(key string, value map[string]interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doMSet("ZADD", key, value)
	SpanErrorFast(span, err)
	if err != nil {
		return false
	}
	return true
}

func (p *DaoRedis) ZGetByScore(key string, sort bool, start int, end int, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var cmd string
	if sort {
		cmd = "ZRANGEBYSCORE"
	} else {
		cmd = "ZREVRANGEBYSCORE"
	}
	var args []interface{}
	args = append(args, p.getKey(key))
	args = append(args, start)
	args = append(args, end)
	err := p.doMGet(cmd, args, value)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedis) ZGet(key string, sort bool, start int, end int, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)

	var cmd string
	if sort {
		cmd = "ZRANGE"
	} else {
		cmd = "ZREVRANGE"
	}

	var args []interface{}
	args = append(args, p.getKey(key))
	args = append(args, start)
	args = append(args, end)

	err := p.doMGet(cmd, args, value)
	SpanErrorFast(span, err)

	return err
}

func (p *DaoRedis) ZGetWithScores(key string, sort bool, start int, end int) (err error, data map[string]string) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)

	var cmd string
	if sort {
		cmd = "ZRANGE"
	} else {
		cmd = "ZREVRANGE"
	}

	var args []interface{}
	args = append(args, p.getKey(key))
	args = append(args, start)
	args = append(args, end)
	args = append(args, "WITHSCORES")

	err, data = p.doMGetStringMap(cmd, args...)
	SpanErrorFast(span, err)

	return
}

func (p *DaoRedis) ZRank(key string, member string, sort bool) (bool, int) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)

	var cmd string
	if sort {
		cmd = "ZRANK"
	} else {
		cmd = "ZREVRANK"
	}

	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)

	if err != nil {
		return false, 0
	}

	key = p.getKey(key)
	p.checkCtx()
	result, errDo := redisClient.Do(p.Ctx, cmd, key, member).Result()
	SpanErrorFast(span, errDo)

	if errDo != nil {
		LogErrorw(LogNameRedis, "ZRank run redis command error",
			"key", key,
			"err", errDo,
			"key", key,
			"member", member,
		)
		return false, 0
	}
	if v, ok := result.(int64); ok {
		return true, int(v)
	}
	return false, 0
}

func (p *DaoRedis) ZScore(key string, member string, value interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)

	var cmd string
	cmd = "ZSCORE"

	result, err := p.doGet(cmd, key, value, member)
	SpanErrorFast(span, err)
	if err == nil && result {
		return true
	}

	return false
}

func (p *DaoRedis) ZRevRange(key string, start int, end int, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	err := p.ZGet(key, false, start, end, value)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedis) ZRem(key string, data ...interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)

	var args []interface{}

	key = p.getKey(key)
	args = append(args, key)

	for _, item := range data {
		args = append(args, item)
	}

	err := p.doDel("ZREM", args...)
	SpanErrorFast(span, err)

	if err != nil {
		return false
	}
	return true
}

//list start

func (p *DaoRedis) LRange(start int, end int, value interface{}) (err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key := ""
	key = p.getKey(key)
	var args []interface{}
	args = append(args, key)
	args = append(args, start)
	args = append(args, end)
	err = p.doMGet("LRANGE", args, value)
	SpanErrorFast(span, err)
	return
}

func (p *DaoRedis) LLen() (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	cmd := "LLEN"
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return 0, err
	}
	key := ""
	key = p.getKey(key)
	var result interface{}
	var errDo error
	var args []interface{}
	args = append(args, key)
	p.checkCtx()
	result, errDo = redisClient.Do(p.Ctx, cmd, key).Result()
	SpanErrorFast(span, errDo)
	if errDo != nil {
		LogErrorw(LogNameRedis, "LLen run redis command error",
			"key", key,
			"err", errDo,
			"key", key,
		)
		return 0, errDo
	}
	if result == nil {
		return 0, nil
	}
	num, ok := result.(int64)
	if !ok {
		return 0, errors.New("result to int64 failed")
	}
	return num, nil
}

func (p *DaoRedis) LREM(count int, data interface{}) int {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return 0
	}
	key := ""
	key = p.getKey(key)
	p.checkCtx()
	result, errDo := redisClient.Do(p.Ctx, "LREM", key, count, data).Result()
	SpanErrorFast(span, errDo)
	if errDo != nil {
		LogErrorw(LogNameRedis, "LREM run redis command error",
			"key", key,
			"err", errDo,
			"key", key,
			"count", count,
			"data", data,
		)
		return 0
	}
	countRem, ok := result.(int)
	if !ok {
		LogErrorw(LogNameRedis, "LREM redis data convert to int error",
			"key", key,
			"err", errDo,
			"key", key,
			"count", count,
			"data", data,
			"result", result,
		)
		return 0
	}
	return countRem
}

func (p *DaoRedis) LTRIM(start int, end int) (err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return
	}
	key := ""
	key = p.getKey(key)
	p.checkCtx()
	err = redisClient.Do(p.Ctx, "LTRIM", key, start, end).Err()
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameRedis, "LTRIM redis data convert to int error",
			"key", key,
			"err", err,
			"key", key,
			"start", start,
			"end", end,
		)
		return
	}
	return
}

func (p *DaoRedis) RPush(value interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	return p.Push(value, false)
}

func (p *DaoRedis) LPush(value interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	return p.Push(value, true)
}

func (p *DaoRedis) Push(value interface{}, isLeft bool) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var cmd string
	if isLeft {
		cmd = "LPUSH"
	} else {
		cmd = "RPUSH"
	}
	key := ""
	_, err := p.doSet(cmd, key, value, -1)
	if err != nil {
		return false
	}
	return true
}

func (p *DaoRedis) RPop(value interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	return p.Pop(value, false)
}

func (p *DaoRedis) LPop(value interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	return p.Pop(value, true)
}

func (p *DaoRedis) Pop(value interface{}, isLeft bool) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var cmd string
	if isLeft {
		cmd = "LPOP"
	} else {
		cmd = "RPOP"
	}
	key := ""
	_, err := p.doGet(cmd, key, value)
	if err == nil {
		return true
	} else {
		return false
	}
}

//list end

//pipeline start

func (p *DaoRedis) PipelineHGet(key []string, fields []interface{}, data []interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var args [][]interface{}

	for k, v := range key {
		var arg []interface{}
		arg = append(arg, p.getKey(v))
		arg = append(arg, fields[k])
		args = append(args, arg)
	}

	err := p.pipeDoGet("HGET", args, data)
	SpanErrorFast(span, err)

	return err
}

func (p *DaoRedis) pipeDoGet(cmd string, args [][]interface{}, value []interface{}) error {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return err
	}
	p.checkCtx()
	pi := redisClient.Pipeline()
	//写入命令
	var cmds []*redis.Cmd
	for _, v := range args {
		argsM := make([]interface{}, len(v)+1)
		argsM[0] = cmd
		copy(argsM[1:], v)
		cmds = append(cmds, pi.Do(p.Ctx, argsM...))
	}
	//执行
	if _, err := pi.Exec(p.Ctx); err != nil {
		LogErrorw(LogNameRedis, "pipeDoGet Flush returned error",
			"cmd", cmd,
			"err", err,
		)
		return err
	}
	//获取结果
	for k, v := range args {
		result, err := cmds[k].Result()
		if err != nil {
			LogErrorw(LogNameRedis, "pipeDoGet Receive returned error",
				"cmd", cmd,
				"err", err,
				LogKNameCommonValue, v,
			)
			return err
		}
		if result == nil {
			value[k] = nil
			continue
		}
		if reflect.TypeOf(result).Kind() == reflect.Slice {

			byteResult := result.([]byte)
			strResult := string(byteResult)

			if strResult == "[]" {
				value[k] = nil
				continue
			}
		}

		errorJson := JSONDecodeUseNumber(result.([]byte), value[k])

		if errorJson != nil {
			LogErrorw(LogNameRedis, "pipeDoGet get command result error",
				"cmd", cmd,
				"err", errorJson,
			)
			return errorJson
		}
	}

	return nil
}

//pipeline end

// Set集合Start
func (p *DaoRedis) SAdd(key string, argPs []interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)

	if err != nil {
		return false
	}

	args := make([]interface{}, len(argPs)+2)
	args[0] = "SADD"
	args[1] = p.getKey(key)
	copy(args[2:], argPs)
	p.checkCtx()
	errDo := redisClient.Do(p.Ctx, args...).Err()
	SpanErrorFast(span, errDo)

	if errDo != nil {
		LogErrorw(LogNameRedis, "SAdd run redis SADD command error",
			"key", key,
			"args", args,
			"err", errDo,
		)
		return false
	}
	return true
}

func (p *DaoRedis) SIsMember(key string, arg interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)

	if err != nil {
		return false
	}

	key = p.getKey(key)
	p.checkCtx()
	reply, errDo := redisClient.Do(p.Ctx, "SISMEMBER", key, arg).Result()
	SpanErrorFast(span, errDo)

	if errDo != nil {
		LogErrorw(LogNameRedis, "SIsMember run redis SISMEMBER command error",
			"key", key,
			"args", arg,
			"err", errDo,
		)
		return false
	}
	if code, ok := reply.(int64); ok && code == int64(1) {
		return true
	}
	return false
}

func (p *DaoRedis) SCard(key string) int64 {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return 0
	}
	key = p.getKey(key)
	p.checkCtx()
	reply, errDo := redisClient.Do(p.Ctx, "SCARD", key).Result()
	SpanErrorFast(span, errDo)
	if errDo != nil {
		LogErrorw(LogNameRedis, "SCARD run redis SCARD command error",
			"key", key,
			"err", errDo,
		)
		return 0
	}
	return reply.(int64)
}

func (p *DaoRedis) SRem(key string, argPs []interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)

	if err != nil {
		return false
	}

	args := make([]interface{}, len(argPs)+2)
	args[0] = "SREM"
	args[1] = p.getKey(key)
	copy(args[2:], argPs)
	p.checkCtx()
	errDo := redisClient.Do(p.Ctx, args...).Err()
	SpanErrorFast(span, errDo)

	if errDo != nil {
		LogErrorw(LogNameRedis, "SRem run redis SREM command error",
			"key", key,
			"args", args,
			"err", errDo,
		)
		return false
	}
	return true
}

func (p *DaoRedis) SPop(key string, value interface{}) bool {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doGet("SPOP", key, value)
	SpanErrorFast(span, err)
	if err == nil {
		return true
	} else {
		return false
	}
}

func (p *DaoRedis) SMembers(key string, value interface{}) (err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var args []interface{}
	args = append(args, p.getKey(key))
	err = p.doMGet("SMEMBERS", args, value)
	SpanErrorFast(span, err)
	return
}

func (p *DaoRedis) HGetAll(key string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var args []interface{}

	args = append(args, p.getKey(key))

	err := p.doMGet("HGETALL", args, data)
	SpanErrorFast(span, err)

	return err
}

func (p *DaoRedis) HGetAllStringMap(key string) (err error, data map[string]string) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	args := p.getKey(key)
	err, data = p.doMGetStringMap("HGETALL", args)
	SpanErrorFast(span, err)
	return
}

func (p *DaoRedis) HGetAllIntMap(key string) (err error, data map[string]int64) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	args := p.getKey(key)
	err, data = p.doMGetIntMap("HGETALL", args)
	SpanErrorFast(span, err)
	return
}

// GetPTtl：获取key的过期时间，单位为毫秒
// 如果key不存在返回-2
// 如果key存在，但是没有设置过期时间，返回-1
func (p *DaoRedis) GetPTtl(key string) (ttl int64, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	ttl, err = p.doGetTtl("PTTL", key)
	SpanErrorFast(span, err)
	return
}

// 如果key不存在返回-2
// 如果key存在，但是没有设置过期时间，返回-1

// GetTtl 获取key的过期时间，单位为秒
func (p *DaoRedis) GetTtl(key string) (ttl int64, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	ttl, err = p.doGetTtl("TTL", key)
	SpanErrorFast(span, err)
	return
}

func (p *DaoRedis) doGetTtl(cmd string, key string) (ttl int64, err error) {
	args := p.getKey(key)
	redisClient, err := p.getRedisConn()
	if err != nil {
		return 0, err
	}
	p.checkCtx()
	ttl, err = redisClient.Do(p.Ctx, cmd, args).Int64()
	if err != nil {
		LogErrorw(LogNameRedis, "doGetTtl run redis command error",
			"cmd", cmd,
			"err", err,
			"args", args,
		)
		return 0, err
	}
	return
}
