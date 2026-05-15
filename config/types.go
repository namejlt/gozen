package config

// =========================================================================
// Standard configuration types used by the framework.
// These mirror the original root-package config structs but live in the
// config package so they can be loaded via Manager.Load().
// =========================================================================

// AppConfig holds application-level settings.
type AppConfig struct {
	Env        string `yaml:"Env" json:"env"`
	Host       string `yaml:"Host" json:"host"`
	Docs       string `yaml:"Docs" json:"docs"`
	DebugVars  string `yaml:"DebugVarsPrefix" json:"debug_vars_prefix"`
	SignSwitch string `yaml:"SignSwitch" json:"sign_switch"`
	AppSecret  string `yaml:"AppSecretKey" json:"app_secret_key"`
	LimitTime  string `yaml:"AppAccessLimitTime" json:"app_access_limit_time"`
	OutFormat  string `yaml:"ApiOutFormat" json:"api_out_format"`
	Extras     map[string]any `yaml:"Configs" json:"configs"`
}

// DBConfig holds database and cache connection parameters.
type DBConfig struct {
	Mysql MysqlConfig `yaml:"Mysql" json:"mysql"`
	Mongo MongoConfig `yaml:"Mongo" json:"mongo"`
	Redis RedisConfig `yaml:"Redis" json:"redis"`
}

// MysqlConfig is a single MySQL instance config (backward compat).
type MysqlConfig struct {
	DbNum  uint32        `yaml:"DbNum"  json:"db_num"`
	DbName string        `yaml:"DbName" json:"db_name"`
	Pool   DBPoolConfig  `yaml:"Pool"   json:"pool"`
	Write  DBNodeConfig  `yaml:"Write"  json:"write"`
	Reads  []DBNodeConfig `yaml:"Reads" json:"reads"`
}

// DBNodeConfig holds connection info for a single DB node.
type DBNodeConfig struct {
	Address  string `yaml:"Address"  json:"address"`
	Port     int    `yaml:"Port"     json:"port"`
	User     string `yaml:"User"     json:"user"`
	Password string `yaml:"Password" json:"password"`
	DbName   string `yaml:"-"        json:"-"`
}

// DBPoolConfig holds connection-pool parameters.
type DBPoolConfig struct {
	MinCap      int `yaml:"PoolMinCap"      json:"min_cap"`
	MaxCap      int `yaml:"PoolMaxCap"      json:"max_cap"`
	IdleTimeout int `yaml:"PoolIdleTimeout" json:"idle_timeout"`
	LifeTimeout int `yaml:"PoolLifeTimeout" json:"life_timeout"`
}

// MongoConfig holds MongoDB connection parameters.
type MongoConfig struct {
	DbNum                  uint32 `yaml:"DbNum"                  json:"db_num"`
	DbName                 string `yaml:"DbName"                 json:"db_name"`
	Options                string `yaml:"Options"                json:"options"`
	User                   string `yaml:"User"                   json:"user"`
	Password               string `yaml:"Password"               json:"password"`
	Servers                string `yaml:"Servers"                json:"servers"`
	ReadOption             string `yaml:"ReadOption"             json:"read_option"`
	Timeout                int    `yaml:"Timeout"                json:"timeout"`
	MaxPoolSize            uint64 `yaml:"MaxPoolSize"            json:"max_pool_size"`
	MinPoolSize            uint64 `yaml:"MinPoolSize"            json:"min_pool_size"`
	SocketTimeout          int    `yaml:"SocketTimeout"          json:"socket_timeout"`
	ConnectTimeout         int    `yaml:"ConnectTimeout"         json:"connect_timeout"`
	MaxConnIdleTime        int    `yaml:"MaxConnIdleTime"        json:"max_conn_idle_time"`
	ServerSelectionTimeout int    `yaml:"ServerSelectionTimeout" json:"server_selection_timeout"`
}

// CacheConfig holds Redis / cache configuration.
type CacheConfig struct {
	Redis   RedisConfig   `yaml:"Redis"   json:"redis"`
	RedisP  RedisConfig   `yaml:"RedisP"  json:"redis_p"`
	Dynamic DynamicConfig `yaml:"Dynamic" json:"dynamic"`
	NoProxy NoProxyConfig `yaml:"NoProxy" json:"no_proxy"`
}

