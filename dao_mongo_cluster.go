package gozen

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"strings"
	"sync"
	"time"
)

/**

读取模式字段：
	Primary 只读主节点 数据强一致性
	PrimaryPreferred 主节点优先 主节点挂后 请求备份节点 进入只读模式 依然可读
	Secondary 只读备份节点 数据一致性要求低
	SecondaryPreferred 备份节点优先 备份节点挂掉 请求主节点
	Nearest 请求低延迟数据
	Eventual is same as Nearest, but may change servers between reads
	Monotonic is same as SecondaryPreferred before first write. Same as Primary after first write
	Strong is same as Primary


*/

type DaoMongodbCluster struct {
	DbExtName    string //库扩展
	TableExtName string //表扩展
	DbSelector   uint32 //库号
	DaoMongo
}

func NewDaoMongodbCluster() *DaoMongodbCluster {
	return &DaoMongodbCluster{}
}

var (
	sessionMongodbCluster sync.Map //*mongo.Client
)

func initMongodbClusterSession() {
	mongodbClusterConfigMap.Range(func(k, v interface{}) bool {
		dbNum := k.(uint32)
		config := v.(*ConfigMongo)
		if config.Servers == "" || config.DbName == "" { //集群mongodb配置不可错
			panic(fmt.Sprintf("mongodb cluster config error - config:%v", config))
			return false
		}
		if strings.Trim(config.ReadOption, " ") == "" {
			config.ReadOption = "nearest"
		}
		var connectionString string
		if config.User != "" && config.Password != "" {
			connectionString = fmt.Sprintf("mongodb://%s:%s@%s/%s?%s", config.User, config.Password,
				config.Servers, config.DbName, config.Options)
		} else {
			connectionString = fmt.Sprintf("mongodb://%s/%s?%s", config.Servers, config.DbName, config.Options)
		}
		//mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]
		mongoOptions := options.Client().ApplyURI(connectionString)
		mongoOptions.SetMaxPoolSize(config.MaxPoolSize)
		mongoOptions.SetMinPoolSize(config.MinPoolSize)
		mongoOptions.SetSocketTimeout(time.Duration(config.SocketTimeout) * time.Second)
		mongoOptions.SetConnectTimeout(time.Duration(config.ConnectTimeout) * time.Second)
		mongoOptions.SetMaxConnIdleTime(time.Duration(config.MaxConnIdleTime) * time.Second)
		mongoOptions.SetServerSelectionTimeout(time.Duration(config.ServerSelectionTimeout) * time.Second)
		sessionMongodbSingle, err := mongo.NewClient(mongoOptions)
		if err != nil {
			panic(fmt.Sprintf("NewClient mongo server error:%v,%s", err, connectionString))
		}
		err = sessionMongodbSingle.Connect(context.Background())
		if err != nil {
			panic(fmt.Sprintf("connect to mongo server error:%v,%s", err, connectionString))
		}
		sessionMongodbCluster.Store(dbNum, sessionMongodbSingle)
		return true
	})
}

func getMongodbClusterSessionOne(dbNum uint32) (data *mongo.Client, ok bool) {
	sessionMongodbCluster, ok := sessionMongodbCluster.Load(dbNum)
	if !ok {
		return
	}
	data = sessionMongodbCluster.(*mongo.Client)
	return
}

func (p *DaoMongodbCluster) GetSession() (session mongo.Session, dbName string, timeout int, err error) {
	configCluster, ok := ConfigMongodbClusterGetOne(p.DbSelector)
	if !ok {
		err = errors.New(fmt.Sprintf("mongodb config null dbNum:%d", p.DbSelector))
		return
	}
	sessionCluster, ok := getMongodbClusterSessionOne(p.DbSelector)
	if !ok {
		err = errors.New(fmt.Sprintf("mongodb session null dbNum:%d", p.DbSelector))
		return
	}
	var opts []*options.SessionOptions
	if p.Mode == "" {
		p.Mode = configCluster.ReadOption
	}
	rpMod, err := readpref.ModeFromString(p.Mode)
	if err != nil {
		err = p.processError(err, "mongo GetSession readpref.ModeFromString:%s", err.Error())
		return
	}
	rp, err := readpref.New(rpMod)
	if err != nil {
		err = p.processError(err, "mongo GetSession readpref.New:%s", err.Error())
		return
	}
	opts = append(opts, options.Session().SetDefaultReadPreference(rp))
	session, err = sessionCluster.StartSession(opts...)
	if p.DbExtName == "" {
		dbName = configCluster.DbName
	} else {
		dbName = configCluster.DbName + "_" + p.DbExtName
	}
	timeout = configCluster.Timeout
	return
}

