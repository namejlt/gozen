package gozen

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type BaseMongo struct {
	CreatedAt time.Time `bson:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at,omitempty"`
}

type BaseMongoEx struct {
	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at"`
}

// 自增id使用
type ModelMongo struct {
	Id        int64 `bson:"_id,omitempty"`
	BaseMongo `bson:",inline"`
}

func (m *ModelMongo) GetCreatedTime() time.Time {
	return m.CreatedAt
}

func (m *ModelMongo) InitTime(t time.Time) {
	m.CreatedAt = t
	m.UpdatedAt = t
}
func (m *ModelMongo) SetUpdatedTime(t time.Time) {
	m.UpdatedAt = t
}

func (m *ModelMongo) SetId(id int64) {
	m.Id = id
}

func (m *ModelMongo) GetId() int64 {
	return m.Id
}

func (m *ModelMongo) SetObjectId() {
}

func (m *ModelMongo) ExistId() (b bool) {
	if m.Id != 0 {
		b = true
	}
	return
}

// 自增id使用
type ModelMongoEx struct {
	Id          int64 `bson:"_id,omitempty" json:"id"`
	BaseMongoEx `bson:",inline"`
}

func (m *ModelMongoEx) GetCreatedTime() time.Time {
	return m.CreatedAt
}

func (m *ModelMongoEx) InitTime(t time.Time) {
	m.CreatedAt = t
	m.UpdatedAt = t
}
func (m *ModelMongoEx) SetUpdatedTime(t time.Time) {
	m.UpdatedAt = t
}

func (m *ModelMongoEx) SetId(id int64) {
	m.Id = id
}

func (m *ModelMongoEx) GetId() int64 {
	return m.Id
}

func (m *ModelMongoEx) SetObjectId() {
}

func (m *ModelMongoEx) ExistId() (b bool) {
	if m.Id != 0 {
		b = true
	}
	return
}

// db生成id使用
type ModelMongoBase struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BaseMongo `bson:",inline"`
}

func (m *ModelMongoBase) GetCreatedTime() time.Time {
	return m.CreatedAt
}

func (m *ModelMongoBase) InitTime(t time.Time) {
	m.CreatedAt = t
	m.UpdatedAt = t
}
func (m *ModelMongoBase) SetUpdatedTime(t time.Time) {
	m.UpdatedAt = t
}

func (m *ModelMongoBase) SetId(id int64) {
}

func (m *ModelMongoBase) GetId() (id int64) {
	return
}

func (m *ModelMongoBase) SetObjectId() {
	m.Id = primitive.NewObjectID()
}

func (m *ModelMongoBase) ExistId() (b bool) {
	return !m.Id.IsZero()
}

// db生成id使用
type ModelMongoBaseEx struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BaseMongoEx `bson:",inline"`
}

func (m *ModelMongoBaseEx) GetCreatedTime() time.Time {
	return m.CreatedAt
}

func (m *ModelMongoBaseEx) InitTime(t time.Time) {
	m.CreatedAt = t
	m.UpdatedAt = t
}
func (m *ModelMongoBaseEx) SetUpdatedTime(t time.Time) {
	m.UpdatedAt = t
}

func (m *ModelMongoBaseEx) SetId(id int64) {
}

func (m *ModelMongoBaseEx) GetId() (id int64) {
	return
}

func (m *ModelMongoBaseEx) SetObjectId() {
	m.Id = primitive.NewObjectID()
}

func (m *ModelMongoBaseEx) ExistId() (b bool) {
	return !m.Id.IsZero()
}
