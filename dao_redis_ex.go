package gozen

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"strconv"
	"strings"
	"sync"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

type DaoRedisEx struct {
	KeyName          string
	Persistent       bool            // 持久化key
	ExpireSecond     int             // 默认过期时间，单实例有效
	tempExpireSecond int             // 临时默认过期时间，单条命令有效
	tempClusterKey   bool            // 临时默认获取集群key，大括号包裹用于mset mget，单条命令有效
	Ctx              context.Context // 必须定义 否则报错
}

const (
	redisDefaultExpireSecond = 600 //最终兜底缓存时间 十分钟
)

/**


redis 操作

1、默认expire为0，使用兜底缓存时间
2、可指定缓存时间，包含永久 -1
3、可设置临时缓存时间


expire 设置 -1 保证永久生效
expire 设置 0 相当于删除
expire 设置 正值 一定有效期


原则上保证每个缓存均有失效时间，在一定业务内失效即可

永久缓存的key不存在


*/

type OpOptionEx func(*DaoRedisEx)

// WithExpire 设置超时时间 - 为了兼容之前的命令写法 无expire时 加expire
func WithExpire(expire int) OpOptionEx {
	return func(p *DaoRedisEx) { p.tempExpireSecond = expire }
}

func WithClusterKey() OpOptionEx {
	return func(p *DaoRedisEx) { p.tempClusterKey = true }
}

// applyOpts 应用扩展属性
func (p *DaoRedisEx) applyOpts(opts []OpOptionEx) {
	for _, opt := range opts {
		opt(p)
	}
}

// resetTempExpireSecond 重置临时过期时间
func (p *DaoRedisEx) resetTempExpireSecond() { //expire 设置完后 reset 因为expire 仅一次 可以放在getExpire里面
	p.tempExpireSecond = 0
}

func (p *DaoRedisEx) resetTempClusterKey() { //key 全部设置完后 reset 可以放在do执行前
	p.tempClusterKey = false
}

func (p *DaoRedisEx) checkCtx() {
	if p.Ctx == nil {
		p.Ctx = context.Background()
	}
	p.resetTempClusterKey()
}

// getExpire 获取过期时间
func (p *DaoRedisEx) getExpire(expire int) int {
	/**

	优先级

	1、临时时间
	2、指定时间
	3、实例时间
	4、兜底时间

	*/

	defer p.resetTempExpireSecond()

	var expireSecond int
	if p.tempExpireSecond != 0 {
		expireSecond = p.tempExpireSecond
	} else if expire != 0 {
		expireSecond = expire
	} else if p.ExpireSecond != 0 {
		expireSecond = p.ExpireSecond
	} else {
		redisConfig := ConfigCacheGetRedisBaseWithConn(p.Persistent)
		expireSecond = redisConfig.Expire
	}

	if expireSecond == 0 {
		expireSecond = redisDefaultExpireSecond
	}

	if expireSecond < 0 { //设置永久
		expireSecond = -1
	}
	return expireSecond
}

func (p *DaoRedisEx) GetExpire(expire int) int {
	return p.getExpire(expire)
}

// 获取redis连接
func (p *DaoRedisEx) getRedisConn() (c redis.UniversalClient, err error) {
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
func (p *DaoRedisEx) GetRedisConn() (c redis.UniversalClient, err error) {
	return p.getRedisConn()
}

func (p *DaoRedisEx) getKey(key string) string {
	redisConfig := ConfigCacheGetRedisBaseWithConn(p.Persistent)
	prefixRedis := redisConfig.Prefix
	key = strings.Trim(key, " ")
	if key == "" {
		return fmt.Sprintf("%s:%s", prefixRedis, p.KeyName)
	}
	if p.tempClusterKey {
		return fmt.Sprintf("{%s:%s}:%s", prefixRedis, p.KeyName, key)
	} else {
		return fmt.Sprintf("%s:%s:%s", prefixRedis, p.KeyName, key)
	}
}

func (p *DaoRedisEx) GetKey(key string) string {
	return p.getKey(key)
}

// 最终执行do
func (p *DaoRedisEx) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return nil, err
	}
	argsM := make([]interface{}, len(args)+1)
	argsM[0] = commandName
	copy(argsM[1:], args)
	p.checkCtx()
	reply, err = redisClient.Do(p.Ctx, argsM...).Result()
	if errors.Is(err, redis.Nil) { //忽略key不存在的情况 返回指定类型空值
		err = nil
	}
	return
}

// doReturnKeyNil 执行并返回 key 不存在的error
func (p *DaoRedisEx) doReturnKeyNil(commandName string, args ...interface{}) (reply interface{}, err error) {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return nil, err
	}
	argsM := make([]interface{}, len(args)+1)
	argsM[0] = commandName
	copy(argsM[1:], args)
	p.checkCtx()
	reply, err = redisClient.Do(p.Ctx, argsM...).Result()
	return
}

func (p *DaoRedisEx) doExpire(key []string, expire int) {
	if expire < 0 { //设置永久
		return
	}
	for _, v := range key {
		_, errExpire := p.do("EXPIRE", v, expire)
		if errExpire != nil {
			UtilLogErrorf("run redis EXPIRE command failed: error:%s,key:%s,time:%d", errExpire.Error(), key, expire)
		}
	}
}

