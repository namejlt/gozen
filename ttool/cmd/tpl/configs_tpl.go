package tpl

var (
	ConfigsDirName   = "configs"
	ConfigsFilesName = []string{
		"app.json",
		"cache.json",
		"code.json",
		"db.json",
		"es.json",
		"mongo_cluster.json",
		"mysql_cluster.json",
		"pool.json",
		"tracer.json",
		"project.yaml",
	}
	ConfigsFilesContent = []string{
		configsAppJson,
		configsCacheJson,
		configsCodeJson,
		configsDbJson,
		configsEsJson,
		configsMongoClusterJson,
		configsMysqlClusterJson,
		configsPoolJson,
		configsTracerJson,
		configsProjectYaml,
	}
)

var (
	configsAppJson = `{
  "Configs": {
    "Env": "dev",
    "Docs": "true",
    "Host": "localhost:8888",
    "DebugVarsPrefix": "{{.Name}}",
    "ServerType": "web",
    "SysName": "{{.Name}}-service"
  }
}
`
	configsCacheJson = `{
  "Redis": {
    "Address": [
      "127.0.0.1:63790"
    ],
    "Prefix": "{{.Name}}",
    "Expire": 604800,
    "ConnectTimeout": 10000,
    "ReadTimeout": 10000,
    "WriteTimeout": 10000,
    "PoolMaxActive": 100,
    "PoolMinActive": 10,
    "PoolIdleTimeout": 180000,
    "Password": ""
  },
  "RedisP": {
    "Address": [
      "127.0.0.1:63790"
    ],
    "Prefix": "{{.Name}}_p",
    "Expire": 604800,
    "ConnectTimeout": 10000,
    "ReadTimeout": 10000,
    "WriteTimeout": 10000,
    "PoolMaxActive": 100,
    "PoolMinActive": 10,
    "PoolIdleTimeout": 180000,
    "Password": ""
  },
  "Dynamic": {
    "DynamicAddress": "http://127.0.0.1:5262/redisServerHttpProxy/obtainRedisProxy",
    "IsDynamic": false,
    "CycleTime": 300
  }
}
`

	configsCodeJson = `{
  "Codes": {
    "1001": "成功",
    "1002": "验证签名失败",
    "1003": "服务器繁忙，请稍后再试",
    "1004": "请求参数错误",
    "1005": "用户未登录",
    "999999": "code描述信息"
  }
}
`

	configsDbJson = `{
  "Mysql": {
    "DbName": "{{.Name}}",
    "Pool": {
      "PoolMinCap": 10,
      "PoolExCap": 5,
      "PoolMaxCap": 40,
      "PoolIdleTimeout": 3600,
      "PoolWaitCount": 1000,
      "PoolWaitTimeout": 30
    },
    "Write": {
      "Address": "127.0.0.1",
      "Port": 33066,
      "User": "root",
      "Password": "password"
    },
    "Reads": [
      {
        "Address": "127.0.0.1",
        "Port": 33066,
        "User": "root",
        "Password": "password"
      }
    ]
  },
  "Mongo": {
    "DbName": "{{.Name}}",
    "Servers": "127.0.0.1:34004",
    "Read_option": "PRIMARY",
    "Timeout": 10000
  }
}
`

	configsMysqlClusterJson = `[
  {
    "DbNum": 1001,
    "DbName": "{{.Name}}",
    "Pool": {
      "PoolMinCap": 10,
      "PoolExCap": 5,
      "PoolMaxCap": 40,
      "PoolIdleTimeout": 3600,
      "PoolWaitCount": 1000,
      "PoolWaitTimeout": 30
    },
    "Write": {
      "Address": "127.0.0.1",
      "Port": 33066,
      "User": "root",
      "Password": "password"
    },
    "Reads": [
      {
        "Address": "127.0.0.1",
        "Port": 33066,
        "User": "root",
        "Password": "password"
      }
    ]
  }
]
`

	configsPoolJson = `{
  "Configs": {
    "grpc-test": {
      "Address": [
        "127.0.0.1:7001",
        "127.0.0.1:7007"
      ],
      "MaxConcurrentStreams": 10,
      "MaxActive": 10,
      "MaxIdle": 1,
      "Reuse": true
    },
    "grpc-test2": {
      "Address": [
        "127.0.0.1:7001",
        "127.0.0.1:7007"
      ],
      "MaxConcurrentStreams": 10,
      "MaxActive": 10,
      "MaxIdle": 1,
      "Reuse": true
    }
  }
}
`

	configsTracerJson = `{
  "Tracer": {
    "ServiceName": "{{.Name}}",
    "Disabled": false,
    "Sampler": {
      "SamplingRate": 1
    },
    "Reporter": {
      "LogSpans": true,
      "BufferFlushInterval": 1,
      "LocalAgentHostPort": "skywalking-server:11800"
    }
  }
}
`

	configsMongoClusterJson = `[
  {
    "DbNum": 2001,
    "DbName": "{{.Name}}",
    "Servers": "127.0.0.1:34006",
    "ReadOption": "primary",
    "Options": "authSource=admin",
    "User": "mongoadmin",
    "Password": "mongopassword",
    "Timeout": 60,
    "MaxPoolSize": 100,
    "MinPoolSize": 10,
    "SocketTimeout": 30,
    "ConnectTimeout": 30,
    "MaxConnIdleTime": 600,
    "ServerSelectionTimeout": 30
  },
  {
    "DbNum": 2002,
    "DbName": "{{.Name}}",
    "Servers": "127.0.0.1:34006",
    "ReadOption": "primary",
    "Options": "authSource=admin",
    "User": "mongoadmin",
    "Password": "mongopassword",
    "Timeout": 60,
    "MaxPoolSize": 100,
    "MinPoolSize": 10,
    "SocketTimeout": 30,
    "ConnectTimeout": 30,
    "MaxConnIdleTime": 600,
    "ServerSelectionTimeout": 30
  }
]
`

	configsEsJson = `{
  "Address": [
    "ip"
  ],
  "Timeout": 3000,
  "HealthcheckTimeout": 1,
  "HealthcheckInterval": 60,
  "HealthcheckEnabled": true,
  "SnifferEnabled": true,
  "TransportMaxIdel": 10
}
`

	configsProjectYaml = `# 综合配置
base:
  # 形式 支持 local nacos，如果本地则直接读取本地文件，固定5分钟更新内存；如果是nacos，根据配置读取更新本地文件
  mode: "local"
  # 格式 支持  json yaml
  format: "json"
  # 配置路径
  configs_path: "configs"

# 配置中心
nacos:
  # nacos 地址
  addr: "127.0.0.1"
  # nacos 端口
  port: 8848
  # 命名空间
  namespace_id: "1abf0eb7-8495-4c66-b003-90fa6a2f8a00"
  # 组名称
  group: "DEFAULT_GROUP"
  # 命名空间
  data_id: [ "app","cache","code","db","es","mongo_cluster","mysql_cluster","pool","tracer" ]
  # 超时时间 毫秒
  timeout_ms: 5000
  # 日志等级 debug,info,warn,error
  log_level: "info"
  # 周期更新时间 秒
  interval: 60

# 日志配置
log:
  # 名称 格式 appcode_app appcode_script
  name: "appcode_app"
  # 路径 /data/logs/appcode_app.log
  path: "./appcode_app.log"
  # 是否debug
  debug: true
  # 日志大小 m
  max_size: 50
  # 日志存在时间 天
  max_age: 10
  # 日志最多保存几个
  max_backups: 2
  # 是否压缩
  compress: false

`
)