func (p *DaoMongodbCluster) GetId() (int64, error) {
	return p.GetNextSequence()
}

func (p *DaoMongodbCluster) GetNextSequence() (int64, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return 0, err
	}
	defer session.EndSession(p.Ctx)
	c := session.Client().Database(dbName).Collection("counters")
	condition := bson.M{"_id": p.CollectionName}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	result := struct {
		ID  string `bson:"_id"`
		Seq int64  `bson:"seq"`
	}{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	opts := new(options.FindOneAndUpdateOptions)
	opts.SetUpsert(true)
	opts.SetReturnDocument(options.After)
	errApply := c.FindOneAndUpdate(ctx, condition, update, opts).Decode(&result)
	SpanErrorFast(span, errApply)
	if errApply != nil {
		errApply = p.processError(errApply, "mongo FindOneAndUpdate counter %s failed:%s", p.CollectionName, errApply.Error())
		return 0, errApply
	}
	return result.Seq, nil
}

func (p *DaoMongodbCluster) GetById(id interface{}, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	errFind := session.Client().Database(dbName).Collection(p.CollectionName).FindOne(ctx, bson.M{"_id": id}).Decode(data)
	SpanErrorFast(span, errFind)
	if errFind != nil {
		errFind = p.processError(errFind, "mongo %s get id failed:%v", p.CollectionName, errFind.Error())
	}
	return errFind
}

func (p *DaoMongodbCluster) Insert(data IModelMongo) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	if !data.ExistId() {
		if p.AutoIncrementId {
			id, err := p.GetNextSequence()
			if err != nil {
				return err
			}
			data.SetId(id)
		} else {
			data.SetObjectId()
		}
	}
	// 是否初始化时间
	createdAt := data.GetCreatedTime()
	if createdAt.Equal(time.Time{}) {
		data.InitTime(time.Now())
	}
	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	coll := session.Client().Database(dbName).Collection(p.CollectionName)
	_, errInsert := coll.InsertOne(ctx, data)
	SpanErrorFast(span, errInsert)
	if errInsert != nil {
		errInsert = p.processError(errInsert, "mongo %s insert failed:%v", p.CollectionName, errInsert.Error())
		return errInsert
	}
	return nil
}

func (p *DaoMongodbCluster) InsertReturn(data IModelMongo) (insertedID interface{}, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	if !data.ExistId() {
		if p.AutoIncrementId {
			var id int64
			id, err = p.GetNextSequence()
			if err != nil {
				return
			}
			data.SetId(id)
		} else {
			data.SetObjectId()
		}
	}
	// 是否初始化时间
	createdAt := data.GetCreatedTime()
	if createdAt.Equal(time.Time{}) {
		data.InitTime(time.Now())
	}
	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	coll := session.Client().Database(dbName).Collection(p.CollectionName)
	var ret *mongo.InsertOneResult
	ret, err = coll.InsertOne(ctx, data)
	SpanErrorFast(span, err)
	if err != nil {
		err = p.processError(err, "mongo %s insert failed:%v", p.CollectionName, err.Error())
		return
	}
	insertedID = ret.InsertedID
	return
}

func (p *DaoMongodbCluster) InsertM(data []IModelMongo) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	for _, item := range data {
		if !item.ExistId() {
			if p.AutoIncrementId {
				id, err := p.GetNextSequence()
				if err != nil {
					return err
				}
				item.SetId(id)
			} else {
				item.SetObjectId()
			}
		}
		// 是否初始化时间
		createdAt := item.GetCreatedTime()
		if createdAt.Equal(time.Time{}) {
			item.InitTime(time.Now())
		}
	}
	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	coll := session.Client().Database(dbName).Collection(p.CollectionName)
	var idata []interface{}
	for i := 0; i < len(data); i++ {
		idata = append(idata, data[i])
	}
	_, errInsert := coll.InsertMany(ctx, idata)
	SpanErrorFast(span, errInsert)
	if errInsert != nil {
		errInsert = p.processError(errInsert, "mongo %s insertM failed:%v", p.CollectionName, errInsert.Error())
		return errInsert
	}
	return nil
}

