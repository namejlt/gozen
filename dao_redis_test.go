package gozen

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

//检测 主从
//检测 cluster
//检测 代理

type ModelRedisHello struct {
	HelloWord string `json:"hello_word"`
	Age       int    `json:"age"`
	Name      string `json:"name"`
}

var _ encoding.BinaryMarshaler = (*ModelRedisHello)(nil)

func (p *ModelRedisHello) MarshalBinary() (data []byte, err error) {
	return json.Marshal(*p)
}

func (p *ModelRedisHello) UnmarshalBinary(data []byte) (err error) {
	return JSONDecodeUseNumber(data, p)
}

func Test_Redis_encode(t *testing.T) {
	a := ModelRedisHello{
		Name:      "namehh",
		HelloWord: "这是好后好adad12121",
		Age:       999,
	}
	var buf bytes.Buffer
	fmt.Fprint(&buf, a)
	fmt.Println(buf.String())
	b := ModelRedisHello{}
	JSONDecodeUseNumber(buf.Bytes(), &b)
	fmt.Println(b)
}

func Benchmark_Redis_MGetGo(b *testing.B) {
	redis := NewRedisTest()
	for i := 0; i < b.N; i++ {
		key := []string{
			"aa01",
			"aa02",
			"aa03",
			"aa04",
		}
		data := ModelRedisHello{
			HelloWord: "jialongtian",
		}
		for _, v := range key {
			err := redis.DaoRedisEx.Set(v, data)
			if err != nil {
				b.Error(err)
				return
			}
		}
		var res []ModelRedisHello
		key = append(key, "aa05")
		err := redis.MGetGo(key, &res)
		if err != nil {
			b.Error(err)
		}
	}
}

func Test_Redis_MGetGo(t *testing.T) {
	redis := NewRedisTest()
	key := []string{
		"aa01",
		"aa02",
		"aa03",
		"aa04",
	}
	data := ModelRedisHello{
		HelloWord: "jialongtian",
	}
	for _, v := range key {
		err := redis.DaoRedisEx.Set(v, data)
		if err != nil {
			t.Error(err)
			return
		}
	}
	var res []ModelRedisHello
	key = append(key, "aa05")
	err := redis.MGetGo(key, &res)
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		fmt.Println(v)
	}
}

