package gozen

import (
	"sync"
)

var (
	cacheConfigMux sync.Mutex
	cacheConfig    *ConfigCache
)

/**

模式优先级


- NoProxy
- Dynamic
- self


*/

type ConfigCache struct {
	Redis   ConfigCacheRedis   `yaml:"Redis"`
	RedisP  ConfigCacheRedis   `yaml:"RedisP"` // 持久化Redis
	Dynamic ConfigCacheDynamic `yaml:"Dynamic"`
	NoProxy ConfigCacheNoProxy `yaml:"NoProxy"` // 平台去代理模式
}

type ConfigCacheRedis struct {
	Address         []string `yaml:"Address"` //多个ip节点，每个都是单独访问，循环建立连接
	addressLen      int
	Bid             string `yaml:"Bid"`             //bid
	Prefix          string `yaml:"Prefix"`          //前缀
	Cluster         bool   `yaml:"Cluster"`         //是否集群模式 集群模式走不一样的初始化和curd
	Expire          int    `yaml:"Expire"`          // 缓存默认时间
	ReadTimeout     int    `yaml:"ReadTimeout"`     //读超时时间
	WriteTimeout    int    `yaml:"WriteTimeout"`    //写超时时间
	ConnectTimeout  int    `yaml:"ConnectTimeout"`  //建立连接的超时时间
	PoolMaxActive   int    `yaml:"PoolMaxActive"`   //池子最大连接数
	PoolMinActive   int    `yaml:"PoolMinActive"`   //池子空闲连接数
	PoolIdleTimeout int    `yaml:"PoolIdleTimeout"` //连接存活时间 超出时间后 连接会重置 要小于服务端链接存活时间
	Password        string `yaml:"Password"`
}

type ConfigCacheDynamic struct {
	IsDynamic      bool   `yaml:"IsDynamic"`      //是否开启代理
	DynamicAddress string `yaml:"DynamicAddress"` //代理地址
	CycleTime      int    `yaml:"CycleTime"`      //循环获取时间 秒
	CycleFlagTime  int    `yaml:"CycleFlagTime"`
}

type ConfigCacheNoProxy struct {
	IsNoProxy      bool   `yaml:"IsNoProxy"`      //是否开启代理
	ServiceAddress string `yaml:"ServiceAddress"` //获取服务地址
	MonitorAddress string `yaml:"MonitorAddress"` //监控上报地址
	Lang           string `yaml:"Lang"`           //lang
	AppCode        string `yaml:"AppCode"`        //app code
	Iv             string `yaml:"Iv"`             //iv
	MonitorTime    int    `yaml:"MonitorTime"`    //心跳时间 秒
}

func configCacheInit() {
	if cacheConfig == nil || len(cacheConfig.Redis.Address) == 0 {
		configFileName := "cache"
		if cfp.configPathExist(configFileName) {
			cacheConfigMux.Lock()
			defer cacheConfigMux.Unlock()
			cacheConfig = new(ConfigCache)
			err := cfp.configGet(configFileName, nil, cacheConfig, nil)
			if err != nil {
				panic("configCacheInit error:" + err.Error())
			} else {
				//判断config 代理和集群互斥
				if (cacheConfig.Redis.Cluster || cacheConfig.RedisP.Cluster) && cacheConfig.Dynamic.IsDynamic {
					panic("configCacheInit error: 代理和集群互斥")
				}
				//判断config 代理和no代理互斥
				if cacheConfig.NoProxy.IsNoProxy && cacheConfig.Dynamic.IsDynamic {
					panic("configCacheInit error: 代理和no代理互斥")
				}

				if !cacheConfig.NoProxy.IsNoProxy {
					//config 初始化操作
					cacheConfig.Redis.addressLen = len(cacheConfig.Redis.Address)
					cacheConfig.RedisP.addressLen = len(cacheConfig.RedisP.Address)
					initRedisPoll(true)
					initRedisPoll(false)
				}

				// 初始化redis 模式
				initNoProxyRedisAddressAll()
				initDynamicRedisAddress()
			}
		}
	}
	return
}

func configCacheReload() {
	cacheConfigMux.Lock()
	defer cacheConfigMux.Unlock()
	cacheConfig = nil
	configCacheInit()
}

func ConfigCacheGetRedis() *ConfigCacheRedis {
	if cacheConfig == nil {
		return new(ConfigCacheRedis)
	}
	return &cacheConfig.Redis
}

func ConfigCacheGetRedisBaseWithConn(persistent bool) *ConfigCacheRedis {
	var redisConfig ConfigCacheRedis
	if !persistent {
		redisConfig = cacheConfig.Redis
	} else {
		redisConfig = cacheConfig.RedisP
	}

	return &redisConfig
}

func ConfigCacheGetRedisWithConn(persistent bool) *ConfigCacheRedis {
	var redisConfig ConfigCacheRedis
	if !persistent {
		redisConfig = cacheConfig.Redis
	} else {
		redisConfig = cacheConfig.RedisP
	}

	if cacheConfig.NoProxy.IsNoProxy {
		// local info
		info := getNoProxyLocalInfo(persistent)
		redisConfig.Cluster = info.IsCluster
		if info.NeedAuth {
			redisConfig.Password = info.ClusterPassWord
		}
		redisConfig.addressLen = info.AddressLen
		redisConfig.Address = info.Address
	}

	return &redisConfig
}

func ConfigCacheGetRedisDynamic() *ConfigCacheDynamic {
	return &cacheConfig.Dynamic
}

func isCacheGetRedisDynamic() bool {
	return ConfigCacheGetRedisDynamic().IsDynamic
}

func isCacheCluster(persistent bool) bool {
	if persistent {
		return cacheConfig.RedisP.Cluster
	}
	return cacheConfig.Redis.Cluster
}