// RedisConfig holds Redis connection parameters.
type RedisConfig struct {
	Address        []string `yaml:"Address"        json:"address"`
	Bid            string   `yaml:"Bid"            json:"bid"`
	Prefix         string   `yaml:"Prefix"         json:"prefix"`
	DB             int      `yaml:"DB"             json:"db"`
	Cluster        bool     `yaml:"Cluster"        json:"cluster"`
	Expire         int      `yaml:"Expire"         json:"expire"`
	ReadTimeout    int      `yaml:"ReadTimeout"    json:"read_timeout"`
	WriteTimeout   int      `yaml:"WriteTimeout"   json:"write_timeout"`
	ConnectTimeout int      `yaml:"ConnectTimeout" json:"connect_timeout"`
	PoolMaxActive  int      `yaml:"PoolMaxActive"  json:"pool_max_active"`
	PoolMinActive  int      `yaml:"PoolMinActive"  json:"pool_min_active"`
	PoolIdleTimeout int     `yaml:"PoolIdleTimeout" json:"pool_idle_timeout"`
	Password       string   `yaml:"Password"       json:"password"`
}

// DynamicConfig holds dynamic-proxy configuration.
type DynamicConfig struct {
	IsDynamic      bool   `yaml:"IsDynamic"      json:"is_dynamic"`
	DynamicAddress string `yaml:"DynamicAddress" json:"dynamic_address"`
	CycleTime      int    `yaml:"CycleTime"      json:"cycle_time"`
	CycleFlagTime  int    `yaml:"CycleFlagTime"  json:"cycle_flag_time"`
}

// NoProxyConfig holds no-proxy mode configuration.
type NoProxyConfig struct {
	IsNoProxy      bool   `yaml:"IsNoProxy"      json:"is_no_proxy"`
	ServiceAddress string `yaml:"ServiceAddress" json:"service_address"`
	MonitorAddress string `yaml:"MonitorAddress" json:"monitor_address"`
	Lang           string `yaml:"Lang"           json:"lang"`
	AppCode        string `yaml:"AppCode"        json:"app_code"`
	Iv             string `yaml:"Iv"             json:"iv"`
	MonitorTime    int    `yaml:"MonitorTime"    json:"monitor_time"`
}

// ESConfig holds Elasticsearch configuration.
type ESConfig struct {
	Address          []string `yaml:"Address"          json:"address"`
	TransportMaxIdel int      `yaml:"TransportMaxIdel" json:"transport_max_idel"`
	Timeout          int      `yaml:"Timeout"          json:"timeout"`
}

// TracerConfig holds SkyWalking tracer configuration.
type TracerConfig struct {
	ServiceName string `yaml:"ServiceName" json:"service_name"`
	Disabled    bool   `yaml:"Disabled"    json:"disabled"`
	Reporter    struct {
		LocalAgentHostPort string `yaml:"LocalAgentHostPort" json:"local_agent_host_port"`
	} `yaml:"Reporter" json:"reporter"`
	Sampler struct {
		SamplingRate float64 `yaml:"SamplingRate" json:"sampling_rate"`
	} `yaml:"Sampler" json:"sampler"`
}

// LogConfig holds logger bootstrap parameters.
type LogConfig struct {
	Name       string `yaml:"Name"       json:"name"`
	Path       string `yaml:"Path"       json:"path"`
	Debug      bool   `yaml:"Debug"      json:"debug"`
	MaxSize    int    `yaml:"MaxSize"    json:"max_size"`
	MaxAge     int    `yaml:"MaxAge"     json:"max_age"`
	MaxBackups int    `yaml:"MaxBackups" json:"max_backups"`
	Compress   bool   `yaml:"Compress"   json:"compress"`
}

// ProjectConfig is the top-level framework bootstrap config (project.yaml).
type ProjectConfig struct {
	Base struct {
		Mode        string `yaml:"mode"         json:"mode"`
		Format      string `yaml:"format"       json:"format"`
		ConfigsPath string `yaml:"configs_path" json:"configs_path"`
	} `yaml:"base" json:"base"`
	Nacos struct {
		Addr        string   `yaml:"addr"         json:"addr"`
		Port        uint64   `yaml:"port"         json:"port"`
		NamespaceID string   `yaml:"namespace_id" json:"namespace_id"`
		Group       string   `yaml:"group"        json:"group"`
		DataID      []string `yaml:"data_id"      json:"data_id"`
		TimeoutMs   uint64   `yaml:"timeout_ms"   json:"timeout_ms"`
		LogLevel    string   `yaml:"log_level"    json:"log_level"`
		Username    string   `yaml:"username"     json:"username"`
		Password    string   `yaml:"password"     json:"password"`
		Interval    int      `yaml:"interval"     json:"interval"`
	} `yaml:"nacos" json:"nacos"`
	Log LogConfig `yaml:"log" json:"log"`
}