func (p *DaoRedisEx) doSet(cmd string, key string, value interface{}, expire int, fields ...string) (interface{}, error) {
	key = p.getKey(key)
	expire = p.getExpire(expire)
	data, errJson := json.Marshal(value)
	if errJson != nil {
		UtilLogErrorf("redis %s marshal data to json:%s", cmd, errJson.Error())
		return nil, errJson
	}
	var reply interface{}
	var errDo error
	if len(fields) == 0 {
		if strings.ToUpper(cmd) == "SET" && expire >= 0 { //小于0 默认永久
			reply, errDo = p.do(cmd, key, data, "ex", expire)
		} else {
			reply, errDo = p.do(cmd, key, data)
		}
	} else {
		field := fields[0]
		reply, errDo = p.do(cmd, key, field, data)
	}
	if errDo != nil {
		UtilLogErrorf("run redis command %s failed:error:%s,key:%s,fields:%v,data:%v", cmd, errDo.Error(), key, fields, value)
		return nil, errDo
	}
	//set expire
	if strings.ToUpper(cmd) != "SET" && expire >= 0 {
		p.doExpire([]string{key}, expire)
	}
	return reply, errDo
}

func (p *DaoRedisEx) doSetNX(cmd string, key string, value interface{}, expire int, field ...string) (num int64, err error) {
	var (
		reply interface{}
		ok    bool
	)
	reply, err = p.doSet(cmd, key, value, expire, field...)
	if err != nil {
		return
	}
	num, ok = reply.(int64)
	if !ok {
		msg := fmt.Sprintf("HSetNX reply to int failed,key:%v,field:%v", key, field)
		UtilLogErrorf(msg)
		err = errors.New(msg)
		return
	}
	return
}

func (p *DaoRedisEx) doMSet(cmd string, key string, value map[string]interface{}) (interface{}, error) {
	var args []interface{}
	var keys []string
	if key != "" { //hmset
		key = p.getKey(key)
		args = append(args, key)
		keys = append(keys, key)
	}
	expire := p.getExpire(0)
	for k, v := range value {
		data, errJson := json.Marshal(v)
		if errJson != nil {
			UtilLogErrorf("redis %s marshal data: %v to json:%s", cmd, v, errJson.Error())
			return nil, errJson
		}
		if key == "" { //mset
			keys = append(keys, p.getKey(k))
			args = append(args, p.getKey(k), data)
		} else { //hmset
			args = append(args, k, data)
		}
	}
	var reply interface{}
	var errDo error
	reply, errDo = p.do(cmd, args...)
	if errDo != nil {
		UtilLogErrorf("run redis command %s failed:error:%s,key:%s,value:%v", cmd, errDo.Error(), key, value)
		return nil, errDo
	}
	//expire
	p.doExpire(keys, expire)
	return reply, errDo
}

func (p *DaoRedisEx) doGet(cmd string, key string, value interface{}, fields ...string) (bool, error) {
	key = p.getKey(key)
	var result interface{}
	var errDo error
	var args []interface{}
	args = append(args, key)
	for _, f := range fields {
		args = append(args, f)
	}
	result, errDo = p.doReturnKeyNil(cmd, args...)
	if errors.Is(errDo, redis.Nil) {
		value = nil
		return false, nil
	}
	if errDo != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,key:%s,fields:%v", cmd, errDo.Error(), key, fields)
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
		UtilLogErrorf("get %s command result failed:%s", cmd, errorJson.Error())
		return false, errorJson
	}
	return true, nil
}

func (p *DaoRedisEx) doMGet(cmd string, args []interface{}, value interface{}) error {
	refValue := reflect.ValueOf(value)
	if refValue.Kind() != reflect.Ptr || refValue.Elem().Kind() != reflect.Slice {
		return errors.New(fmt.Sprintf("value is not *[]object:  %v", refValue.Elem().Type().Elem().Kind()))
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
		UtilLogErrorf("run redis %s command failed: error:%s,args:%v", cmd, errDo.Error(), args)
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
					UtilLogErrorf("%s command result failed:%s", cmd, errorJson.Error())
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

func (p *DaoRedisEx) doMGetGo(keys []string, value interface{}) error {
	var (
		args     []interface{}
		keysMap  sync.Map
		keysLen  int
		rDo      interface{}
		errDo    error
		resultDo bool
		wg       sync.WaitGroup
	)
	keysLen = len(keys)
	if keysLen == 0 {
		return nil
	}
	refValue := reflect.ValueOf(value)
	if refValue.Kind() != reflect.Ptr || refValue.Elem().Kind() != reflect.Slice {
		return errors.New(fmt.Sprintf("value is not *[]object:  %v", refValue.Elem().Type().Elem().Kind()))
	}
	refSlice := refValue.Elem()
	refItem := refSlice.Type().Elem()
	resultDo = true
	for _, v := range keys {
		args = append(args, p.getKey(v))
	}
	wg.Add(keysLen)
	for _, v := range args {
		getK := v
		go GoFuncOne(func() error {
			rDo, errDo = p.do("GET", getK)
			//获取 nil 也报错 忽略该错误
			if errors.Is(redis.Nil, errDo) {
				errDo = nil
			}
			if errDo != nil {
				UtilLogErrorf("run redis GET command failed: error:%s,args:%v", errDo.Error(), getK)
				resultDo = false
			} else {
				keysMap.Store(getK, rDo)
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
				UtilLogErrorf("GET command result failed:%s", errorJson.Error())
				return errorJson
			}
			refSlice.Set(reflect.Append(refSlice, item.Elem()))
		} else {
			refSlice.Set(reflect.Append(refSlice, reflect.Zero(refItem)))
		}
	}
	return nil
}

func (p *DaoRedisEx) doGetSlice(cmd string, args ...interface{}) (err error, data []interface{}) {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return err, nil
	}
	p.checkCtx()
	argsM := make([]interface{}, len(args)+1)
	argsM[0] = cmd
	copy(argsM[1:], args)
	cmdDo := redis.NewSliceCmd(p.Ctx, argsM...)
	_ = redisClient.Process(p.Ctx, cmdDo)
	data, err = cmdDo.Result()
	if err != nil {
		UtilLogErrorf("run redis %s command failed: error:%v, args:%v", cmd, err, args)
		return err, nil
	}
	return
}

func (p *DaoRedisEx) doGetInterface(cmd string, args ...interface{}) (err error, data interface{}) {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return err, nil
	}
	p.checkCtx()
	argsM := make([]interface{}, len(args)+1)
	argsM[0] = cmd
	copy(argsM[1:], args)
	cmdDo := redis.NewCmd(p.Ctx, argsM...)
	_ = redisClient.Process(p.Ctx, cmdDo)
	data, err = cmdDo.Result()
	if err != nil {
		UtilLogErrorf("run redis %s command failed: error:%v, args:%v", cmd, err, args)
		return err, nil
	}
	return
}

func (p *DaoRedisEx) doMGetStringMap(cmd string, args ...interface{}) (err error, data map[string]string) {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return err, nil
	}
	p.checkCtx()
	argsM := make([]interface{}, len(args)+1)
	argsM[0] = cmd
	copy(argsM[1:], args)
	cmdDo := redis.NewStringStringMapCmd(p.Ctx, argsM...)
	_ = redisClient.Process(p.Ctx, cmdDo)
	data, err = cmdDo.Result()
	if err != nil {
		UtilLogErrorf("run redis %s command failed: error:%v, args:%v", cmd, err, args)
		return err, nil
	}
	return
}