func (p *DaoMongodbCluster) InsertMReturn(data []IModelMongo) (insertedIDs []interface{}, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	for _, item := range data {
		if !item.ExistId() {
			if p.AutoIncrementId {
				var id int64
				id, err = p.GetNextSequence()
				if err != nil {
					return
				}
				item.SetId(id)
			} else {
				item.SetObjectId()
			}
		}
		// 是否初始化时间
		createdAt := item.GetCreatedTime()
		if createdAt.Equal(time.Time{}) {
			item.InitTime(time.Now())
		}
	}
	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	coll := session.Client().Database(dbName).Collection(p.CollectionName)
	var idata []interface{}
	for i := 0; i < len(data); i++ {
		idata = append(idata, data[i])
	}
	var ret *mongo.InsertManyResult
	ret, err = coll.InsertMany(ctx, idata)
	SpanErrorFast(span, err)
	if err != nil {
		err = p.processError(err, "mongo %s insertM failed:%v", p.CollectionName, err.Error())
		return
	}
	insertedIDs = ret.InsertedIDs
	return
}

func (p *DaoMongodbCluster) Count(condition interface{}) (int, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return 0, err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	count, errCount := session.Client().Database(dbName).Collection(p.CollectionName).CountDocuments(ctx, condition)
	SpanErrorFast(span, errCount)
	if errCount != nil {
		errCount = p.processError(errCount, "mongo %s count failed:%v", p.CollectionName, errCount.Error())
	}
	return int(count), errCount
}

func (p *DaoMongodbCluster) Find(condition interface{}, limit int, skip int, data interface{}, sortFields bson.D) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	var opts []*options.FindOptions
	if len(sortFields) == 0 {
		sortFields = append(sortFields, bson.E{Key: "_id", Value: -1}) // id生成倒序 即时间倒序
	}
	opts = append(opts, options.Find().SetSort(sortFields))
	if skip > 0 {
		opts = append(opts, options.Find().SetSkip(int64(skip)))
	}
	if limit > 0 {
		opts = append(opts, options.Find().SetLimit(int64(limit)))
		opts = append(opts, options.Find().SetBatchSize(int32(limit)))
	}
	c, err := session.Client().Database(dbName).Collection(p.CollectionName).Find(ctx, condition, opts...)
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	errSelect := c.All(ctx, data)
	if errSelect != nil {
		errSelect = p.processError(errSelect, "mongo %s find failed:%v", p.CollectionName, errSelect.Error())
	}
	SpanErrorFast(span, errSelect)
	return errSelect
}

// Distinct 只返回field的字段
func (p *DaoMongodbCluster) Distinct(condition interface{}, field string) (data []interface{}, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	data, err = session.Client().Database(dbName).Collection(p.CollectionName).Distinct(ctx, field, condition)
	SpanErrorFast(span, err)
	if err != nil {
		err = p.processError(err, "mongo %s distinct failed:%s", p.CollectionName, err.Error())
	}
	return
}

func (p *DaoMongodbCluster) Sum(condition interface{}, sumField string) (int, error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return 0, err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	coll := session.Client().Database(dbName).Collection(p.CollectionName)
	pipe := mongo.Pipeline{
		{{"$match", condition}},
		{{"$group", bson.D{{"_id", nil}, {"sum", bson.D{{"$sum", "$" + sumField}}}}}},
	}
	type SumStruct struct {
		Sum int `bson:"sum"`
	}
	var result []SumStruct
	c, err := coll.Aggregate(ctx, pipe)
	SpanErrorFast(span, err)
	if err != nil {
		return 0, err
	}
	err = c.All(ctx, &result)
	SpanErrorFast(span, err)
	if err != nil {
		err = p.processError(err, "mongo %s sum failed: %s", p.CollectionName, err.Error())
		return 0, err
	}
	if len(result) == 1 {
		return result[0].Sum, nil
	}
	return 0, nil
}

func (p *DaoMongodbCluster) DistinctCount(condition interface{}, field string) (data []DaoMongoCountStruct, err error) {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	coll := session.Client().Database(dbName).Collection(p.CollectionName)

	pipe := mongo.Pipeline{
		{{"$match", condition}},
		{{"$group", bson.D{{"_id", fmt.Sprintf("$%s", field)}, {"count", bson.D{{"$sum", 1}}}}}},
	}
	c, err := coll.Aggregate(ctx, pipe)
	SpanErrorFast(span, err)
	if err != nil {
		return
	}
	err = c.All(ctx, &data)
	SpanErrorFast(span, err)
	if err != nil {
		err = p.processError(err, "mongo %s DistinctCount failed: %s", p.CollectionName, err.Error())
		return
	}
	return
}

