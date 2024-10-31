package tpl

var (
	ModelApiMDirName   = "model/apim"
	ModelApiMFilesName = []string{
		"userc.go",
	}
	ModelApiMFilesContent = []string{
		ModelApiMUserCGo,
	}

	ModelMApiDirName   = "model/mapi"
	ModelMApiFilesName = []string{
		"tests.go",
	}
	ModelMApiFilesContent = []string{
		ModelMApiTestsGo,
	}

	ModelMBaseDirName   = "model/mbase"
	ModelMBaseFilesName = []string{
		"img.go",
		"paginate.go",
		"token.go",
		"video.go",
	}
	ModelMBaseFilesContent = []string{
		ModelMBaseImgGo,
		ModelMBasePaginateGo,
		ModelMBaseTokenGo,
		ModelMBaseVideoGo,
	}

	ModelMMongoDirName   = "model/mmongo"
	ModelMMongoFilesName = []string{
		"user.go",
	}
	ModelMMongoFilesContent = []string{
		ModelMMongoUserGo,
	}

	ModelMMysqlDirName   = "model/mmysql"
	ModelMMysqlFilesName = []string{
		"user.go",
	}
	ModelMMysqlFilesContent = []string{
		ModelMMysqlUserGo,
	}

	ModelMParamDirName   = "model/mparam"
	ModelMParamFilesName = []string{
		"base.go",
		"tests.go",
	}
	ModelMParamFilesContent = []string{
		ModelMParamBaseGo,
		ModelMParamTestsGo,
	}

	ModelMRedisDirName   = "model/mredis"
	ModelMRedisFilesName = []string{
		"user.go",
	}
	ModelMRedisFilesContent = []string{
		ModelMRedisUserGo,
	}
)