func (p *DaoRedisEx) doMGetIntMap(cmd string, args ...interface{}) (err error, data map[string]int64) {
	redisClient, err := p.getRedisConn()
	if err != nil {
		return err, nil
	}
	p.checkCtx()
	argsM := make([]interface{}, len(args)+1)
	argsM[0] = cmd
	copy(argsM[1:], args)
	cmdDo := redis.NewStringIntMapCmd(p.Ctx, argsM...)
	_ = redisClient.Process(p.Ctx, cmdDo)
	data, err = cmdDo.Result()
	if err != nil {
		UtilLogErrorf("run redis %s command failed: error:%v, args:%v", cmd, err, args)
		return err, nil
	}
	return
}

func (p *DaoRedisEx) doIncr(cmd string, key string, value int, expire int, fields ...string) (num int64, err error) {
	var (
		data interface{}
		ok   bool
	)
	expire = p.getExpire(expire)
	key = p.getKey(key)
	if len(fields) == 0 {
		data, err = p.do(cmd, key, value)
	} else {
		field := fields[0]
		data, err = p.do(cmd, key, field, value)
	}
	if err != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,key:%s,fields:%v,value:%d", cmd, err.Error(), key, fields, value)
		return
	}
	num, ok = data.(int64)
	if !ok {
		msg := fmt.Sprintf("get %s command result failed:%v ,is %v", cmd, data, reflect.TypeOf(data))
		UtilLogErrorf(msg)
		err = errors.New(msg)
		return
	}
	p.doExpire([]string{key}, expire)
	return
}

func (p *DaoRedisEx) doIncrNoExpire(cmd string, key string, value int, fields ...string) (num int64, err error) {
	var (
		data interface{}
		ok   bool
	)
	key = p.getKey(key)
	if len(fields) == 0 {
		data, err = p.do(cmd, key, value)
	} else {
		field := fields[0]
		data, err = p.do(cmd, key, field, value)
	}
	if err != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,key:%s,fields:%v,value:%d", cmd, err.Error(), key, fields, value)
		return
	}
	num, ok = data.(int64)
	if !ok {
		msg := fmt.Sprintf("get %s command result failed:%v ,is %v", cmd, data, reflect.TypeOf(data))
		UtilLogErrorf(msg)
		err = errors.New(msg)
		return
	}
	return
}

// 比较成功就添加
func (p *DaoRedisEx) compareWithAdd(cmd string, key string, addValue int, compareValue int) (num int64, err error) {
	key = p.getKey(key)
	expire := p.getExpire(0)
	redisClient, err := p.getRedisConn()
	if err != nil {
		return
	}
	reduceLuaCmd := fmt.Sprintf("local ex=redis.call('EXISTS', KEYS[1]); if (ex == 1) then local ck=redis.call('INCRBY', KEYS[1], ARGV[1]); if (ck > %v) then redis.call('DECRBY', KEYS[1], ARGV[1]) return -1  else return ck end else return -2 end", compareValue)
	p.checkCtx()
	data, err := redisClient.Do(p.Ctx, "EVAL", reduceLuaCmd, 1, key, addValue).Result()
	if err != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,key:%s,value:%d", cmd, err.Error(), key, addValue)
		return
	}
	switch data.(type) {
	case []byte:
		str := string(data.([]byte))
		num, _ = strconv.ParseInt(str, 10, 64)
		return
	case string:
		str := data.(string)
		num, _ = strconv.ParseInt(str, 10, 64)
		return
	case int64:
		num = data.(int64)
		return
	}
	msg := fmt.Sprintf("get %s command result failed:%v ,is %v", cmd, data, reflect.TypeOf(data))
	UtilLogErrorf(msg)
	err = errors.New(msg)
	p.doExpire([]string{key}, expire)
	return
}

func (p *DaoRedisEx) compareWithReduce(cmd string, key string, reduceValue int, compareValue int) (num int64, err error) {
	key = p.getKey(key)
	expire := p.getExpire(0)
	redisClient, err := p.getRedisConn()
	if err != nil {
		return
	}
	reduceLuaCmd := fmt.Sprintf("local ex=redis.call('EXISTS', KEYS[1]); if (ex == 1) then local ck=redis.call('DECRBY', KEYS[1],ARGV[1]); if (ck < %v) then redis.call('INCRBY', KEYS[1], ARGV[1]) return -1  else return ck end else return -2 end", compareValue)
	p.checkCtx()
	data, err := redisClient.Do(p.Ctx, "EVAL", reduceLuaCmd, 1, key, reduceValue).Result()
	if err != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,key:%s,value:%d", cmd, err.Error(), key, reduceValue)
		return
	}
	switch data.(type) {
	case []byte:
		str := string(data.([]byte))
		num, _ = strconv.ParseInt(str, 10, 64)
		return
	case string:
		str := data.(string)
		num, _ = strconv.ParseInt(str, 10, 64)
		return
	case int64:
		num = data.(int64)
		return
	}
	msg := fmt.Sprintf("get %s command result failed:%v ,is %v", cmd, data, reflect.TypeOf(data))
	UtilLogErrorf(msg)
	err = errors.New(msg)
	p.doExpire([]string{key}, expire)
	return
}