func (p *DaoMongodbCluster) Update(condition interface{}, data map[string]interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	coll := session.Client().Database(dbName).Collection(p.CollectionName)
	setBson := bson.M{}
	for key, value := range data {
		setBson[fmt.Sprintf("%s", key)] = value
	}
	updateData := bson.M{"$set": setBson, "$currentDate": bson.M{"updated_at": true}}
	_, errUpdate := coll.UpdateMany(ctx, condition, updateData)
	SpanErrorFast(span, errUpdate)
	if errUpdate != nil {
		errUpdate = p.processError(errUpdate, "mongo %s update failed: %s", p.CollectionName, errUpdate.Error())
	}
	return errUpdate
}

func (p *DaoMongodbCluster) Upsert(condition interface{}, data map[string]interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	coll := session.Client().Database(dbName).Collection(p.CollectionName)
	updateData := bson.M{
		"$set":         data,
		"$setOnInsert": bson.M{"created_at": time.Now()},
		"$currentDate": bson.M{"updated_at": true},
	}
	opts := options.Update().SetUpsert(true)
	_, errUpsert := coll.UpdateMany(ctx, condition, updateData, opts)
	SpanErrorFast(span, errUpsert)
	if errUpsert != nil {
		errUpsert = p.processError(errUpsert, "mongo %s errUpsert failed: %s", p.CollectionName, errUpsert.Error())
	}
	return errUpsert
}

func (p *DaoMongodbCluster) UpsertNum(condition interface{}, data map[string]interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	coll := session.Client().Database(dbName).Collection(p.CollectionName)
	updateData := bson.M{"$inc": data, "$currentDate": bson.M{"updated_at": true}}
	opts := options.Update().SetUpsert(true)
	_, errUpsert := coll.UpdateMany(ctx, condition, updateData, opts)
	SpanErrorFast(span, errUpsert)
	if errUpsert != nil {
		errUpsert = p.processError(errUpsert, "mongo %s errUpsert failed: %s", p.CollectionName, errUpsert.Error())
	}
	return errUpsert
}

func (p *DaoMongodbCluster) RemoveId(id interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	coll := session.Client().Database(dbName).Collection(p.CollectionName)
	filter := bson.M{"_id": id}
	_, errRemove := coll.DeleteOne(ctx, filter)
	SpanErrorFast(span, errRemove)
	if errRemove != nil {
		errRemove = p.processError(errRemove, "mongo %s removeId failed: %s, id:%v", p.CollectionName, errRemove.Error(), id)
	}
	return errRemove
}

func (p *DaoMongodbCluster) RemoveAll(selector interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	coll := session.Client().Database(dbName).Collection(p.CollectionName)
	_, errRemove := coll.DeleteMany(ctx, selector)
	SpanErrorFast(span, errRemove)
	if errRemove != nil {
		errRemove = p.processError(errRemove, "mongo %s removeAll failed: %s, selector:%v", p.CollectionName, errRemove.Error(), selector)
	}
	return errRemove
}

func (p *DaoMongodbCluster) UpdateOne(condition interface{}, data map[string]interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	coll := session.Client().Database(dbName).Collection(p.CollectionName)
	setBson := bson.M{}
	for key, value := range data {
		setBson[fmt.Sprintf("%s", key)] = value
	}
	updateData := bson.M{"$set": setBson, "$currentDate": bson.M{"updated_at": true}}
	_, errUpdate := coll.UpdateOne(ctx, condition, updateData)
	SpanErrorFast(span, errUpdate)
	if errUpdate != nil {
		errUpdate = p.processError(errUpdate, "mongo %s update one failed: %s", p.CollectionName, errUpdate.Error())
	}
	return errUpdate
}

func (p *DaoMongodbCluster) processError(err error, formatter string, a ...interface{}) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	str := fmt.Sprintf(formatter, p.CollectionName, a)
	LogErrorw(LogNameMongodb, str)
	return err
}

func (p *DaoMongodbCluster) FindOne(condition interface{}, data interface{}, sortFields bson.D) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoMongoDb, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)

	session, dbName, timeout, err := p.GetSession()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer session.EndSession(p.Ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	var opts []*options.FindOneOptions
	if len(sortFields) == 0 {
		sortFields = append(sortFields, bson.E{Key: "_id", Value: -1}) // id生成倒序 即时间倒序
	}
	opts = append(opts, options.FindOne().SetSort(sortFields))
	err = session.Client().Database(dbName).Collection(p.CollectionName).FindOne(ctx, condition, opts...).Decode(data)
	if err != nil {
		err = p.processError(err, "mongo %s findOne failed: %s", p.CollectionName, err.Error())
	}
	SpanErrorFast(span, err)
	return err
}
