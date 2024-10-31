package gozen

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

type ModelMongoHello struct {
	ModelMongo `bson:",inline"`
	HelloWord  string `bson:"hello_word"`
	Name       string `bson:"name"`
}

type ModelMongoHelloObj struct {
	ModelMongoBase `bson:",inline"`
	HelloWord      string `bson:"hello_word"`
	Name           string `bson:"name"`
}

type TestDaoMongo struct {
	DaoMongo
}

func NewMongoTest() *TestDaoMongo {
	initMongodbSession()
	dao := &TestDaoMongo{}
	dao.AutoIncrementId = true
	dao.CollectionName = "test-mongo"
	dao.Mode = "primary"
	return dao
}

func NewMongoTestObj() *TestDaoMongo {
	initMongodbSession()
	dao := &TestDaoMongo{}
	dao.AutoIncrementId = false
	dao.CollectionName = "test-mongo-obj"
	dao.Mode = "primary"
	return dao
}

func Test_MongoInsertOne(t *testing.T) {
	mongo := NewMongoTest()
	for i := 1; i <= 10; i++ {
		model := &ModelMongoHello{}
		model.HelloWord = "1"
		model.Name = "name2"
		err := mongo.Insert(model)
		if err != nil {
			t.Errorf("insert error:%s", err.Error())
		}
	}
}