func Test_Redis_ForProxyMGetGo(t *testing.T) {
	for i := 0; i < 10; i++ {
		redis := NewRedisTest()
		key := []string{
			"aa01",
			"aa02",
			"aa03",
			"aa04",
		}
		data := ModelRedisHello{
			HelloWord: "jialongtian",
			Age:       1231,
			Name:      "是是是",
		}
		for _, v := range key {
			err := redis.DaoRedisEx.Set(v, data)
			if err != nil {
				t.Error(err)
				return
			}
		}
		var res []ModelRedisHello
		key = append(key, "aa05")
		err := redis.MGetGo(key, &res)
		if err != nil {
			t.Error(err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func Test_Redis_ProxyMGetGo(t *testing.T) {
	redis := NewRedisTest()
	key := []string{
		"aa01",
		"aa02",
		"aa03",
		"aa04",
	}
	data := ModelRedisHello{
		HelloWord: "jialongtian",
		Age:       1231,
		Name:      "是是是",
	}
	for _, v := range key {
		err := redis.DaoRedisEx.Set(v, data)
		if err != nil {
			t.Error(err)
			return
		}
	}
	var res []ModelRedisHello
	key = append(key, "aa05")
	err := redis.MGetGo(key, &res)
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		fmt.Println(v)
	}
}

func Test_Redis_INCRBYNX(t *testing.T) {
	redis := NewRedisTest()
	key := "key_with_incr_nx"
	err := redis.DaoRedisEx.Set(key, 1)
	if err != nil {
		t.Error(err)
	}
	num, err := redis.IncrByNX(key, 100)
	//fmt.Println("set", num, err)
	if err != nil {
		t.Error(err)
	}
	if num != 101 {
		t.Error("num error", num)
	}
	err = redis.DaoRedisEx.Del(key)
	if err != nil {
		t.Error(err)
	}
	num, err = redis.IncrByNX(key, 100)
	if err == nil {
		t.Log(err)
	}
	if num != 0 {
		t.Error("num error", num)
	}
}

func Test_Redis_LIMITER(t *testing.T) {
	redis := NewRedisTest()
	key := "key_with_limiter"
	err := redis.DaoRedisEx.Del(key)
	if err != nil {
		t.Error(err)
	}
	var (
		expire = 10
		maxNum = 10
	)
	for i := 1; i <= maxNum+1; i++ {
		allow, err := redis.DaoRedisEx.Limiter(key, expire, maxNum)
		if err != nil {
			t.Error(err)
		}
		if i <= maxNum {
			if !allow {
				t.Error("allow must true")
			}
		} else {
			if allow {
				t.Error("allow must false")
			}
		}
	}
}

func Test_Redis_GetTtl(t *testing.T) {
	redis := NewRedisTest()
	// key存在有有效期
	redis.SetEx("key_with_expire", "123456", 100)
	ttl, err := redis.GetTtl("key_with_expire")
	if err != nil {
		t.Error(err)
	}
	if ttl <= 0 {
		t.Error("test key_with_expire:result err")
	}
	// key不存在
	ttl, err = redis.GetTtl("key_not_exist")
	if err != nil {
		t.Error(err)
	}
	if ttl != -2 {
		t.Error("test key_not_exist:result err")
	}
}

func Test_Redis_Set(t *testing.T) {
	redis := NewRedisTest()
	value := "s:35:\"h5_9989f070d21dc983cc7bbc5c6a013080\";"
	redis.Set("tonyjt", value)
	var data string
	result := redis.Get("tonyjt", &data)
	if result != nil {
		t.Error("result false")
	}
	if data != value {
		t.Error("get false")
	} else {
		t.Log(data)
	}
}

func Test_Redis_Get(t *testing.T) {
	redis := NewRedisTest()
	var data string
	result := redis.Set("h5_2", "haha")
	if result != nil {
		t.Error("result set false")
	}
	result = redis.Get("h5_2", &data)
	if result != nil {
		t.Error("result get false")
	}
	fmt.Println(data)
}

func Test_Redis_PushEx(t *testing.T) {
	redis := NewRedisTest()
	result := redis.PushList("PushList", "dadadadasd", 600)
	if result != nil {
		t.Error("result false")
	}
}

func Test_Redis_SetEx(t *testing.T) {
	redis := NewRedisTest()
	result := redis.SetEx("setex", "asdfsdf", 60)
	if result != nil {
		t.Error("result false")
	}
}

func Test_Redis_Incr(t *testing.T) {
	redis := NewRedisTest()
	_, result := redis.Incr("incr", WithExpire(600))
	if result != nil {
		t.Error("result false")
	}
}

func Test_Redis_HSet(t *testing.T) {
	redis := NewRedisTest()
	result := redis.HSet("hset", "k1", "sdfsdf")
	if result != nil {
		t.Error("result false")
	}
}

func Test_Redis_HSetNX(t *testing.T) {
	redis := NewRedisTest()
	_, result := redis.HSetNX("hsetnx", "h1", "123123")
	if result != nil {
		t.Error("result false")
	}
}

func Test_Redis_Del(t *testing.T) {
	redis := NewRedisTest()
	result := redis.Del("tonyjt")
	if result != nil {
		t.Error("result false")
	}
}
func Test_Redis_HIncrby(t *testing.T) {
	redis := NewRedisTest()
	_, result := redis.HIncrby("hincr", "1", 1)
	if result != nil {
		t.Error("result false")
	}
}

func Test_Redis_HMSet(t *testing.T) {
	redis := NewRedisTest()
	datas := make(map[string]ModelRedisHello)
	datas["1"] = ModelRedisHello{HelloWord: "HelloWord1"}
	datas["2"] = ModelRedisHello{HelloWord: "HelloWord2"}
	result := redis.HMSet("hmset1", datas)
	if result != nil {
		t.Errorf("result false")
	}
}

func Test_Redis_ZAddM(t *testing.T) {
	redis := NewRedisTest()
	datas := make(map[int]int)
	datas[3] = 3
	datas[2] = 2
	datas[1] = 1
	result := redis.ZAddM("zaddm1", datas, 600)
	if result != nil {
		t.Errorf("result false")
	}
}

func Test_Redis_ZRem(t *testing.T) {
	redis := NewRedisTest()
	result := redis.ZRem("zaddm1", 2, 3)
	if result != nil {
		t.Errorf("result false")
	}
}

func Test_Redis_MSet(t *testing.T) {
	tes := NewRedisTest()
	value := make(map[string]ModelRedisHello)
	value["mset1"] = ModelRedisHello{HelloWord: "1"}
	value["mset2"] = ModelRedisHello{HelloWord: "2"}
	value["mset3"] = ModelRedisHello{HelloWord: "3"}
	result := tes.MSet(value, 60)
	if result != nil {
		t.Error("result false")
	}
}

func Test_Redis_MGet(t *testing.T) {
	redis := NewRedisTest()
	value, err := redis.MGet("mset1", "mset2", "mset4", "mset3") //没有 mset4
	if err != nil {
		t.Errorf("result false:%s", err.Error())
	} else if len(value) != 4 {
		t.Errorf("len is < 4:%d", len(value))
	}
	for _, v := range value {
		t.Log(v)
	}
}

func Test_Redis_HDel(t *testing.T) {
	redis := NewRedisTest()
	key := "hmset1"
	result := redis.HDel(key, "1", "2")
	if result != nil {
		t.Error("result false")
	}
}

func Test_Redis_HMGet(t *testing.T) {
	redis := NewRedisTest()
	data, err := redis.HMGet("hmset1", "1", "2", "3")
	if err != nil {
		t.Errorf("result false:%s", err.Error())
	} else if len(data) != 3 {
		t.Errorf("len is lt :%d", len(data))
	}
}

func Test_Redis_SADD(t *testing.T) {
	redis := NewRedisTest()
	err := redis.SAdd("set001", []interface{}{"12", "2d3", "35t"})
	if err != nil {
		t.Errorf("result false:%s", err.Error())
	}
}

func Test_Redis_SMembers(t *testing.T) {
	redis := NewRedisTest()
	data, err := redis.SMembers("set001")
	if err != nil {
		t.Errorf("result false:%s", err.Error())
	}
	t.Log(data)
}

func Test_Redis_ZRevRange(t *testing.T) {
	redis := NewRedisTest()
	setName := "setaddhaha"
	for i := 0; i < 10; i++ {
		err := redis.ZAdd(setName, i, ModelRedisHello{
			HelloWord: "this i sss" + fmt.Sprint(i),
			Age:       i,
		})
		if err != nil {
			t.Error(err)
			return
		}
	}

	data, err := redis.ZRevRange(setName, 0, 3)
	if err != nil {
		t.Errorf("result false:%s", err.Error())
	} else {
		t.Logf("result,len:%d,d:%v", len(data), data)
	}
}

func Test_Redis_ZGetWithScoresLimit(t *testing.T) {
	redis := NewRedisTest()
	zSetKey := "test_zset-001"
	err := redis.Del(zSetKey)
	if err != nil {
		t.Errorf("redis.Del(zSetKey):%s", err.Error())
	}
	//添加数据
	err = redis.ZAdd(zSetKey, 1.234, "a")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.234, "b")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.234, "c")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.234, "d")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.234, "aa")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.234, "cc")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	//返回带积分数组
	err, data := redis.ZGetWithScoresLimit(zSetKey, true, 0, 6)
	if err != nil {
		t.Errorf("redis.ZGetWithScoresLimit(zSetKey):%s", err.Error())
	} else {
		t.Log(data)
	}
}