var (
	ModelApiMUserCGo = `package apim

type UsercResp struct {
	Code    int
	Message string
	Data    map[int64]*Userc
}

type Userc struct {
	Uid              int64                 ` + "`json:\"uid\"`" + `
	AccountType      uint8                 ` + "`json:\"accountType\"`" + `
	Mobile           string                ` + "`json:\"mobile\"`" + `
	Email            string                ` + "`json:\"email\"`" + `
	LoginAccount     string                ` + "`json:\"loginAccount\"`" + `
	WechatOpenid     map[string]string     ` + "`json:\"wechatOpenid\"`" + `
	WechatUnionid    string                ` + "`json:\"wechatUnionid\"`" + `
	UserType         uint8                 ` + "`json:\"userType\"`" + `
	DiffSrcRegTime   string                ` + "`json:\"diffSrcRegTime\"`" + `
	Rating           uint8                 ` + "`json:\"rating\"`" + `
	RelationWithBaby uint8                 ` + "`json:\"relationWithBaby\"`" + `
	Nickname         string                ` + "`json:\"nickname\"`" + `
	Truename         string                ` + "`json:\"truename\"`" + `
	Sex              uint8                 ` + "`json:\"sex\"`" + `
	Birthday         uint64                ` + "`json:\"birthday\"`" + `
	QQNumber         uint64                ` + "`json:\"qQNumber\"`" + `
	Phone            string                ` + "`json:\"phone\"`" + `
	Fax              string                ` + "`json:\"fax\"`" + `
	Region           string                ` + "`json:\"region\"`" + `
	Address          string                ` + "`json:\"address\"`" + `
	ShortAddress     string                ` + "`json:\"short_address\"`" + `
	Community        string                ` + "`json:\"community\"`" + `
	Postcode         string                ` + "`json:\"postcode\"`" + `
	IdentityType     uint8                 ` + "`json:\"identityType\"`" + `
	IdentityNum      string                ` + "`json:\"identityNum\"`" + `
	Recruiter        string                ` + "`json:\"recruiter\"`" + `
	Euid             string                ` + "`json:\"euid\"`" + `
	Photo            string                ` + "`json:\"photo\"`" + `
	Signature        string                ` + "`json:\"signature\"`" + `
	UserTypeName     string                ` + "`json:\"userTypeName\"`" + `
	FullPhoto        string                ` + "`json:\"full_photo\"`" + `
	BabyDesc         string                ` + "`json:\"baby_desc\"`" + `
	BabyBirthday     uint64                ` + "`json:\"baby_birthday\"`" + `
	IsFake           uint8                 ` + "`json:\"is_fake\"`" + `
}
`
	ModelMApiTestsGo = `package mapi

import "time"

type TestRes struct {
	Data interface{} ` + "`json:\"data\"`" + ` //入参数据
	Time time.Time   ` + "`json:\"time\"`" + ` //服务器当前时间
}
`
	ModelMBaseImgGo = `package mbase

import "encoding/json"

//图片公共结构
type ImgInfo struct {
	Url    string ` + "`json:\"url\"`" + `
	Width  int    ` + "`json:\"width\"`" + `
	Height int    ` + "`json:\"height\"`" + `
}

func (p *ImgInfo) Encode() string {
	strByte, _ := json.Marshal(*p)
	return string(strByte)
}

func (p *ImgInfo) IsEmpty() bool {
	if p.Url == "" || p.Width == 0 || p.Height == 0 {
		return false
	}

	return true
}
`
	ModelMBasePaginateGo = `package mbase

type Paginate struct {
	Page     uint32 ` + "`json:\"page\" form:\"page\"`" + `           //分页号
	LimitNum uint32 ` + "`json:\"limit_num\" form:\"limit_num\"`" + ` //每页限制数量
	Offset   uint32 //计算获取offset
}

func (p *Paginate) GetOffset() uint32 {
	offset := (p.Page - 1) * p.LimitNum
	if offset < 0 {
		offset = 0
	}
	return offset
}
`
	ModelMBaseTokenGo = `package mbase

import (
	"errors"

	"github.com/speps/go-hashids"
)

func (p *TokenParam) getSalt() string {
	return "1234hahaha"
}

func (p *TokenParam) getTokenLen() int {
	return 10
}

func NewTokenParam(uid uint64, objectType uint8, objectId uint32, objectRecordId uint32) *TokenParam {
	t := new(TokenParam)
	t.Uid = uid
	t.ObjectType = objectType
	t.ObjectId = objectId
	t.ObjectRecordId = objectRecordId
	return t
}

type TokenParam struct {
	Uid            uint64 ` + "`json:\"uid\"`" + `
	ObjectType     uint8  ` + "`json:\"object_type\"`" + `
	ObjectId       uint32 ` + "`json:\"object_id\"`" + `
	ObjectRecordId uint32 ` + "`json:\"object_record_id\"`" + `
}

func (p *TokenParam) Encode() (str string, err error) {
	hd := hashids.NewData()
	hd.Salt = p.getSalt()
	hd.MinLength = p.getTokenLen()
	var h *hashids.HashID
	h, err = hashids.NewWithData(hd)
	if err != nil {
		return
	}
	str, err = h.EncodeInt64([]int64{int64(p.Uid), int64(p.ObjectType), int64(p.ObjectId), int64(p.ObjectRecordId)})
	return
}

func (p *TokenParam) Decode(str string) (err error) {
	hd := hashids.NewData()
	hd.Salt = p.getSalt()
	hd.MinLength = p.getTokenLen()
	var h *hashids.HashID
	h, err = hashids.NewWithData(hd)
	if err != nil {
		return
	}
	var params []int64
	params, err = h.DecodeInt64WithError(str)
	if err != nil {
		return
	}
	if len(params) == 4 {
		p.Uid = uint64(params[0])
		p.ObjectType = uint8(params[1])
		p.ObjectId = uint32(params[2])
		p.ObjectRecordId = uint32(params[3])
	} else {
		err = errors.New("decode len is wrong")
		return
	}
	return
}
`
	ModelMBaseVideoGo = `package mbase

import "encoding/json"

//图片公共结构
type VideoInfo struct {
	Url      string ` + "`json:\"url\"`" + `
	Duration uint32 ` + "`json:\"url\"`" + ` // 时长
	Width    uint32 ` + "`json:\"width\"`" + `
	Height   uint32 ` + "`json:\"height\"`" + `
}

func (p *VideoInfo) Encode() string {
	strByte, _ := json.Marshal(*p)
	return string(strByte)
}

func (p *VideoInfo) IsEmpty() bool {
	if p.Url == "" || p.Width == 0 ||
		p.Height == 0 || p.Duration == 0 {
		return false
	}

	return true
}
`
	ModelMMongoUserGo = `package mmongo

import "github.com/namejlt/gozen"

type User struct {
	gozen.ModelMongoBaseEx ` + "`bson:\",inline\"`" + `
	Uid                  uint64 ` + "`json:\"uid\"`" + `
	Name                 string ` + "`json:\"name\"`" + `
	Age                  uint32 ` + "`json:\"age\"`" + `
	Sex                  uint32 ` + "`json:\"sex\"`" + `
}
`
	ModelMMysqlUserGo = `package mmysql

import "time"

type User struct {
	Id        uint64    ` + "`json:\"id\" gorm:\"type:bigint(20) unsigned AUTO_INCREMENT;NOT NULL; COMMENT:'id'; primary_key\"`" + `                            //id
	TenantId  uint64    ` + "`json:\"tenant_id\" gorm:\"type:bigint(20) unsigned; DEFAULT:0; NOT NULL; COMMENT:'租户ID'\"`" + `                                   //租户ID
	Name      string    ` + "`json:\"name\" gorm:\"type:varchar(20) binary;NOT NULL; COMMENT:'name';\"`" + `                                                    //名称
	UpdatedAt time.Time ` + "`json:\"updated_at\" gorm:\"type:datetime; NOT NULL; DEFAULT: CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;COMMENT:'更新时间'\"`" + ` //更新时间
	CreatedAt time.Time ` + "`json:\"created_at\" gorm:\"type:datetime; NOT NULL; DEFAULT: CURRENT_TIMESTAMP;COMMENT:'创建时间'\"`" + `                             //创建时间
}
`
	ModelMParamBaseGo = `package mparam

import "{{.Name}}/pconst"

type PageParam struct {
	Page     int ` + "`json:\"page\" form:\"page\"`" + `
	LimitNum int ` + "`json:\"limit_num\" form:\"limit_num\"`" + `
}

func (p *PageParam) FixPage(defaultNum int) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.LimitNum <= 0 {
		p.LimitNum = defaultNum
	} else if p.LimitNum > pconst.COMMON_PAGE_LIMIT_NUM_MAX {
		p.LimitNum = pconst.COMMON_PAGE_LIMIT_NUM_MAX
	}
}
`
	ModelMParamTestsGo = `package mparam

// TestParam 获取测试请求数据 入参
type TestParam struct {
	Name    string ` + "`json:\"name\" form:\"name\" binding:\"required\"`" + ` //名称
	Age     int    ` + "`json:\"age\" form:\"age\"`" + `                      //年龄
	Content string ` + "`json:\"content\" form:\"content\"`" + `              //内容
}
`
	ModelMRedisUserGo = `package mredis

import "{{.Name}}/model/mmysql"

type User struct {
	mmysql.User
}
`
)