// incr if exist
// 之前bug: 现象: 并发量增大,程序彻底阻塞.
// 1.eval lua脚本 => get conn / defer release conn
// 2.expire key => get conn / defer release conn

// 并发上来的话,会造成连接池里的连接都被1取完,到步骤2需要wait 1释放的链接。
// 而1的连接只有等2执行完才会释放,造成死锁了
// 所以这儿要保证对应步骤用完conn直接释放,或者步骤1和2复用一个con
func (p *DaoRedisEx) doIncrNX(cmd string, key string, value int, expire int) (num int64, err error) {
	expire = p.getExpire(expire)
	key = p.getKey(key)
	num, err = p.evalIncrNX(cmd, key, value)
	if err != nil {
		return
	}
	p.doExpire([]string{key}, expire)
	return
}

// 简单封装一下
func (p *DaoRedisEx) evalIncrNX(cmd string, key string, value int) (num int64, err error) {
	var (
		data interface{}
		ok   bool
	)
	redisClient, err := p.getRedisConn()
	if err != nil {
		return
	}
	luaCmd := "local ck=redis.call('EXISTS', KEYS[1]); if (ck == 1) then return redis.call('INCRBY', KEYS[1], ARGV[1]) else return 'null' end"
	p.checkCtx()
	data, err = redisClient.Do(p.Ctx, "EVAL", luaCmd, 1, key, value).Result()
	if err != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,key:%s,value:%d", cmd, err.Error(), key, value)
		return
	}
	var luaRet []byte
	if luaRet, ok = data.([]byte); ok { // key 不存在
		if string(luaRet) == "null" {
			err = errors.New("INCRBY key not exists")
			LogErrorw(LogNameRedis, "doIncrNX",
				LogKNameCommonErr, err,
				LogKNameCommonKey, key,
				LogKNameCommonValue, value,
			)
			return
		}
	}
	var luaRetStr string
	if luaRetStr, ok = data.(string); ok { // key 不存在
		if luaRetStr == "null" {
			err = errors.New("INCRBY key not exists")
			LogInfow(LogNameRedis, "doIncrNX",
				LogKNameCommonErr, err,
				LogKNameCommonKey, key,
				LogKNameCommonValue, value,
			)
			return
		}
	}
	num, ok = data.(int64)
	if !ok {
		msg := fmt.Sprintf("get %s command result failed:%v ,is %v", cmd, data, reflect.TypeOf(data))
		UtilLogErrorf(msg)
		err = errors.New(msg)
		return
	}

	return
}

func (p *DaoRedisEx) doDel(cmd string, data ...interface{}) error {
	_, errDo := p.do(cmd, data...)
	if errDo != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,data:%v", cmd, errDo.Error(), data)
	}
	return errDo
}

func (p *DaoRedisEx) doDelWithReply(cmd string, data ...interface{}) (interface{}, error) {
	reply, errDo := p.do(cmd, data...)
	if errDo != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,data:%v", cmd, errDo.Error(), data)
	}
	return reply, errDo
}

/*基础结束*/
func (p *DaoRedisEx) Set(key string, value interface{}, ops ...OpOptionEx) (err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	_, err = p.doSet("SET", key, value, 0)
	SpanErrorFast(span, err)
	return
}

// MSet mset
func (p *DaoRedisEx) MSet(datas map[string]interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	_, err := p.doMSet("MSET", "", datas)
	SpanErrorFast(span, err)
	return err
}

// SetEx setex
func (p *DaoRedisEx) SetEx(key string, value interface{}, expire int, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	_, err := p.doSet("SET", key, value, expire)
	SpanErrorFast(span, err)
	return err
}

// Expire expire
func (p *DaoRedisEx) Expire(key string, expire int, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	key = p.getKey(key)
	_, err := p.do("EXPIRE", key, expire)
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis EXPIRE command failed: error:%s,key:%s,time:%d", err.Error(), key, expire)
		return err
	}
	return nil
}

// Persist 永久生效
func (p *DaoRedisEx) Persist(key string, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	key = p.getKey(key)
	_, err := p.do("PERSIST", key)
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis PERSIST command failed: error:%s,key:%s", err.Error(), key)
		return err
	}
	return nil
}

func (p *DaoRedisEx) Get(key string, data interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	_, err := p.doGet("GET", key, data)
	SpanErrorFast(span, err)
	return err
}

// 返回 1. key是否存在 2. error
func (p *DaoRedisEx) GetRaw(key string, data interface{}, ops ...OpOptionEx) (bool, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	b, err := p.doGet("GET", key, data)
	SpanErrorFast(span, err)
	return b, err
}

func (p *DaoRedisEx) MGet(keys []string, data interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	var args []interface{}
	for _, v := range keys {
		args = append(args, p.getKey(v))
	}
	//集群模式 args 增加 统一前缀 {大括号包裹，用于hash到一个slot上}，规则是 {}
	err := p.doMGet("MGET", args, data)
	SpanErrorFast(span, err)
	return err
}

// 封装mget通过go并发get
func (p *DaoRedisEx) MGetGo(keys []string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	err := p.doMGetGo(keys, data)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) Incr(key string, ops ...OpOptionEx) (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	n, err := p.doIncr("INCRBY", key, 1, 0)
	SpanErrorFast(span, err)
	return n, err
}