func Test_Redis_ZGetWithScoresSlice(t *testing.T) {
	redis := NewRedisTest()
	zSetKey := "test_zset-001"
	err := redis.Del(zSetKey)
	if err != nil {
		t.Errorf("redis.Del(zSetKey):%s", err.Error())
	}
	//添加数据
	err = redis.ZAdd(zSetKey, 1.234, "a")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.234, "b")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.234, "c")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.234, "d")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.234, "aa")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.234, "cc")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	//返回带积分数组
	err, data := redis.ZGetWithScoresSlice(zSetKey, true, 0, -1)
	if err != nil {
		t.Errorf("redis.ZGetWithScoresSlice(zSetKey):%s", err.Error())
	} else {
		t.Log(data)
	}
}

func Test_Redis_ZRank(t *testing.T) {
	redis := NewRedisTest()
	zSetKey := "test_zset-001"
	err := redis.Del(zSetKey)
	if err != nil {
		t.Errorf("redis.Del(zSetKey):%s", err.Error())
	}
	//添加数据
	err = redis.ZAdd(zSetKey, 1.231, "a")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.232, "b")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.233, "c")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.234, "d")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.235, "dd")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	err = redis.ZAdd(zSetKey, 1.236, "cc")
	if err != nil {
		t.Errorf("redis.ZAdd(zSetKey):%s", err.Error())
	}
	//返回带积分数组
	err, data := redis.ZRank(zSetKey, "dd", true)
	if err != nil {
		t.Errorf("redis.ZRank(zSetKey):%s", err.Error())
	} else {
		t.Log(data)
	}
}