func Test_MongoCount(t *testing.T) {
	mongo := NewMongoTest()
	condition := bson.M{"name": "name3"}
	ret, err := mongo.Count(condition)
	if err != nil {
		t.Errorf("insert error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoDistinct(t *testing.T) {
	mongo := NewMongoTest()
	condition := bson.M{}
	ret, err := mongo.Distinct(condition, "name")
	if err != nil {
		t.Errorf("Distinct error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoSum(t *testing.T) {
	mongo := NewMongoTest()
	condition := bson.M{"name": "name3"}
	ret, err := mongo.Sum(condition, "_id")
	if err != nil {
		t.Errorf("Test_MongoSum error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoDistinctNameCount(t *testing.T) {
	mongo := NewMongoTest()
	condition := bson.M{}
	ret, err := mongo.DistinctCount(condition, "name")
	if err != nil {
		t.Errorf("Test_MongoDistinctCount error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoDistinctIdCount(t *testing.T) {
	mongo := NewMongoTest()
	condition := bson.M{}
	ret, err := mongo.DistinctCount(condition, "_id")
	if err != nil {
		t.Errorf("Test_MongoDistinctCount error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoDistinctObjIdCount(t *testing.T) {
	mongo := NewMongoTestObj()
	condition := bson.M{}
	ret, err := mongo.DistinctCount(condition, "_id")
	if err != nil {
		t.Errorf("Test_MongoDistinctCount error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoUpdate(t *testing.T) {
	mongo := NewMongoTest()
	condition := bson.M{
		"_id": 44,
	}
	updateData := map[string]interface{}{
		"name": "贾龙天:" + time.Now().String(),
	}
	err := mongo.Update(condition, updateData)
	if err != nil {
		t.Errorf("Test_MongoUpdate error:%s", err.Error())
	} else {
		var data ModelMongoHello
		err = mongo.FindOne(condition, &data, bson.D{})
		if err != nil {
			t.Errorf("FindOne error:%s", err.Error())
		}
		t.Log("data", data)
	}
}

func Test_MongoUpdateOne(t *testing.T) {
	mongo := NewMongoTest()
	condition := bson.M{
		"_id": 1,
	}
	updateData := map[string]interface{}{
		"name": "贾龙天222:" + time.Now().String(),
	}
	err := mongo.UpdateOne(condition, updateData)
	if err != nil {
		t.Errorf("Test_MongoUpdate error:%s", err.Error())
	} else {
		var data ModelMongoHello
		err = mongo.FindOne(condition, &data, bson.D{})
		if err != nil {
			t.Errorf("FindOne error:%s", err.Error())
		}
		t.Log("data", data)
	}
}

func Test_MongoUpsert(t *testing.T) {
	mongo := NewMongoTest()
	condition := bson.M{
		"_id": int64(2),
	}
	updateData := map[string]interface{}{
		"name":       "贾龙天name:" + time.Now().String(),
		"hello_word": "贾龙天hello:" + time.Now().String(),
	}
	err := mongo.Upsert(condition, updateData)
	if err != nil {
		t.Errorf("Test_MongoUpsert error:%s", err.Error())
	} else {
		var data ModelMongoHello
		err = mongo.FindOne(condition, &data, bson.D{})
		if err != nil {
			t.Errorf("FindOne error:%s", err.Error())
		}
		t.Log("data", data)
	}
}

func Test_MongoUpsertNum(t *testing.T) {
	mongo := NewMongoTest()
	condition := bson.M{
		"_id": int64(2),
	}
	updateData := map[string]interface{}{
		"num": 1,
	}
	err := mongo.UpsertNum(condition, updateData)
	if err != nil {
		t.Errorf("Test_MongoUpsertNum error:%s", err.Error())
	} else {
		var data ModelMongoHello
		err = mongo.FindOne(condition, &data, bson.D{})
		if err != nil {
			t.Errorf("FindOne error:%s", err.Error())
		}
		t.Log("data", data)
	}
}

func Test_MongoRemoveId(t *testing.T) {
	mongo := NewMongoTest()
	err := mongo.RemoveId(2)
	if err != nil {
		t.Errorf("Test_MongoUpsertNum error:%s", err.Error())
	}
}

func Test_MongoRemoveAll(t *testing.T) {
	mongo := NewMongoTest()
	condition := bson.M{
		"_id": int64(22),
	}
	err := mongo.RemoveAll(condition)
	if err != nil {
		t.Errorf("Test_MongoUpsertNum error:%s", err.Error())
	}
}

func Test_MongoInsertM(t *testing.T) {
	mongo := NewMongoTest()
	var modelList []IModelMongo
	for i := 1; i <= 10; i++ {
		model := &ModelMongoHello{}
		model.HelloWord = "1"
		model.Name = "name2insertM"
		modelList = append(modelList, model)
	}
	err := mongo.InsertM(modelList)
	if err != nil {
		t.Errorf("insert error:%s", err.Error())
	}
}

func Test_MongoInsertMReturnInt(t *testing.T) {
	mongo := NewMongoTest()
	var modelList []IModelMongo
	for i := 1; i <= 10; i++ {
		model := &ModelMongoHello{}
		model.HelloWord = "1"
		model.Name = "name2insertM"
		modelList = append(modelList, model)
	}
	ids, err := mongo.InsertMReturn(modelList)
	if err != nil {
		t.Errorf("insert error:%s", err.Error())
	} else {
		for _, v := range ids {
			t.Log("id", v, v.(int64))
		}
	}
}

func Test_MongoInsertMReturnObj(t *testing.T) {
	mongo := NewMongoTestObj()
	var modelList []IModelMongo
	for i := 1; i <= 10; i++ {
		model := &ModelMongoHelloObj{}
		model.HelloWord = "1"
		model.Name = "name2insertM"
		modelList = append(modelList, model)
	}
	ids, err := mongo.InsertMReturn(modelList)
	if err != nil {
		t.Errorf("insert error:%s", err.Error())
	} else {
		for _, v := range ids {
			t.Log("id", v, v.(primitive.ObjectID))
		}
	}
}

func Test_MongoInsertReturnInt(t *testing.T) {
	mongo := NewMongoTest()
	for i := 1; i <= 10; i++ {
		model := &ModelMongoHello{}
		model.HelloWord = "1"
		model.Name = "name2"
		r, err := mongo.InsertReturn(model)
		if err != nil {
			t.Errorf("insert error:%s", err.Error())
		} else {
			t.Log("ret", r, r.(int64))
		}
	}
}

func Test_MongoInsertReturnObj(t *testing.T) {
	mongo := NewMongoTestObj()
	for i := 1; i <= 10; i++ {
		model := &ModelMongoHelloObj{}
		model.HelloWord = "1"
		model.Name = "name2"
		r, err := mongo.InsertReturn(model)
		if err != nil {
			t.Errorf("insert error:%s", err.Error())
		} else {
			t.Log("ret", r, r.(primitive.ObjectID))
		}
	}
}

func Test_MongoFindList(t *testing.T) {
	mongo := NewMongoTest()
	condition := bson.M{
		"_id": bson.M{"$lt": 43},
	}
	limit := 2
	skip := 0
	var sortFields bson.D
	var data []ModelMongoHello
	err := mongo.Find(condition, limit, skip, &data, sortFields)
	if err != nil {
		t.Errorf("mongo find error,%s", err.Error())
	} else {
		t.Logf("data len is %v is %v", len(data), data)
	}
}

func Test_MongoFindOne(t *testing.T) {
	mongo := NewMongoTest()
	condition := bson.M{
		"_id": bson.M{"$lt": 43},
	}
	var sortFields bson.D
	var data ModelMongoHello
	err := mongo.FindOne(condition, &data, sortFields)
	if err != nil {
		t.Errorf("mongo find error,%s", err.Error())
	} else {
		t.Logf("data  is %v", data)
	}
}

func Test_MongoGetById(t *testing.T) {
	var (
		helloWord = "djaskdjaskldja"
		name      = "37128371289371289"
	)

	mongo := NewMongoTestObj()
	// insert one data
	model := &ModelMongoHelloObj{}
	model.HelloWord = helloWord
	model.Name = name

	err := mongo.Insert(model)
	if err != nil {
		t.Error("Test_MongoGetById Insert error", err)
		return
	}

	// GetById查询
	var queryData ModelMongoHelloObj
	err = mongo.GetById(model.Id, &queryData)
	if err != nil {
		t.Error("Test_MongoGetById GetById error", err)
		return
	}
	if queryData.Name != name || queryData.HelloWord != helloWord {
		t.Error("Test_MongoGetById failed")
		return
	}
}

func BenchmarkMongoFind(b *testing.B) {

	/**

	go test -test.bench BenchmarkMongoFind -test.run BenchmarkMongoFind -benchtime=5s -cpuprofile cpu.out -memprofile mem5s.out
	go test -test.bench BenchmarkMongoFind -test.run BenchmarkMongoFind -benchtime=10s  -cpuprofile cpu.out -memprofile mem10s.out
	go test -test.bench BenchmarkMongoFind -test.run BenchmarkMongoFind -benchtime=20s  -cpuprofile cpu.out -memprofile mem20s.out
	go test -test.bench BenchmarkMongoFind -test.run BenchmarkMongoFind -benchtime=30s  -cpuprofile cpu.out -memprofile mem30s.out

	go test -test.bench BenchmarkMongoFind -test.run BenchmarkMongoFind -benchtime=5s -cpuprofile cpu.out -memprofile mem5sn.out
	go test -test.bench BenchmarkMongoFind -test.run BenchmarkMongoFind -benchtime=10s  -cpuprofile cpu.out -memprofile mem10sn.out
	go test -test.bench BenchmarkMongoFind -test.run BenchmarkMongoFind -benchtime=20s  -cpuprofile cpu.out -memprofile mem20sn.out


	*/
	mongo := NewMongoTest()
	condition := bson.M{
		"_id": bson.M{"$lt": 43},
	}
	limit := 2
	skip := 0
	var sortFields bson.D
	var data []ModelMongoHello
	var (
		helloWord = "djaskdjaskldja"
		name      = "37128371289371289"
	)

	for i := 0; i < b.N; i++ {
		err := mongo.Find(condition, limit, skip, &data, sortFields)
		if err != nil {
			b.Errorf("BenchmarkMongoFind find error,%s", err.Error())
		}

		model := &ModelMongoHelloObj{}
		model.HelloWord = helloWord
		model.Name = name

		err = mongo.Insert(model)
		if err != nil {
			b.Error("BenchmarkMongoFind Insert error", err)
		}
	}
}