func (p *DaoRedisEx) IncrBy(key string, value int, ops ...OpOptionEx) (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	n, err := p.doIncr("INCRBY", key, value, 0)
	SpanErrorFast(span, err)
	return n, err
}

// 不操作过期时间
// deprecated
func (p *DaoRedisEx) IncrByNoExpire(key string, value int) (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span) //expire 无用
	n, err := p.doIncrNoExpire("INCRBY", key, value)
	SpanErrorFast(span, err)
	return n, err
}

// 存在key 才会自增
func (p *DaoRedisEx) IncrNX(key string, ops ...OpOptionEx) (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	n, err := p.doIncrNX("INCRBY", key, 1, 0)
	SpanErrorFast(span, err)
	return n, err
}

// 存在key 才会更新数值
func (p *DaoRedisEx) IncrByNX(key string, value int, ops ...OpOptionEx) (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	n, err := p.doIncrNX("INCRBY", key, value, 0)
	SpanErrorFast(span, err)
	return n, err
}

// 存在key, 减值，不能小于0（专用于扣减原子性）
func (p *DaoRedisEx) CompareWithReduce(key string, value int, cvalue int, ops ...OpOptionEx) (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	n, err := p.compareWithReduce("DECRBY", key, value, cvalue)
	SpanErrorFast(span, err)
	return n, err
}

func (p *DaoRedisEx) CompareWithAdd(key string, value int, cvalue int, ops ...OpOptionEx) (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	n, err := p.compareWithAdd("INCRBY", key, value, cvalue)
	SpanErrorFast(span, err)
	return n, err
}

// 针对key进行一定时间内访问次数的限流
func (p *DaoRedisEx) Limiter(key string, expire int, max int) (allow bool, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var (
		data interface{}
		ok   bool
	)
	if expire <= 0 {
		err = errors.New("limiter expire must gt 0")
		return
	}
	key = p.getKey(key)
	redisClient, err := p.getRedisConn()
	SpanErrorFast(span, err)
	if err != nil {
		return
	}
	luaCmd := `
local times = redis.call('INCR', KEYS[1])

if times == 1 then
    redis.call('expire', KEYS[1], ARGV[1])
end

if times > tonumber(ARGV[2]) then
    return 0
end

return 1
`
	p.checkCtx()
	data, err = redisClient.Do(p.Ctx, "EVAL", luaCmd, 1, key, expire, max).Result()
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,key:%s,expire:%d,max:%d", luaCmd, err.Error(), key, expire, max)
		return
	}
	var ret int64
	ret, ok = data.(int64)
	if !ok {
		msg := fmt.Sprintf("get %s command result failed:%v ,is %v", luaCmd, data, reflect.TypeOf(data))
		UtilLogErrorf(msg)
		err = errors.New(msg)
		return
	}
	if ret == 1 {
		allow = true
	}
	return
}

func (p *DaoRedisEx) SetNX(key string, value interface{}, ops ...OpOptionEx) (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	n, err := p.doSetNX("SETNX", key, value, 0)
	SpanErrorFast(span, err)
	return n, err
}

func (p *DaoRedisEx) SetNXNoExpire(key string, value interface{}, ops ...OpOptionEx) (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	n, err := p.doSetNX("SETNX", key, value, -1)
	SpanErrorFast(span, err)
	return n, err
}

func (p *DaoRedisEx) Del(key string, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	key = p.getKey(key)
	err := p.doDel("DEL", key)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) DelWithReply(key string, ops ...OpOptionEx) (ret interface{}, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	key = p.getKey(key)
	ret, err = p.doDelWithReply("DEL", key)
	SpanErrorFast(span, err)
	return
}

func (p *DaoRedisEx) MDel(key []string, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	var keys []interface{}
	for _, v := range key {
		keys = append(keys, p.getKey(v))
	}
	err := p.doDel("DEL", keys...)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) Exists(key string, ops ...OpOptionEx) (bool, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	key = p.getKey(key)
	data, err := p.do("EXISTS", key)
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis EXISTS command failed: error:%s,key:%s", err.Error(), key)
		return false, err
	}
	count, result := data.(int64)
	if !result {
		err := errors.New(fmt.Sprintf("get EXISTS command result failed:%v ,is %v", data, reflect.TypeOf(data)))
		UtilLogErrorf(err.Error())
		return false, err
	}
	if count == 1 {
		return true, nil
	}

	return false, nil
}

// hash start
func (p *DaoRedisEx) HIncrby(key string, field string, value int, ops ...OpOptionEx) (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	r, err := p.doIncr("HINCRBY", key, value, 0, field)
	SpanErrorFast(span, err)
	return r, err
}

func (p *DaoRedisEx) HGet(key string, field string, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doGet("HGET", key, value, field)
	SpanErrorFast(span, err)
	return err
}

// HGetRaw 返回 1. key是否存在 2. error
func (p *DaoRedisEx) HGetRaw(key string, field string, value interface{}) (bool, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	b, err := p.doGet("HGET", key, value, field)
	SpanErrorFast(span, err)
	return b, err
}

func (p *DaoRedisEx) HMGet(key string, fields []interface{}, data interface{}) error {
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

func (p *DaoRedisEx) HSet(key string, field string, value interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	_, err := p.doSet("HSET", key, value, 0, field)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) HSetNX(key string, field string, value interface{}, ops ...OpOptionEx) (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	r, err := p.doSetNX("HSETNX", key, value, 0, field)
	SpanErrorFast(span, err)
	return r, err
}

// HMSet value是filed:data
func (p *DaoRedisEx) HMSet(key string, value map[string]interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	_, err := p.doMSet("HMSet", key, value)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) HLen(key string, data *int) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key = p.getKey(key)
	resultData, err := p.do("HLEN", key)
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis HLEN command failed: error:%s,key:%s", err.Error(), key)
		return err
	}
	length, b := resultData.(int64)
	if !b {
		msg := fmt.Sprintf("redis data convert to int64 failed:%v", resultData)
		UtilLogErrorf(msg)
		err = errors.New(msg)
		return err
	}
	*data = int(length)
	return nil
}

