package gozen

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

type TestDaoMongoCluster struct {
	DaoMongodbCluster
}

func NewMongoClusterTest() *TestDaoMongoCluster {
	initMongodbClusterSession()
	dao := &TestDaoMongoCluster{DaoMongodbCluster{
		DbSelector: 2001,
	}}
	dao.AutoIncrementId = true
	dao.CollectionName = "test_cluster"
	return dao
}

func NewMongoClusterTestObj() *TestDaoMongoCluster {
	initMongodbClusterSession()
	dao := &TestDaoMongoCluster{DaoMongodbCluster{
		DbSelector: 2001,
	}}
	dao.CollectionName = "test_cluster_obj"
	return dao
}

func Test_MongoClusterInsertOne(t *testing.T) {
	mongo := NewMongoClusterTest()
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

func Test_MongoClusterCount(t *testing.T) {
	mongo := NewMongoClusterTest()
	condition := bson.M{"name": "name3"}
	ret, err := mongo.Count(condition)
	if err != nil {
		t.Errorf("insert error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoClusterCountNull(t *testing.T) {
	mongo := NewMongoClusterTest()
	mongo.CollectionName = "thisisnull"
	condition := bson.M{"name": "name3"}
	ret, err := mongo.Count(condition)
	if err != nil {
		t.Errorf("count error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoClusterDistinct(t *testing.T) {
	mongo := NewMongoClusterTest()
	condition := bson.M{}
	ret, err := mongo.Distinct(condition, "name")
	if err != nil {
		t.Errorf("Distinct error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoClusterSum(t *testing.T) {
	mongo := NewMongoClusterTest()
	condition := bson.M{"name": "name3"}
	ret, err := mongo.Sum(condition, "_id")
	if err != nil {
		t.Errorf("Test_MongoSum error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoClusterDistinctNameCount(t *testing.T) {
	mongo := NewMongoClusterTest()
	condition := bson.M{}
	ret, err := mongo.DistinctCount(condition, "name")
	if err != nil {
		t.Errorf("Test_MongoDistinctCount error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoClusterDistinctIdCount(t *testing.T) {
	mongo := NewMongoClusterTest()
	condition := bson.M{}
	ret, err := mongo.DistinctCount(condition, "_id")
	if err != nil {
		t.Errorf("Test_MongoDistinctCount error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoClusterDistinctObjIDCount(t *testing.T) {
	mongo := NewMongoClusterTestObj()
	condition := bson.M{}
	ret, err := mongo.DistinctCount(condition, "_id")
	if err != nil {
		t.Errorf("Test_MongoDistinctCount error:%s", err.Error())
	} else {
		t.Log("ret", ret)
	}
}

func Test_MongoClusterUpdate(t *testing.T) {
	mongo := NewMongoClusterTest()
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

func Test_MongoClusterUpdateOne(t *testing.T) {
	mongo := NewMongoClusterTest()
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

func Test_MongoClusterUpsert(t *testing.T) {
	mongo := NewMongoClusterTest()
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

func Test_MongoClusterUpsertNum(t *testing.T) {
	mongo := NewMongoClusterTest()
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

func Test_MongoClusterRemoveId(t *testing.T) {
	mongo := NewMongoClusterTest()
	err := mongo.RemoveId(2)
	if err != nil {
		t.Errorf("Test_MongoUpsertNum error:%s", err.Error())
	}
}

func Test_MongoClusterRemoveAll(t *testing.T) {
	mongo := NewMongoClusterTest()
	condition := bson.M{
		"_id": int64(22),
	}
	err := mongo.RemoveAll(condition)
	if err != nil {
		t.Errorf("Test_MongoUpsertNum error:%s", err.Error())
	}
}

func Test_MongoClusterInsertM(t *testing.T) {
	mongo := NewMongoClusterTest()
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

func Test_MongoClusterInsertMReturnInt(t *testing.T) {
	mongo := NewMongoClusterTest()
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

func Test_MongoClusterInsertMReturnObj(t *testing.T) {
	mongo := NewMongoClusterTestObj()
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

func Test_MongoClusterInsertReturnInt(t *testing.T) {
	mongo := NewMongoClusterTest()
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

func Test_MongoClusterInsertReturnObj(t *testing.T) {
	mongo := NewMongoClusterTestObj()
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

func Test_MongoClusterFindList(t *testing.T) {
	mongo := NewMongoClusterTest()
	condition := bson.M{
		"_id": bson.M{"$lt": 43},
	}
	limit := 2
	skip := 0
	var data []ModelMongoHello
	err := mongo.Find(condition, limit, skip, &data, bson.D{bson.E{Key: "_id", Value: 1}})
	if err != nil {
		t.Errorf("mongo find error,%s", err.Error())
	} else {
		t.Logf("data len is %v is %v", len(data), data)
	}
}

func Test_MongoClusterFindListNull(t *testing.T) {
	mongo := NewMongoClusterTest()
	mongo.CollectionName = "thisisnullaaa"
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

func Test_MongoClusterFindOne(t *testing.T) {
	mongo := NewMongoClusterTest()
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