type TestDaoRedis struct {
	DaoRedisEx
}

func NewRedisTest() *TestDaoRedis {
	componentDao := &TestDaoRedis{
		DaoRedisEx{
			KeyName:    "test",
			Ctx:        context.Background(),
			Persistent: false,
		},
	}
	return componentDao
}

func (c *TestDaoRedis) PushList(key string, value string, expire int) error {
	return c.DaoRedisEx.LPushEx(key, value, WithExpire(expire))
}

func (c *TestDaoRedis) Set(name string, key string) error {
	return c.DaoRedisEx.Set(name, key)
}

func (c *TestDaoRedis) Get(name string, key *string) error {
	return c.DaoRedisEx.Get(name, key)
}

func (c *TestDaoRedis) MSet(value map[string]ModelRedisHello, expire int) error {
	datas := make(map[string]interface{})
	for k, v := range value {
		datas[k] = v
	}
	return c.DaoRedisEx.MSet(datas, WithExpire(expire), WithClusterKey())
}

func (c *TestDaoRedis) MGet(keys ...string) ([]*ModelRedisHello, error) {
	var datas []*ModelRedisHello
	err := c.DaoRedisEx.MGet(keys, &datas, WithClusterKey())
	return datas, err
}

func (c *TestDaoRedis) Del(key string) error {
	return c.DaoRedisEx.Del(key)
}

func (c *TestDaoRedis) HMSet(key string, value map[string]ModelRedisHello) error {
	datas := make(map[string]interface{})
	for k, v := range value {
		datas[k] = v
	}
	return c.DaoRedisEx.HMSet(key, datas)
}

func (c *TestDaoRedis) HMGet(key string, fields ...string) ([]*ModelRedisHello, error) {
	var datas []*ModelRedisHello
	var args []interface{}
	for _, item := range fields {
		args = append(args, item)
	}
	err := c.DaoRedisEx.HMGet(key, args, &datas)
	return datas, err
}

func (c *TestDaoRedis) ZAddM(key string, value map[int]int, expire int) error {
	datas := make(map[float64]interface{})
	for k, v := range value {
		datas[float64(k)] = v
	}
	return c.DaoRedisEx.ZAddM(key, datas, WithExpire(expire))
}

func (c *TestDaoRedis) ZRem(key string, data ...interface{}) error {
	return c.DaoRedisEx.ZRem(key, data...)
}

func (c *TestDaoRedis) ZRevRange(key string, start int, end int) ([]ModelRedisHello, error) {
	var data []*ModelRedisHello
	err := c.DaoRedisEx.ZRevRange(key, start, end, &data)
	var value []ModelRedisHello
	if err == nil {
		for _, item := range data {
			if item != nil {
				value = append(value, *item)
			} else {
				value = append(value, ModelRedisHello{})
			}
		}
	}
	return value, err
}

func TestGetAddress(t *testing.T) {
	a := []int{1, 2, 3, 4}
	b := len(a)

	for i := 0; i < 100; i++ {
		fmt.Println(i % b)
	}
}