func (p *DaoRedisEx) HDel(key string, data ...interface{}) error {
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
		UtilLogErrorf("run redis HDEL command failed: error:%s,key:%s,data:%v", err.Error(), key, data)
	}
	return err
}

func (p *DaoRedisEx) HExists(key string, field string) (bool, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key = p.getKey(key)
	data, err := p.do("HEXISTS", key, field)
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis HEXISTS command failed: error:%s,key:%s", err.Error(), key)
		return false, err
	}
	count, result := data.(int64)
	if !result {
		err := errors.New(fmt.Sprintf("get HEXISTS command result failed:%v ,is %v", data, reflect.TypeOf(data)))
		UtilLogErrorf(err.Error())
		return false, err
	}
	if count == 1 {
		return true, nil
	}
	return false, nil
}

// hash end

// sorted set start
func (p *DaoRedisEx) ZAdd(key string, score interface{}, value interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key = p.getKey(key)
	p.applyOpts(ops)
	expire := p.getExpire(0)
	data, errJson := json.Marshal(value)
	if errJson != nil {
		UtilLogErrorf("redis %s marshal data to json:%s", "ZAdd", errJson.Error())
		return errJson
	}
	_, errDo := p.do("ZADD", key, score, data)
	SpanErrorFast(span, errDo)
	if errDo != nil {
		UtilLogErrorf("run redis ZADD command failed: error:%s,key:%s,score:%d,data:%v", errDo.Error(), key, score, data)
		return errDo
	}
	p.doExpire([]string{key}, expire)
	return nil
}

func (p *DaoRedisEx) ZCard(key string) (data int, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key = p.getKey(key)
	var reply interface{}
	reply, err = p.do("ZCARD", key)
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis ZCARD command failed: error:%v,key:%s", err, key)
		return
	}
	if v, ok := reply.(int64); ok {
		data = int(v)
		return
	} else {
		err = errors.New(fmt.Sprintf("ZCard get replay is not int64:%v", reply))
		return
	}
}

func (p *DaoRedisEx) ZCount(key string, min, max int) (data int, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key = p.getKey(key)
	var reply interface{}
	reply, err = p.do("ZCOUNT", key, min, max)
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis ZCOUNT command failed: error:%v,key:%s,min:%d,max:%d", err, key, min, max)
		return
	}
	if v, ok := reply.(int64); ok {
		data = int(v)
		return
	} else {
		err = errors.New(fmt.Sprintf("ZCount get replay is not int64:%v", reply))
		return
	}
}

func (p *DaoRedisEx) ZIncrBy(key string, increment int, member interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key = p.getKey(key)
	p.applyOpts(ops)
	expire := p.getExpire(0)
	_, errDo := p.do("ZINCRBY", key, increment, member) //不存在 等同于 zadd
	SpanErrorFast(span, errDo)
	if errDo != nil {
		UtilLogErrorf("run redis ZINCRBY command failed: error:%s,key:%s,increment:%d,data:%v", errDo.Error(), key, increment, member)
		return errDo
	}
	p.doExpire([]string{key}, expire)
	return nil
}

func (p *DaoRedisEx) ZAddM(key string, value map[float64]interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key = p.getKey(key)
	p.applyOpts(ops)
	expire := p.getExpire(0)
	var args []interface{}
	args = append(args, key)
	for k, v := range value {
		args = append(args, k, v)
	}
	_, errDo := p.do("ZADD", args...)
	SpanErrorFast(span, errDo)
	if errDo != nil {
		UtilLogErrorf("run redis ZAddM command failed: error:%s,key:%s,data:%v", errDo.Error(), key, value)
		return errDo
	}
	p.doExpire([]string{key}, expire)
	return nil
}

// ZRANGEBYSCORE <KEY> -inf +inf limit <offset> <limit>
func (p *DaoRedisEx) ZGetByScoreLimit(key string, sort bool, offset uint32, limit uint32, value interface{}) error {
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
	if sort {
		args = append(args, "-inf")
		args = append(args, "+inf")
	} else {
		args = append(args, "+inf")
		args = append(args, "-inf")
	}
	args = append(args, "limit")
	args = append(args, offset)
	args = append(args, limit)
	err := p.doMGet(cmd, args, value)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) ZGetByScore(key string, sort bool, start int, end int, value interface{}) error {
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

// ZRANGEBYSCORE <KEY> min max limit offset limit
func (p *DaoRedisEx) ZGetByScoreWithSize(key string, sort bool, start int, end int, offset uint32, limit uint32, value interface{}) error {
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
	args = append(args, "limit")
	args = append(args, offset)
	args = append(args, limit)
	err := p.doMGet(cmd, args, value)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) ZGet(key string, sort bool, start int, end int, value interface{}) error {
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

// ZRANGEBYSCORE <KEY> -inf +inf WITHSCORES limit <offset> <limit>
func (p *DaoRedisEx) ZGetWithScoresLimit(key string, sort bool, offset uint32, limit uint32) (err error, data []ModelRedisZSetListWithScore) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var cmd string
	if sort {
		cmd = "ZRANGEBYSCORE"
	} else {
		cmd = "ZREVRANGEBYSCORE"
	}
	var args []interface{}
	args = append(args, cmd)
	args = append(args, p.getKey(key))
	if sort {
		args = append(args, "-inf")
		args = append(args, "+inf")
	} else {
		args = append(args, "+inf")
		args = append(args, "-inf")
	}
	args = append(args, "WITHSCORES")
	args = append(args, "limit")
	args = append(args, offset)
	args = append(args, limit)
	var origin []string //分别是 value score 排序

	redisClient, err := p.getRedisConn()
	if err != nil {
		return err, nil
	}
	p.checkCtx()
	origin, err = redisClient.Do(p.Ctx, args...).StringSlice()

	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis %s command failed: error:%v, args:%v", cmd, err, args)
		return err, nil
	}
	if len(origin)%2 != 0 {
		err = errors.New("redigo: k-v expects even number of values result")
		return
	}
	for i := 0; i < len(origin); {
		item := ModelRedisZSetListWithScore{}
		item.Key = origin[i]
		item.Score = origin[i+1]
		data = append(data, item)
		i = i + 2
	}
	return
}

