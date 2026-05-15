package log

// Logger name constants — used to route log entries to distinct output files
// or to add a "component" tag.
const (
	NameDefault = "default"
	NameRedis   = "redis"
	NameMysql   = "mysql"
	NameMongodb = "mongodb"
	NameApi     = "api"
	NameAo      = "ao"
	NameGRpc    = "grpc"
	NameEs      = "es"
	NameTmq     = "tmq"
	NameAmq     = "amq"
	NameLogic   = "logic"
	NameFile    = "file"
	NameNet     = "net"
)

// Standard structured-logging field keys.  Use these constants as the keys in
// sugared log calls (Infow, Errorw, etc.) so that downstream consumers
// (Elasticsearch, Splunk, etc.) see a consistent schema.
const (
	KNameCommonErr       = "log-common-err"
	KNameCommonFields    = "log-common-fields"
	KNameCommonCondition = "log-common-condition"
	KNameCommonAddress   = "log-common-address"
	KNameCommonName      = "log-common-name"
	KNameCommonCmd       = "log-common-cmd"
	KNameCommonData      = "log-common-data"
	KNameCommonDataType  = "log-common-data-type"
	KNameCommonKey       = "log-common-key"
	KNameCommonValue     = "log-common-value"
	KNameCommonUrl       = "log-common-url"
	KNameCommonNum       = "log-common-num"
	KNameCommonId        = "log-common-id"
	KNameCommonUid       = "log-common-uid"
	KNameCommonCode      = "log-common-code"
	KNameCommonLevel     = "log-common-level"
	KNameCommonCookie    = "log-common-cookie"
	KNameCommonReq       = "log-common-req"
	KNameCommonRes       = "log-common-res"
	KNameCommonTime      = "log-common-time"
	KNameCommonTenantId  = "log-common-tenant-id"
	KNameCommonRecordId  = "log-common-record-id"
	KNameCommonUniqueId  = "log-common-unique-id"

	KNameRedisKey  = "log-redis-key"
	KNameRedisData = "log-redis-data"

	KNameMysqlParam = "log-mysql-param"
	KNameMysqlData  = "log-mysql-data"

	KNameMongodbParam = "log-mongodb-param"
	KNameMongodbData  = "log-mongodb-data"

	KNameApiParam = "log-api-param"
	KNameApiUrl   = "log-api-url"
	KNameApiRes   = "log-api-res"

	KNameAoParam = "log-ao-param"
	KNameAoReq   = "log-ao-req"
	KNameAoRes   = "log-ao-res"

	KNameGRpcReq = "log-grpc-req"
	KNameGRpcRes = "log-grpc-res"

	KNameEsReq = "log-es-req"
	KNameEsRes = "log-es-res"

	KNameTmqTopic = "log-tmq-topic"
	KNameTmqReq   = "log-tmq-req"
	KNameTmqRes   = "log-tmq-res"
)

// logKNameSet contains all KName constants. Keys that are NOT in this set are
// passed through as-is in ZapLogger.sugarLogParamToStr.
var logKNameSet = map[any]struct{}{
	KNameCommonErr:       {},
	KNameCommonFields:    {},
	KNameCommonCondition: {},
	KNameCommonAddress:   {},
	KNameCommonName:      {},
	KNameCommonData:      {},
	KNameCommonDataType:  {},
	KNameCommonKey:       {},
	KNameCommonValue:     {},
	KNameCommonUrl:       {},
	KNameCommonNum:       {},
	KNameCommonId:        {},
	KNameCommonUid:       {},
	KNameCommonCode:      {},
	KNameCommonLevel:     {},
	KNameCommonCookie:    {},
	KNameCommonReq:       {},
	KNameCommonRes:       {},
	KNameCommonTime:      {},
	KNameCommonTenantId:  {},
	KNameCommonRecordId:  {},
	KNameCommonUniqueId:  {},
	KNameRedisKey:        {},
	KNameRedisData:       {},
	KNameMysqlParam:      {},
	KNameMysqlData:       {},
	KNameMongodbParam:    {},
	KNameMongodbData:     {},
	KNameApiParam:        {},
	KNameApiUrl:          {},
	KNameApiRes:          {},
	KNameAoParam:         {},
	KNameAoReq:           {},
	KNameAoRes:           {},
	KNameGRpcReq:         {},
	KNameGRpcRes:         {},
	KNameEsReq:           {},
	KNameEsRes:           {},
	KNameTmqTopic:        {},
	KNameTmqReq:          {},
	KNameTmqRes:          {},
}