func (p *DaoRedisEx) ZGetWithScores(key string, sort bool, start int, end int) (err error, data map[string]string) {
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

func (p *DaoRedisEx) ZGetWithScoresSlice(key string, sort bool, start int, end int) (err error, data []ModelRedisZSetListWithScore) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var cmd string
	if sort {
		cmd = "ZRANGE"
	} else {
		cmd = "ZREVRANGE"
	}
	var args []interface{}
	args = append(args, cmd)
	args = append(args, p.getKey(key))
	args = append(args, start)
	args = append(args, end)
	args = append(args, "WITHSCORES")
	var origin []string //分别是 value score 排序

	redisClient, err := p.getRedisConn()
	if err != nil {
		return err, nil
	}
	p.checkCtx()
	origin, err = redisClient.Do(p.Ctx, args...).StringSlice()

	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis %s command failed: error:%v, args:%v", cmd, err, args)
		return err, nil
	}
	if len(origin)%2 != 0 {
		err = errors.New("redigo: k-v expects even number of values result")
		return
	}
	for i := 0; i < len(origin); {
		item := ModelRedisZSetListWithScore{}
		item.Key = origin[i]
		item.Score = origin[i+1]
		data = append(data, item)
		i = i + 2
	}
	return
}

func (p *DaoRedisEx) ZRank(key string, member string, sort bool) (error, int) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var cmd string
	if sort {
		cmd = "ZRANK"
	} else {
		cmd = "ZREVRANK"
	}
	key = p.getKey(key)
	result, errDo := p.do(cmd, key, member)
	SpanErrorFast(span, errDo)
	if errDo != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,key:%s,data:%s", cmd, errDo.Error(), key, member)
		return errDo, 0
	}
	if v, ok := result.(int64); ok {
		return nil, int(v)
	} else {
		return nil, -1
	}
}

func (p *DaoRedisEx) ZScore(key string, member string, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	cmd := "ZSCORE"
	_, err := p.doGet(cmd, key, value, member)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) ZRevRange(key string, start int, end int, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	err := p.ZGet(key, false, start, end, value)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) ZRem(key string, data ...interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var args []interface{}
	key = p.getKey(key)
	args = append(args, key)
	args = append(args, data...)
	err := p.doDel("ZREM", args...)
	SpanErrorFast(span, err)
	return err
}

//list start

func (p *DaoRedisEx) LRange(start int, end int, value interface{}) (err error) {
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

func (p *DaoRedisEx) LLen() (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	cmd := "LLEN"
	key := ""
	key = p.getKey(key)
	var result interface{}
	var errDo error
	var args []interface{}
	args = append(args, key)
	result, errDo = p.do(cmd, key)
	SpanErrorFast(span, errDo)
	if errDo != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,key:%s", cmd, errDo.Error(), key)
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

func (p *DaoRedisEx) LREM(count int, data interface{}) (error, int) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key := ""
	key = p.getKey(key)
	result, errDo := p.do("LREM", key, count, data)
	SpanErrorFast(span, errDo)
	if errDo != nil {
		UtilLogErrorf("run redis command LREM failed: error:%s,key:%s,count:%d,data:%v", errDo.Error(), key, count, data)
		return errDo, 0
	}
	countRem, ok := result.(int)
	if !ok {
		msg := fmt.Sprintf("redis data convert to int failed:%v", result)
		UtilLogErrorf(msg)
		return errors.New(msg), 0
	}
	return nil, countRem
}

func (p *DaoRedisEx) LTRIM(start int, end int) (err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key := ""
	key = p.getKey(key)
	_, err = p.do("LTRIM", key, start, end)
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis command LTRIM failed: error:%v,key:%s,start:%d,end:%d", err, key, start, end)
		return
	}
	return
}

func (p *DaoRedisEx) RPush(value interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	err := p.push(value, false)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) LPush(value interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	err := p.push(value, true)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) push(value interface{}, isLeft bool) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var cmd string
	if isLeft {
		cmd = "LPUSH"
	} else {
		cmd = "RPUSH"
	}
	key := ""
	_, err := p.doSet(cmd, key, value, -1) //默认永久
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) RPop(value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	err := p.pop(value, false)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) LPop(value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	err := p.pop(value, true)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) pop(value interface{}, isLeft bool) error {
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
	SpanErrorFast(span, err)
	return err
}

//list end

//list start

func (p *DaoRedisEx) LRangeEx(key string, start int, end int, value interface{}) (err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key = p.getKey(key)
	var args []interface{}
	args = append(args, key)
	args = append(args, start)
	args = append(args, end)
	err = p.doMGet("LRANGE", args, value)
	SpanErrorFast(span, err)
	return
}

func (p *DaoRedisEx) LLenEx(key string) (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	cmd := "LLEN"
	key = p.getKey(key)
	var result interface{}
	var errDo error
	var args []interface{}
	args = append(args, key)
	result, errDo = p.do(cmd, key)
	SpanErrorFast(span, errDo)
	if errDo != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,key:%s", cmd, errDo.Error(), key)
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

func (p *DaoRedisEx) LREMEx(key string, count int, data interface{}) (error, int) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key = p.getKey(key)
	result, errDo := p.do("LREM", key, count, data)
	SpanErrorFast(span, errDo)
	if errDo != nil {
		UtilLogErrorf("run redis command LREM failed: error:%s,key:%s,count:%d,data:%v", errDo.Error(), key, count, data)
		return errDo, 0
	}
	countRem, ok := result.(int)
	if !ok {
		msg := fmt.Sprintf("redis data convert to int failed:%v", result)
		UtilLogErrorf(msg)
		return errors.New(msg), 0
	}
	return nil, countRem
}

func (p *DaoRedisEx) LTRIMEx(key string, start int, end int) (err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key = p.getKey(key)
	_, err = p.do("LTRIM", key, start, end)
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis command LTRIM failed: error:%v,key:%s,start:%d,end:%d", err, key, start, end)
		return
	}
	return
}

func (p *DaoRedisEx) RPushEx(key string, value interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	err := p.pushEx(key, value, false)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) LPushEx(key string, value interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	p.applyOpts(ops)
	err := p.pushEx(key, value, true)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) pushEx(key string, value interface{}, isLeft bool) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var cmd string
	if isLeft {
		cmd = "LPUSH"
	} else {
		cmd = "RPUSH"
	}
	_, err := p.doSet(cmd, key, value, -1) //默认永久
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) RPopEx(key string, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	err := p.popEx(key, value, false)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) LPopEx(key string, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	err := p.popEx(key, value, true)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) popEx(key string, value interface{}, isLeft bool) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var cmd string
	if isLeft {
		cmd = "LPOP"
	} else {
		cmd = "RPOP"
	}
	_, err := p.doGet(cmd, key, value)
	SpanErrorFast(span, err)
	return err
}

//list end

//pipeline start

func (p *DaoRedisEx) PipelineHGet(key []string, fields []interface{}, data []interface{}) error {
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

func (p *DaoRedisEx) pipeDoGet(cmd string, args [][]interface{}, value []interface{}) error {
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
			UtilLogErrorf("Receive(%v) returned error %v", v, err)
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
			UtilLogErrorf("get %s command result failed:%s", cmd, errorJson.Error())
			return errorJson
		}
	}
	return nil
}

//pipeline end

// Set集合Start
func (p *DaoRedisEx) SAdd(key string, argPs []interface{}, ops ...OpOptionEx) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	args := make([]interface{}, len(argPs)+1)
	key = p.getKey(key)
	args[0] = key
	p.applyOpts(ops)
	expire := p.getExpire(0)
	copy(args[1:], argPs)
	_, errDo := p.do("SADD", args...)
	SpanErrorFast(span, errDo)
	if errDo != nil {
		UtilLogErrorf("run redis SADD command failed: error:%s,key:%s,args:%v", errDo.Error(), key, args)
		return errDo
	}
	p.doExpire([]string{key}, expire)
	return errDo
}

func (p *DaoRedisEx) SIsMember(key string, arg interface{}) (b bool, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	key = p.getKey(key)
	var reply interface{}
	reply, err = p.do("SISMEMBER", key, arg)
	SpanErrorFast(span, err)
	if err != nil {
		UtilLogErrorf("run redis SISMEMBER command failed: error:%v,key:%s,member:%s", err, key, arg)
		return
	}
	if code, ok := reply.(int64); ok && code == int64(1) {
		b = true
	}
	return
}

func (p *DaoRedisEx) SCard(key string) int64 {
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

func (p *DaoRedisEx) SRem(key string, argPs []interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	args := make([]interface{}, len(argPs)+1)
	args[0] = p.getKey(key)
	copy(args[1:], argPs)
	_, errDo := p.do("SREM", args...)
	SpanErrorFast(span, errDo)
	if errDo != nil {
		UtilLogErrorf("run redis SREM command failed: error:%s,key:%s,member:%s", errDo.Error(), key, args)
	}
	return errDo
}

func (p *DaoRedisEx) SPop(key string, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	_, err := p.doGet("SPOP", key, value)
	SpanErrorFast(span, err)
	return err
}

func (p *DaoRedisEx) SMembers(key string) (data []interface{}, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var args []interface{}
	args = append(args, p.getKey(key))
	err, data = p.doGetSlice("SMEMBERS", args...)
	SpanErrorFast(span, err)
	return
}

func (p *DaoRedisEx) HGetAll(key string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	var args []interface{}

	args = append(args, p.getKey(key))

	err := p.doMGet("HGETALL", args, data)
	SpanErrorFast(span, err)

	return err
}

func (p *DaoRedisEx) HGetAllStringMap(key string) (err error, data map[string]string) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	args := p.getKey(key)
	err, data = p.doMGetStringMap("HGETALL", args)
	SpanErrorFast(span, err)
	return
}

func (p *DaoRedisEx) HGetAllIntMap(key string) (err error, data map[string]int64) {
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
func (p *DaoRedisEx) GetPTtl(key string) (ttl int64, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	ttl, err = p.doGetTtl("PTTL", key)
	SpanErrorFast(span, err)
	return
}

// GetTtl：获取key的过期时间，单位为秒
// 如果key不存在返回-2
// 如果key存在，但是没有设置过期时间，返回-1
func (p *DaoRedisEx) GetTtl(key string) (ttl int64, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoRedis, RunFuncNameUp(), v3.SpanLayer_Cache)
	defer SpanEnd(span)
	ttl, err = p.doGetTtl("TTL", key)
	SpanErrorFast(span, err)
	return
}

func (p *DaoRedisEx) doGetTtl(cmd string, key string) (ttl int64, err error) {
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
